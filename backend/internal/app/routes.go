package app

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/via-justa/overpacked-app/backend/internal/api"
	"github.com/via-justa/overpacked-app/backend/internal/http/handlers"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

const (
	personIDParam        = "personId"
	jsonContentType      = "application/json"
	invalidPersonIDError = "invalid person id"
)

type apiServer struct {
	auth          *handlers.AuthHandler
	persons       *handlers.PersonsHandler
	settings      *handlers.SettingsHandler
	manufacturers *handlers.ManufacturersHandler
	items         *handlers.ItemsHandler
	itemTypes     *handlers.ItemTypesHandler
	labels        *handlers.LabelsHandler
	sets          *handlers.SetsHandler
	packs         *handlers.PacksHandler
	packingLists  *handlers.PackingListsHandler
	trips         *handlers.TripsHandler
}

func (s *apiServer) AuthLogin(w http.ResponseWriter, r *http.Request)   { s.auth.Login(w, r) }
func (s *apiServer) AuthLogout(w http.ResponseWriter, r *http.Request)  { s.auth.Logout(w, r) }
func (s *apiServer) AuthRefresh(w http.ResponseWriter, r *http.Request) { s.auth.Refresh(w, r) }

func (s *apiServer) ListPersons(w http.ResponseWriter, r *http.Request) { s.persons.ListPersons(w, r) }
func (s *apiServer) CreatePerson(w http.ResponseWriter, r *http.Request) {
	s.persons.CreatePerson(w, r)
}
func (s *apiServer) DeletePerson(w http.ResponseWriter, r *http.Request, personId openapi_types.UUID) {
	s.persons.DeletePerson(w, r, personId)
}
func (s *apiServer) GetPerson(w http.ResponseWriter, r *http.Request, personId openapi_types.UUID) {
	s.persons.GetPerson(w, r, personId)
}
func (s *apiServer) UpdatePerson(w http.ResponseWriter, r *http.Request, personId openapi_types.UUID) {
	s.persons.UpdatePerson(w, r, personId)
}

func (s *apiServer) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *apiServer) ListItems(w http.ResponseWriter, r *http.Request) { s.items.ListItems(w, r) }
func (s *apiServer) CreateItem(w http.ResponseWriter, r *http.Request) {
	s.items.CreateItem(w, r)
}
func (s *apiServer) DeleteItem(w http.ResponseWriter, r *http.Request, itemId openapi_types.UUID) {
	s.items.DeleteItem(w, r, itemId)
}
func (s *apiServer) GetItem(w http.ResponseWriter, r *http.Request, itemId openapi_types.UUID) {
	s.items.GetItem(w, r, itemId)
}
func (s *apiServer) UpdateItem(w http.ResponseWriter, r *http.Request, itemId openapi_types.UUID) {
	s.items.UpdateItem(w, r, itemId)
}

func (s *apiServer) ListItemTypes(w http.ResponseWriter, r *http.Request) {
	s.itemTypes.ListItemTypes(w, r)
}

func (s *apiServer) CreateItemType(w http.ResponseWriter, r *http.Request) {
	s.itemTypes.CreateItemType(w, r)
}

func (s *apiServer) GetItemType(w http.ResponseWriter, r *http.Request, typeId string) {
	s.itemTypes.GetItemType(w, r, typeId)
}

func (s *apiServer) UpdateItemType(w http.ResponseWriter, r *http.Request, typeId string) {
	s.itemTypes.UpdateItemType(w, r, typeId)
}

func (s *apiServer) DeleteItemType(w http.ResponseWriter, r *http.Request, typeId string) {
	s.itemTypes.DeleteItemType(w, r, typeId)
}

func (s *apiServer) ListItemTypeFields(w http.ResponseWriter, r *http.Request, typeId string) {
	s.itemTypes.ListItemTypeFields(w, r, typeId)
}

func (s *apiServer) ReplaceItemTypeFields(w http.ResponseWriter, r *http.Request, typeId string) {
	s.itemTypes.ReplaceItemTypeFields(w, r, typeId)
}

