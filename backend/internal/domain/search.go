package domain

// SearchEntityType identifies which entity a search result refers to.
type SearchEntityType string

const (
	SearchEntityItem         SearchEntityType = "item"
	SearchEntitySet          SearchEntityType = "set"
	SearchEntityPackingList  SearchEntityType = "packing_list"
	SearchEntityPerson       SearchEntityType = "person"
	SearchEntityManufacturer SearchEntityType = "manufacturer"
	SearchEntityTrip         SearchEntityType = "trip"
)

// SearchResult is a single match returned by global search across entities.
type SearchResult struct {
	EntityType SearchEntityType
	ID         string
	Title      string
	Subtitle   *string
	Score      float64
}
