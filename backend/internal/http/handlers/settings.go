package handlers

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/via-justa/overpacked-app/backend/internal/api"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

type SettingsHandler struct {
	store       *store.Store
	appPassword string
}

func NewSettingsHandler(st *store.Store, appPassword string) *SettingsHandler {
	return &SettingsHandler{store: st, appPassword: appPassword}
}

func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.store.Settings.Get(r.Context())
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "settings not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get settings")
		return
	}

	writeJSON(w, http.StatusOK, settingsToAPI(settings))
}

func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req api.SettingsUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	settings, err := h.store.Settings.Get(r.Context())
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "settings not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get settings")
		return
	}

	if req.WeightUnit != nil {
		settings.WeightUnit = domain.WeightUnit(*req.WeightUnit)
	}
	if req.DistanceUnit != nil {
		settings.DistanceUnit = domain.DistanceUnit(*req.DistanceUnit)
	}
	if req.TemperatureUnit != nil {
		settings.TemperatureUnit = domain.TemperatureUnit(*req.TemperatureUnit)
	}
	if req.VolumeUnit != nil {
		settings.VolumeUnit = domain.VolumeUnit(*req.VolumeUnit)
	}
	if req.Currency != nil {
		settings.Currency = domain.Currency(*req.Currency)
	}

	if err := h.store.Settings.Update(r.Context(), settings); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update settings")
		return
	}

	writeJSON(w, http.StatusOK, settingsToAPI(settings))
}

func (h *SettingsHandler) StartFresh(w http.ResponseWriter, r *http.Request) {
	var req api.StartFreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	if subtle.ConstantTimeCompare([]byte(req.Password), []byte(h.appPassword)) != 1 {
		writeError(w, http.StatusUnauthorized, "password confirmation failed")
		return
	}

	reseed := req.Reseed != nil && *req.Reseed
	if err := h.store.Settings.StartFresh(r.Context(), reseed); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reset data")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func settingsToAPI(s *domain.Settings) api.Settings {
	return api.Settings{
		Id:              s.ID,
		WeightUnit:      api.SettingsWeightUnit(s.WeightUnit),
		DistanceUnit:    api.SettingsDistanceUnit(s.DistanceUnit),
		TemperatureUnit: api.SettingsTemperatureUnit(s.TemperatureUnit),
		VolumeUnit:      api.SettingsVolumeUnit(s.VolumeUnit),
		Currency:        api.SettingsCurrency(s.Currency),
	}
}
