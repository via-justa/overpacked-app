package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/via-justa/overpacked-app/backend/internal/api"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

const (
	errPersonNotFound = "person not found"
)

type PersonsHandler struct {
	store *store.Store
}

func NewPersonsHandler(st *store.Store) *PersonsHandler {
	return &PersonsHandler{store: st}
}

func (h *PersonsHandler) ListPersons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	persons, err := h.store.Persons.List(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list persons")
		return
	}

	resp := make([]api.Person, len(persons))
	for i, p := range persons {
		resp[i] = personToAPI(&p)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *PersonsHandler) CreatePerson(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req api.PersonCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	domainPerson := &domain.Person{
		ID:   uuid.New(),
		Name: req.Name,
	}

	if req.Gender != nil {
		g := domain.Gender(*req.Gender)
		domainPerson.Gender = &g
	}
	if req.Birthdate != nil {
		bd := req.Birthdate.Time
		domainPerson.Birthdate = &bd
	}
	if req.BodyWeightGrams != nil {
		bw := float64(*req.BodyWeightGrams)
		domainPerson.BodyWeightGrams = &bw
	}
	if req.ConditioningLevel != nil {
		cl := domain.ConditioningLevel(*req.ConditioningLevel)
		domainPerson.ConditioningLevel = &cl
	}

	if err := h.store.Persons.Create(ctx, domainPerson); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create person")
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, http.StatusCreated, personToAPI(domainPerson))
}

func (h *PersonsHandler) GetPerson(w http.ResponseWriter, r *http.Request, personId types.UUID) {
	ctx := r.Context()

	person, err := h.store.Persons.GetByID(ctx, uuid.UUID(personId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, errPersonNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get person")
		return
	}

	writeJSON(w, http.StatusOK, personToAPI(person))
}

func (h *PersonsHandler) UpdatePerson(w http.ResponseWriter, r *http.Request, personId types.UUID) {
	ctx := r.Context()

	var req api.PersonUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	person, err := h.store.Persons.GetByID(ctx, uuid.UUID(personId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, errPersonNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get person")
		return
	}

	// Apply updates
	if req.Name != nil {
		person.Name = *req.Name
	}
	if req.Gender != nil {
		g := domain.Gender(*req.Gender)
		person.Gender = &g
	}
	if req.Birthdate != nil {
		bd := req.Birthdate.Time
		person.Birthdate = &bd
	}
	if req.BodyWeightGrams != nil {
		bw := float64(*req.BodyWeightGrams)
		person.BodyWeightGrams = &bw
	}
	if req.ConditioningLevel != nil {
		cl := domain.ConditioningLevel(*req.ConditioningLevel)
		person.ConditioningLevel = &cl
	}

	if err := h.store.Persons.Update(ctx, person); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update person")
		return
	}

	writeJSON(w, http.StatusOK, personToAPI(person))
}

func (h *PersonsHandler) DeletePerson(w http.ResponseWriter, r *http.Request, personId types.UUID) {
	ctx := r.Context()

	err := h.store.Persons.Delete(ctx, uuid.UUID(personId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, errPersonNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete person")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper to convert domain.Person to api.Person
func personToAPI(p *domain.Person) api.Person {
	var gender *api.PersonGender
	if p.Gender != nil {
		g := api.PersonGender(*p.Gender)
		gender = &g
	}

	var birthdate *types.Date
	if p.Birthdate != nil {
		bd := types.Date{Time: *p.Birthdate}
		birthdate = &bd
	}

	var bodyWeight *float32
	if p.BodyWeightGrams != nil {
		bw := float32(*p.BodyWeightGrams)
		bodyWeight = &bw
	}

	var conditioningLevel *api.PersonConditioningLevel
	if p.ConditioningLevel != nil {
		cl := api.PersonConditioningLevel(*p.ConditioningLevel)
		conditioningLevel = &cl
	}

	return api.Person{
		Id:                types.UUID(p.ID),
		Name:              p.Name,
		Gender:            gender,
		Birthdate:         birthdate,
		BodyWeightGrams:   bodyWeight,
		ConditioningLevel: conditioningLevel,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}