func (s *apiServer) ListLabels(w http.ResponseWriter, r *http.Request) { s.labels.ListLabels(w, r) }
func (s *apiServer) CreateLabel(w http.ResponseWriter, r *http.Request) {
	s.labels.CreateLabel(w, r)
}
func (s *apiServer) GetLabel(w http.ResponseWriter, r *http.Request, labelId openapi_types.UUID) {
	s.labels.GetLabel(w, r, labelId)
}
func (s *apiServer) UpdateLabel(w http.ResponseWriter, r *http.Request, labelId openapi_types.UUID) {
	s.labels.UpdateLabel(w, r, labelId)
}
func (s *apiServer) DeleteLabel(w http.ResponseWriter, r *http.Request, labelId openapi_types.UUID) {
	s.labels.DeleteLabel(w, r, labelId)
}

func (s *apiServer) ListItemLabels(w http.ResponseWriter, r *http.Request, itemId openapi_types.UUID) {
	s.labels.ListItemLabels(w, r, itemId)
}
func (s *apiServer) AddItemLabel(w http.ResponseWriter, r *http.Request, itemId openapi_types.UUID) {
	s.labels.AddItemLabel(w, r, itemId)
}
func (s *apiServer) RemoveItemLabel(w http.ResponseWriter, r *http.Request, itemId, labelId openapi_types.UUID) {
	s.labels.RemoveItemLabel(w, r, itemId, labelId)
}

func (s *apiServer) ListManufacturers(w http.ResponseWriter, r *http.Request) {
	s.manufacturers.ListManufacturers(w, r)
}
func (s *apiServer) CreateManufacturer(w http.ResponseWriter, r *http.Request) {
	s.manufacturers.CreateManufacturer(w, r)
}
func (s *apiServer) DeleteManufacturer(w http.ResponseWriter, r *http.Request, manufacturerId openapi_types.UUID) {
	s.manufacturers.DeleteManufacturer(w, r, manufacturerId)
}
func (s *apiServer) GetManufacturer(w http.ResponseWriter, r *http.Request, manufacturerId openapi_types.UUID) {
	s.manufacturers.GetManufacturer(w, r, manufacturerId)
}
func (s *apiServer) UpdateManufacturer(w http.ResponseWriter, r *http.Request, manufacturerId openapi_types.UUID) {
	s.manufacturers.UpdateManufacturer(w, r, manufacturerId)
}

func (s *apiServer) ListPacks(w http.ResponseWriter, r *http.Request) { s.packs.ListPacks(w, r) }
func (s *apiServer) CreatePack(w http.ResponseWriter, r *http.Request) {
	s.packs.CreatePack(w, r)
}
func (s *apiServer) DeletePack(w http.ResponseWriter, r *http.Request, packId openapi_types.UUID) {
	s.packs.DeletePack(w, r, packId)
}
func (s *apiServer) GetPack(w http.ResponseWriter, r *http.Request, packId openapi_types.UUID) {
	s.packs.GetPack(w, r, packId)
}
func (s *apiServer) UpdatePack(w http.ResponseWriter, r *http.Request, packId openapi_types.UUID) {
	s.packs.UpdatePack(w, r, packId)
}
func (s *apiServer) ListPackItems(w http.ResponseWriter, r *http.Request, packId openapi_types.UUID) {
	s.packs.ListPackItems(w, r, packId)
}
func (s *apiServer) AddPackItem(w http.ResponseWriter, r *http.Request, packId openapi_types.UUID) {
	s.packs.AddPackItem(w, r, packId)
}
func (s *apiServer) RemovePackItem(w http.ResponseWriter, r *http.Request, packId, itemId openapi_types.UUID) {
	s.packs.RemovePackItem(w, r, packId, itemId)
}
func (s *apiServer) UpdatePackItem(w http.ResponseWriter, r *http.Request, packId, itemId openapi_types.UUID) {
	s.packs.UpdatePackItem(w, r, packId, itemId)
}

