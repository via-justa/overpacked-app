package handlers

import (
	"net/http"
	"strings"

	"github.com/via-justa/overpacked-app/backend/internal/api"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

const (
	searchMinQueryLength = 2
	searchDefaultLimit   = 20
	searchMaxLimit       = 50

	searchErrQueryTooShort = "search query must be at least 2 characters"
	searchErrInvalidType   = "invalid entity type filter"
	searchErrFailed        = "failed to perform search"
)

type SearchHandler struct {
	store *store.Store
}

func NewSearchHandler(st *store.Store) *SearchHandler {
	return &SearchHandler{store: st}
}

func (h *SearchHandler) SearchGlobal(w http.ResponseWriter, r *http.Request, params api.SearchGlobalParams) {
	query := strings.TrimSpace(params.Q)
	if len(query) < searchMinQueryLength {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": searchErrQueryTooShort})
		return
	}

	types, ok := normalizeSearchTypes(params.Types)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": searchErrInvalidType})
		return
	}

	results, err := h.store.Search.Search(r.Context(), query, types, resolveSearchLimit(params.Limit))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": searchErrFailed})
		return
	}

	resp := make([]api.SearchResult, len(results))
	for i := range results {
		resp[i] = searchResultToAPI(&results[i])
	}

	writeJSON(w, http.StatusOK, resp)
}

// normalizeSearchTypes validates the requested entity types and converts them
// to domain types. It returns false when any type is not a known enum value.
func normalizeSearchTypes(types *[]api.SearchEntityType) ([]domain.SearchEntityType, bool) {
	if types == nil || len(*types) == 0 {
		return nil, true
	}

	out := make([]domain.SearchEntityType, 0, len(*types))
	for _, t := range *types {
		if !t.Valid() {
			return nil, false
		}
		out = append(out, domain.SearchEntityType(t))
	}

	return out, true
}

// resolveSearchLimit applies the default and caps the requested limit.
func resolveSearchLimit(limit *int) int {
	if limit == nil {
		return searchDefaultLimit
	}
	if *limit < 1 {
		return searchDefaultLimit
	}
	if *limit > searchMaxLimit {
		return searchMaxLimit
	}
	return *limit
}

func searchResultToAPI(result *domain.SearchResult) api.SearchResult {
	return api.SearchResult{
		EntityType: api.SearchEntityType(result.EntityType),
		Id:         result.ID,
		Title:      result.Title,
		Subtitle:   result.Subtitle,
		Score:      result.Score,
	}
}
