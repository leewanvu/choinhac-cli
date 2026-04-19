package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"choinhaccli/internal/library"

	"github.com/go-chi/chi/v5"
)

type PlaylistHandler struct {
	store *library.Store
}

func NewPlaylistHandler(store *library.Store) *PlaylistHandler {
	return &PlaylistHandler{store: store}
}

func (h *PlaylistHandler) List(w http.ResponseWriter, r *http.Request) {
	playlists, err := h.store.ListPlaylists()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if playlists == nil {
		playlists = []library.Playlist{}
	}
	jsonOK(w, map[string]any{"playlists": playlists})
}

func (h *PlaylistHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		jsonError(w, "name required", http.StatusBadRequest)
		return
	}
	id, err := h.store.CreatePlaylist(body.Name)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonCreated(w, map[string]any{"id": id, "name": body.Name})
}

func (h *PlaylistHandler) Rename(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		jsonError(w, "name required", http.StatusBadRequest)
		return
	}
	if err := h.store.RenamePlaylist(id, body.Name); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]any{"ok": true})
}

func (h *PlaylistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.store.DeletePlaylist(id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PlaylistHandler) GetTracks(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	tracks, err := h.store.GetPlaylistTracks(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if tracks == nil {
		tracks = []library.Track{}
	}
	jsonOK(w, map[string]any{"tracks": tracks})
}

func (h *PlaylistHandler) AddTrack(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var body struct {
		TrackID int64 `json:"track_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.TrackID == 0 {
		jsonError(w, "track_id required", http.StatusBadRequest)
		return
	}
	if err := h.store.AddTrackToPlaylist(pid, body.TrackID); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonCreated(w, map[string]any{"ok": true})
}

func (h *PlaylistHandler) RemoveTrack(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	tid, err := parseID(r, "track_id")
	if err != nil {
		jsonError(w, "invalid track_id", http.StatusBadRequest)
		return
	}
	if err := h.store.RemoveTrackFromPlaylist(pid, tid); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PlaylistHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var body struct {
		TrackIDs []int64 `json:"track_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.store.ReorderPlaylist(pid, body.TrackIDs); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]any{"ok": true})
}

func parseID(r *http.Request, param string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, param), 10, 64)
}
