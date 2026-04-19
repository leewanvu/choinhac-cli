package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"choinhaccli/internal/library"
	"choinhaccli/internal/web/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	http     *http.Server
	store    *library.Store
	musicDir string
}

func NewServer(addr string, port int, store *library.Store, musicDir, coverDir string) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	lib := handlers.NewLibraryHandler(store)
	stream := handlers.NewStreamHandler(store, musicDir)
	scan := handlers.NewScanHandler(store, musicDir, coverDir)
	playlist := handlers.NewPlaylistHandler(store)
	cover := handlers.NewCoverHandler(coverDir)

	r.Route("/api", func(r chi.Router) {
		r.Get("/library/tracks", lib.ListTracks)
		r.Get("/library/albums", lib.ListAlbums)
		r.Get("/library/albums/{id}", lib.GetAlbumDetail)
		r.Get("/library/artists", lib.ListArtists)
		r.Get("/library/artists/{id}", lib.GetArtistDetail)
		r.Get("/search", lib.Search)

		r.Post("/scan", scan.StartScan)
		r.Get("/scan/status", scan.ScanStatus)

		r.Get("/playlists", playlist.List)
		r.Post("/playlists", playlist.Create)
		r.Put("/playlists/{id}", playlist.Rename)
		r.Delete("/playlists/{id}", playlist.Delete)
		r.Get("/playlists/{id}/tracks", playlist.GetTracks)
		r.Post("/playlists/{id}/tracks", playlist.AddTrack)
		r.Delete("/playlists/{id}/tracks/{track_id}", playlist.RemoveTrack)
		r.Put("/playlists/{id}/reorder", playlist.Reorder)
	})

	r.Get("/stream/{id}", stream.Stream)
	r.Get("/cover/{albumID}", cover.ServeAlbumCover)

	r.Handle("/*", SPAHandler())

	return &Server{
		http: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", addr, port),
			Handler:      r,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 0, // streaming — no write deadline
		},
		store:    store,
		musicDir: musicDir,
	}
}

func (s *Server) Start() error {
	log.Printf("musicweb listening on http://%s", s.http.Addr)
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// corsMiddleware allows Vite dev server (localhost:5173) during development.
var devOrigins = map[string]bool{
	"http://localhost:5173":  true,
	"http://127.0.0.1:5173": true,
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if devOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
