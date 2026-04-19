package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"choinhaccli/internal/library"

	"github.com/go-chi/chi/v5"
)

type LibraryHandler struct {
	store *library.Store
}

func NewLibraryHandler(store *library.Store) *LibraryHandler {
	return &LibraryHandler{store: store}
}

func (h *LibraryHandler) ListTracks(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	tracks, err := h.store.ListTracks(limit, offset)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if tracks == nil {
		tracks = []library.Track{}
	}
	jsonOK(w, map[string]any{
		"tracks": tracks,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *LibraryHandler) ListAlbums(w http.ResponseWriter, r *http.Request) {
	albums, err := h.store.ListAlbums()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if albums == nil {
		albums = []library.Album{}
	}
	jsonOK(w, map[string]any{"albums": albums})
}

func (h *LibraryHandler) ListArtists(w http.ResponseWriter, r *http.Request) {
	artists, err := h.store.ListArtists()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if artists == nil {
		artists = []library.Artist{}
	}
	jsonOK(w, map[string]any{"artists": artists})
}

func (h *LibraryHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		jsonOK(w, map[string]any{"tracks": []any{}, "albums": []any{}, "artists": []any{}})
		return
	}
	tracks, err := h.store.SearchTracks(q, 20)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	albums, err := h.store.SearchAlbums(q, 10)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	artists, err := h.store.SearchArtists(q, 10)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if tracks == nil {
		tracks = []library.Track{}
	}
	if albums == nil {
		albums = []library.Album{}
	}
	if artists == nil {
		artists = []library.Artist{}
	}
	jsonOK(w, map[string]any{"tracks": tracks, "albums": albums, "artists": artists})
}

func (h *LibraryHandler) GetAlbumDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	album, err := h.store.GetAlbum(id)
	if err != nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	tracks, err := h.store.ListAlbumTracks(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if tracks == nil {
		tracks = []library.Track{}
	}
	jsonOK(w, map[string]any{"album": album, "tracks": tracks})
}

func (h *LibraryHandler) GetArtistDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	artist, err := h.store.GetArtist(id)
	if err != nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	albums, err := h.store.ListArtistAlbums(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if albums == nil {
		albums = []library.Album{}
	}
	jsonOK(w, map[string]any{"artist": artist, "albums": albums})
}

// --- helpers ---

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return def
	}
	return n
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func jsonCreated(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