func (s *apiServer) ListSets(w http.ResponseWriter, r *http.Request) { s.sets.ListSets(w, r) }
func (s *apiServer) CreateSet(w http.ResponseWriter, r *http.Request) {
	s.sets.CreateSet(w, r)
}
func (s *apiServer) DeleteSet(w http.ResponseWriter, r *http.Request, setId openapi_types.UUID) {
	s.sets.DeleteSet(w, r, setId)
}
func (s *apiServer) GetSet(w http.ResponseWriter, r *http.Request, setId openapi_types.UUID) {
	s.sets.GetSet(w, r, setId)
}
func (s *apiServer) UpdateSet(w http.ResponseWriter, r *http.Request, setId openapi_types.UUID) {
	s.sets.UpdateSet(w, r, setId)
}
func (s *apiServer) ListSetItems(w http.ResponseWriter, r *http.Request, setId openapi_types.UUID) {
	s.sets.ListSetItems(w, r, setId)
}
func (s *apiServer) AddSetItem(w http.ResponseWriter, r *http.Request, setId openapi_types.UUID) {
	s.sets.AddSetItem(w, r, setId)
}
func (s *apiServer) RemoveSetItem(w http.ResponseWriter, r *http.Request, setId, itemId openapi_types.UUID) {
	s.sets.RemoveSetItem(w, r, setId, itemId)
}
func (s *apiServer) UpdateSetItem(w http.ResponseWriter, r *http.Request, setId, itemId openapi_types.UUID) {
	s.sets.UpdateSetItem(w, r, setId, itemId)
}

// Trip handlers
func (s *apiServer) ListTrips(w http.ResponseWriter, r *http.Request) { s.trips.ListTrips(w, r) }
func (s *apiServer) CreateTrip(w http.ResponseWriter, r *http.Request) {
	s.trips.CreateTrip(w, r)
}
func (s *apiServer) GetTripById(w http.ResponseWriter, r *http.Request, tripId openapi_types.UUID) {
	s.trips.GetTripById(w, r, tripId)
}
func (s *apiServer) UpdateTrip(w http.ResponseWriter, r *http.Request, tripId openapi_types.UUID) {
	s.trips.UpdateTrip(w, r, tripId)
}
func (s *apiServer) DeleteTrip(w http.ResponseWriter, r *http.Request, tripId openapi_types.UUID) {
	s.trips.DeleteTrip(w, r, tripId)
}

// Trip packs handlers
func (s *apiServer) ListTripPacks(w http.ResponseWriter, r *http.Request, tripId openapi_types.UUID) {
	s.trips.ListTripPacks(w, r, tripId)
}
func (s *apiServer) AddTripPack(w http.ResponseWriter, r *http.Request, tripId openapi_types.UUID) {
	s.trips.AddTripPack(w, r, tripId)
}
func (s *apiServer) RemoveTripPack(w http.ResponseWriter, r *http.Request, tripId, packId openapi_types.UUID) {
	s.trips.RemoveTripPack(w, r, tripId, packId)
}

// Trip items handlers
func (s *apiServer) ListTripItems(w http.ResponseWriter, r *http.Request, tripId openapi_types.UUID) {
	s.trips.ListTripItems(w, r, tripId)
}
func (s *apiServer) AddTripItem(w http.ResponseWriter, r *http.Request, tripId openapi_types.UUID) {
	s.trips.AddTripItem(w, r, tripId)
}
func (s *apiServer) UpdateTripItem(w http.ResponseWriter, r *http.Request, tripId, itemId openapi_types.UUID) {
	s.trips.UpdateTripItem(w, r, tripId, itemId)
}
func (s *apiServer) RemoveTripItem(w http.ResponseWriter, r *http.Request, tripId, itemId openapi_types.UUID) {
	s.trips.RemoveTripItem(w, r, tripId, itemId)
}

// Trip sets handlers
func (s *apiServer) ListTripSets(w http.ResponseWriter, r *http.Request, tripId openapi_types.UUID) {
	s.trips.ListTripSets(w, r, tripId)
}
func (s *apiServer) AddTripSet(w http.ResponseWriter, r *http.Request, tripId openapi_types.UUID) {
	s.trips.AddTripSet(w, r, tripId)
}
func (s *apiServer) RemoveTripSet(w http.ResponseWriter, r *http.Request, tripId, setId openapi_types.UUID) {
	s.trips.RemoveTripSet(w, r, tripId, setId)
}

