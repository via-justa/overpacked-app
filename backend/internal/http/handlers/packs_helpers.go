package handlers

import (
	"github.com/oapi-codegen/runtime/types"
	"github.com/via-justa/overpacked-app/backend/internal/api"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
)

// Pack helper functions for converting domain models to API models
// Packs are no longer managed as standalone resources - they are managed through trips

func packToAPI(p *domain.Pack) api.Pack {
	tripType := api.PackTripTypeDayHike
	if p.TripType != nil {
		tripType = api.PackTripType(*p.TripType)
	}

	resp := api.Pack{
		Id:        types.UUID(p.ID),
		Name:      p.Name,
		TripType:  tripType,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	if p.PersonID != nil {
		pid := types.UUID(*p.PersonID)
		resp.PersonId = &pid
	}
	return resp
}
