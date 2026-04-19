package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CoverHandler struct {
	coverDir string
}

func NewCoverHandler(coverDir string) *CoverHandler {
	return &CoverHandler{coverDir: coverDir}
}

func (h *CoverHandler) ServeAlbumCover(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "albumID"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	path := filepath.Join(h.coverDir, fmt.Sprintf("%d.jpg", id))
	if _, err := os.Stat(path); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "max-age=3600")
		_, _ = w.Write([]byte(placeholderSVG))
		return
	}

	w.Header().Set("Cache-Control", "max-age=604800")
	http.ServeFile(w, r, path)
}

const placeholderSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">` +
	`<rect fill="#282828" width="100" height="100"/>` +
	`<text x="50" y="62" text-anchor="middle" fill="#555" font-size="40" font-family="sans-serif">&#9835;</text>` +
	`</svg>`