// Trip persons handlers
func (s *apiServer) ListTripPersons(w http.ResponseWriter, r *http.Request, tripId openapi_types.UUID) {
	s.trips.ListTripPersons(w, r, tripId)
}
func (s *apiServer) AddTripPerson(w http.ResponseWriter, r *http.Request, tripId openapi_types.UUID) {
	s.trips.AddTripPerson(w, r, tripId)
}
func (s *apiServer) RemoveTripPerson(w http.ResponseWriter, r *http.Request, tripId, personId openapi_types.UUID) {
	s.trips.RemoveTripPerson(w, r, tripId, personId)
}

// PackingLists handlers
func (s *apiServer) ListPackingLists(w http.ResponseWriter, r *http.Request) {
	s.packingLists.ListPackingLists(w, r)
}
func (s *apiServer) CreatePackingList(w http.ResponseWriter, r *http.Request) {
	s.packingLists.CreatePackingList(w, r)
}
func (s *apiServer) GetPackingListById(w http.ResponseWriter, r *http.Request, listId openapi_types.UUID) {
	s.packingLists.GetPackingListById(w, r, listId)
}
func (s *apiServer) UpdatePackingList(w http.ResponseWriter, r *http.Request, listId openapi_types.UUID) {
	s.packingLists.UpdatePackingList(w, r, listId)
}
func (s *apiServer) DeletePackingList(w http.ResponseWriter, r *http.Request, listId openapi_types.UUID) {
	s.packingLists.DeletePackingList(w, r, listId)
}
func (s *apiServer) ListPackingListLabels(w http.ResponseWriter, r *http.Request, listId openapi_types.UUID) {
	s.packingLists.ListPackingListLabels(w, r, listId)
}
func (s *apiServer) AddPackingListLabel(w http.ResponseWriter, r *http.Request, listId openapi_types.UUID) {
	s.packingLists.AddPackingListLabel(w, r, listId)
}
func (s *apiServer) RemovePackingListLabel(w http.ResponseWriter, r *http.Request, listId, labelId openapi_types.UUID) {
	s.packingLists.RemovePackingListLabel(w, r, listId, labelId)
}

func (s *apiServer) GetSettings(w http.ResponseWriter, r *http.Request) { s.settings.GetSettings(w, r) }
func (s *apiServer) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	s.settings.UpdateSettings(w, r)
}
func (s *apiServer) StartFresh(w http.ResponseWriter, r *http.Request) {
	s.settings.StartFresh(w, r)
}

func NewRouter(authHandler *handlers.AuthHandler, st *store.Store, appPassword string) chi.Router {
	r := chi.NewRouter()

	h := api.HandlerWithOptions(&apiServer{
		auth:          authHandler,
		persons:       handlers.NewPersonsHandler(st),
		settings:      handlers.NewSettingsHandler(st, appPassword),
		manufacturers: handlers.NewManufacturersHandler(st),
		items:         handlers.NewItemsHandler(st),
		itemTypes:     handlers.NewItemTypesHandler(st),
		labels:        handlers.NewLabelsHandler(st),
		sets:          handlers.NewSetsHandler(st),
		packs:         handlers.NewPacksHandler(st),
		packingLists:  handlers.NewPackingListsHandler(st.PackingLists, st.Labels),
		trips:         handlers.NewTripsHandler(st),
	}, api.StdHTTPServerOptions{
		ErrorHandlerFunc: handleOpenAPIError,
	})

	r.Mount("/", h)

	return r
}

func setupRoutes(authHandler *handlers.AuthHandler, st *store.Store, appPassword string) chi.Router {
	return NewRouter(authHandler, st, appPassword)
}

func handleOpenAPIError(w http.ResponseWriter, _ *http.Request, err error) {
	var invalidParamErr *api.InvalidParamFormatError
	if errors.As(err, &invalidParamErr) && invalidParamErr.ParamName == personIDParam {
		writeJSONError(w, http.StatusBadRequest, invalidPersonIDError)
		return
	}

	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + message + `"}`))
}
