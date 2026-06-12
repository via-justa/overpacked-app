package handlers

import (
	"archive/zip"
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/via-justa/overpacked-app/backend/internal/api"
	"github.com/via-justa/overpacked-app/backend/internal/backup"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

const (
	maxImportBytes      = 512 << 20 // 512 MiB
	maxImportFormMemory = 32 << 20  // 32 MiB kept in memory while parsing
)

const (
	backupErrConfigLoad   = "failed to load backup config"
	backupErrConfigSave   = "failed to save backup config"
	backupErrPassword     = "password confirmation failed"
	backupErrInvalidCron  = "invalid schedule expression"
	backupErrImport       = "failed to import backup"
	backupErrRun          = "failed to run backup"
	backupErrExport       = "failed to build export"
	backupErrInvalidInput = "invalid request"
)

type BackupHandler struct {
	service     *backup.Service
	store       *store.Store
	scheduler   *backup.Scheduler
	appPassword string
}

func NewBackupHandler(service *backup.Service, st *store.Store, scheduler *backup.Scheduler, appPassword string) *BackupHandler {
	return &BackupHandler{service: service, store: st, scheduler: scheduler, appPassword: appPassword}
}

// ExportBackup streams a full backup ZIP as a download.
func (h *BackupHandler) ExportBackup(w http.ResponseWriter, r *http.Request) {
	filename := fmt.Sprintf("overpacked-backup-%s.zip", time.Now().UTC().Format("20060102-150405"))
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))

	if err := h.service.BuildArchive(r.Context(), w); err != nil {
		// Headers are already sent; surface the failure in logs via the panic-free path.
		http.Error(w, backupErrExport, http.StatusInternalServerError)
	}
}

// ImportBackup restores data from an uploaded backup ZIP.
func (h *BackupHandler) ImportBackup(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxImportBytes)
	if err := r.ParseMultipartForm(maxImportFormMemory); err != nil {
		writeError(w, http.StatusBadRequest, backupErrInvalidInput)
		return
	}

	if subtle.ConstantTimeCompare([]byte(r.FormValue("password")), []byte(h.appPassword)) != 1 {
		writeError(w, http.StatusUnauthorized, backupErrPassword)
		return
	}

	mode := backup.Mode(r.FormValue("mode"))

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing backup file")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read backup file")
		return
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		writeError(w, http.StatusBadRequest, "uploaded file is not a valid archive")
		return
	}

	result, err := h.service.Import(r.Context(), zr, mode)
	if err != nil {
		h.writeImportError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, importResultToAPI(result))
}

func (h *BackupHandler) writeImportError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, backup.ErrInvalidMode),
		errors.Is(err, backup.ErrInvalidArchive),
		errors.Is(err, backup.ErrUnsupportedVersion):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, backupErrImport)
	}
}

// GetBackupConfig returns the scheduled-backup configuration.
func (h *BackupHandler) GetBackupConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.store.BackupConfig.Get(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, backupErrConfigLoad)
		return
	}
	writeJSON(w, http.StatusOK, backupConfigToAPI(cfg))
}

// UpdateBackupConfig validates and persists the configuration, then reloads the scheduler.
func (h *BackupHandler) UpdateBackupConfig(w http.ResponseWriter, r *http.Request) {
	var req api.BackupConfigUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, backupErrInvalidInput)
		return
	}
	defer r.Body.Close()

	if _, err := cron.ParseStandard(req.Schedule); err != nil {
		writeError(w, http.StatusBadRequest, backupErrInvalidCron)
		return
	}
	if req.RetentionCount < 1 {
		writeError(w, http.StatusBadRequest, "retention count must be at least 1")
		return
	}

	cfg, err := h.store.BackupConfig.Get(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, backupErrConfigLoad)
		return
	}
	cfg.Enabled = req.Enabled
	cfg.Schedule = req.Schedule
	cfg.RetentionCount = req.RetentionCount

	if err := h.store.BackupConfig.Update(r.Context(), cfg); err != nil {
		writeError(w, http.StatusInternalServerError, backupErrConfigSave)
		return
	}

	if err := h.scheduler.Reload(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to apply schedule")
		return
	}

	writeJSON(w, http.StatusOK, backupConfigToAPI(cfg))
}

// RunBackup writes a backup to the configured path immediately. It is non-destructive
// (it only creates a file) so it relies on the JWT auth middleware rather than a password.
func (h *BackupHandler) RunBackup(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.store.BackupConfig.Get(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, backupErrConfigLoad)
		return
	}

	ranAt := time.Now()
	path, runErr := h.service.Run(r.Context(), cfg.RetentionCount, ranAt)
	if statusErr := h.store.BackupConfig.UpdateRunStatus(r.Context(), ranAt, runErr); statusErr != nil {
		// Non-fatal: the run outcome below is what the caller cares about.
		_ = statusErr
	}
	if runErr != nil {
		if errors.Is(runErr, backup.ErrBackupDirNotConfigured) {
			writeError(w, http.StatusBadRequest, runErr.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, backupErrRun)
		return
	}

	writeJSON(w, http.StatusOK, api.BackupRunResult{Path: path})
}

// ExportItems streams a generic items-only CSV (or ZIP with images) export.
func (h *BackupHandler) ExportItems(w http.ResponseWriter, r *http.Request, params api.ExportItemsParams) {
	includeImages := params.IncludeImages != nil && *params.IncludeImages

	if includeImages {
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="items-export.zip"`)
	} else {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", `attachment; filename="items.csv"`)
	}

	if err := h.service.ExportItemsCSV(r.Context(), w, includeImages); err != nil {
		http.Error(w, backupErrExport, http.StatusInternalServerError)
	}
}

func backupConfigToAPI(c *domain.BackupConfig) api.BackupConfig {
	out := api.BackupConfig{
		Enabled:        c.Enabled,
		RetentionCount: c.RetentionCount,
		Schedule:       c.Schedule,
		LastRunAt:      c.LastRunAt,
		LastError:      c.LastError,
	}
	if c.LastStatus != nil {
		status := api.BackupConfigLastStatus(*c.LastStatus)
		out.LastStatus = &status
	}
	return out
}

func importResultToAPI(r backup.ImportResult) api.BackupImportResult {
	return api.BackupImportResult{
		Mode:   api.BackupImportResultMode(r.Mode),
		Counts: r.Counts,
	}
}
