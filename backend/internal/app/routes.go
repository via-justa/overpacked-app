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
	sets          *handlers.SetsHandler
	packs         *handlers.PacksHandler
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
func (s *apiServer) ListPackSets(w http.ResponseWriter, r *http.Request, packId openapi_types.UUID) {
	s.packs.ListPackSets(w, r, packId)
}
func (s *apiServer) AddPackSet(w http.ResponseWriter, r *http.Request, packId openapi_types.UUID) {
	s.packs.AddPackSet(w, r, packId)
}
func (s *apiServer) RemovePackSet(w http.ResponseWriter, r *http.Request, packId, setId openapi_types.UUID) {
	s.packs.RemovePackSet(w, r, packId, setId)
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
		sets:          handlers.NewSetsHandler(st),
		packs:         handlers.NewPacksHandler(st),
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
