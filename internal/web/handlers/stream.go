package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"choinhaccli/internal/library"

	"github.com/go-chi/chi/v5"
)

type StreamHandler struct {
	store    *library.Store
	musicDir string
}

func NewStreamHandler(store *library.Store, musicDir string) *StreamHandler {
	return &StreamHandler{store: store, musicDir: musicDir}
}

var mimeTypes = map[string]string{
	"flac": "audio/flac",
	"wav":  "audio/wav",
	"mp3":  "audio/mpeg",
}

func (h *StreamHandler) Stream(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	track, err := h.store.GetTrack(id)
	if err != nil {
		http.Error(w, "track not found", http.StatusNotFound)
		return
	}

	// Path traversal guard: file must be under music dir
	abs, err := filepath.Abs(track.FilePath)
	if err != nil {
		http.Error(w, "invalid path", http.StatusInternalServerError)
		return
	}
	musicAbs, err := filepath.Abs(h.musicDir)
	if err != nil {
		http.Error(w, "invalid music dir", http.StatusInternalServerError)
		return
	}
	if !strings.HasPrefix(abs, musicAbs+string(filepath.Separator)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	f, err := os.Open(abs)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	mime := mimeTypes[track.Format]
	if mime == "" {
		mime = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Accept-Ranges", "bytes")

	http.ServeContent(w, r, filepath.Base(abs), time.Time{}, f)
}
