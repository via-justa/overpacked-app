package store

import "database/sql"

type Store struct {
	Settings      *SettingsStore
	Persons       *PersonStore
	Manufacturers *ManufacturerStore
	Items         *ItemStore
	ItemTypes     *ItemTypeStore
	Labels        *LabelStore
	Sets          *SetStore
	Packs         *PackStore
	PackingLists  *PackingListStore
	Trips         *TripStore
}

func New(db *sql.DB) *Store {
	return &Store{
		Settings:      NewSettingsStore(db),
		Persons:       NewPersonStore(db),
		Manufacturers: NewManufacturerStore(db),
		Items:         NewItemStore(db),
		ItemTypes:     NewItemTypeStore(db),
		Labels:        NewLabelStore(db),
		Sets:          NewSetStore(db),
		Packs:         NewPackStore(db),
		PackingLists:  NewPackingListStore(db),
		Trips:         NewTripStore(db),
	}
}
