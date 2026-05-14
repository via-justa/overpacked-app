package store

import "database/sql"

type Store struct {
	Settings      *SettingsStore
	Persons       *PersonStore
	Manufacturers *ManufacturerStore
	Items         *ItemStore
	ItemTypes     *ItemTypeStore
	Sets          *SetStore
	Packs         *PackStore
}

func New(db *sql.DB) *Store {
	return &Store{
		Settings:      NewSettingsStore(db),
		Persons:       NewPersonStore(db),
		Manufacturers: NewManufacturerStore(db),
		Items:         NewItemStore(db),
		ItemTypes:     NewItemTypeStore(db),
		Sets:          NewSetStore(db),
		Packs:         NewPackStore(db),
	}
}
