package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/via-justa/overpacked-app/backend/internal/domain"
)

type SearchStore struct {
	db *sql.DB
}

func NewSearchStore(db *sql.DB) *SearchStore {
	return &SearchStore{db: db}
}

// searchBranches maps each entity type to the SELECT statement that contributes
// its rows to the global search UNION. Every branch exposes the same columns:
// entity_type, id, title, subtitle, score. Placeholders: $1 = raw query (for
// similarity and the trigram operator), $2 = escaped ILIKE pattern.
var searchBranches = map[domain.SearchEntityType]string{
	domain.SearchEntityItem: `
		SELECT 'item' AS entity_type, id::text AS id, name AS title, description AS subtitle,
			GREATEST(similarity(name, $1), similarity(COALESCE(description, ''), $1)) AS score
		FROM items
		WHERE name % $1 OR description % $1 OR name ILIKE $2 OR COALESCE(description, '') ILIKE $2`,
	domain.SearchEntitySet: `
		SELECT 'set' AS entity_type, id::text AS id, name AS title, description AS subtitle,
			GREATEST(similarity(name, $1), similarity(COALESCE(description, ''), $1)) AS score
		FROM item_sets
		WHERE name % $1 OR description % $1 OR name ILIKE $2 OR COALESCE(description, '') ILIKE $2`,
	domain.SearchEntityPackingList: `
		SELECT 'packing_list' AS entity_type, id::text AS id, name AS title, description AS subtitle,
			GREATEST(similarity(name, $1), similarity(COALESCE(description, ''), $1)) AS score
		FROM packing_lists
		WHERE name % $1 OR description % $1 OR name ILIKE $2 OR COALESCE(description, '') ILIKE $2`,
	domain.SearchEntityPerson: `
		SELECT 'person' AS entity_type, id::text AS id, name AS title, NULL::text AS subtitle,
			similarity(name, $1) AS score
		FROM persons
		WHERE name % $1 OR name ILIKE $2`,
	domain.SearchEntityManufacturer: `
		SELECT 'manufacturer' AS entity_type, id::text AS id, name AS title, NULL::text AS subtitle,
			similarity(name, $1) AS score
		FROM manufacturers
		WHERE name % $1 OR name ILIKE $2`,
	domain.SearchEntityTrip: `
		SELECT 'trip' AS entity_type, id::text AS id, name AS title, notes AS subtitle,
			GREATEST(similarity(name, $1), similarity(COALESCE(notes, ''), $1)) AS score
		FROM trips
		WHERE name % $1 OR notes % $1 OR name ILIKE $2 OR COALESCE(notes, '') ILIKE $2`,
}

// searchEntityOrder fixes the branch order so generated SQL is deterministic.
var searchEntityOrder = []domain.SearchEntityType{
	domain.SearchEntityItem,
	domain.SearchEntitySet,
	domain.SearchEntityPackingList,
	domain.SearchEntityPerson,
	domain.SearchEntityManufacturer,
	domain.SearchEntityTrip,
}

// Search runs a fuzzy global search across the requested entity types. When
// types is empty, all entity types are searched. Results are ordered by
// descending trigram similarity score and capped at limit.
func (s *SearchStore) Search(ctx context.Context, query string, types []domain.SearchEntityType, limit int) ([]domain.SearchResult, error) {
	branches := s.selectBranches(types)
	if len(branches) == 0 {
		return []domain.SearchResult{}, nil
	}

	sqlQuery := fmt.Sprintf(`
		SELECT entity_type, id, title, subtitle, score
		FROM (
			%s
		) AS results
		ORDER BY score DESC, title ASC
		LIMIT $3`, strings.Join(branches, "\nUNION ALL\n"))

	rows, err := s.db.QueryContext(ctx, sqlQuery, query, ilikePattern(query), limit)
	if err != nil {
		return nil, fmt.Errorf("search entities: %w", err)
	}
	defer rows.Close()

	out := make([]domain.SearchResult, 0)
	for rows.Next() {
		var result domain.SearchResult
		var subtitle sql.NullString
		if err := rows.Scan(&result.EntityType, &result.ID, &result.Title, &subtitle, &result.Score); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}
		result.Subtitle = strPtr(subtitle)
		out = append(out, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search results: %w", err)
	}

	return out, nil
}

func (s *SearchStore) selectBranches(types []domain.SearchEntityType) []string {
	requested := make(map[domain.SearchEntityType]struct{}, len(types))
	for _, t := range types {
		requested[t] = struct{}{}
	}

	branches := make([]string, 0, len(searchEntityOrder))
	for _, entity := range searchEntityOrder {
		if len(requested) > 0 {
			if _, ok := requested[entity]; !ok {
				continue
			}
		}
		branches = append(branches, searchBranches[entity])
	}

	return branches
}

// ilikePattern escapes LIKE wildcards in the user query and wraps it for a
// case-insensitive substring match.
func ilikePattern(query string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + replacer.Replace(query) + "%"
}
