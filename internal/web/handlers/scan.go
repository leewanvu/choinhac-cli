package handlers

import (
	"net/http"
	"sync"

	"choinhaccli/internal/library"
)

type ScanHandler struct {
	store    *library.Store
	musicDir string
	coverDir string

	mu      sync.RWMutex
	running bool
	scanned int
	total   int
	done    bool
}

func NewScanHandler(store *library.Store, musicDir, coverDir string) *ScanHandler {
	return &ScanHandler{store: store, musicDir: musicDir, coverDir: coverDir}
}

func (h *ScanHandler) StartScan(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		jsonOK(w, map[string]any{"started": false, "message": "scan already running"})
		return
	}
	h.running = true
	h.done = false
	h.scanned = 0
	h.total = 0
	h.mu.Unlock()

	go func() {
		ch := library.ScanAsync(h.musicDir, h.store, h.coverDir)
		for p := range ch {
			h.mu.Lock()
			h.scanned = p.Scanned
			h.total = p.Total
			if p.Done {
				h.running = false
				h.done = true
			}
			h.mu.Unlock()
		}
	}()

	jsonOK(w, map[string]any{"started": true})
}

func (h *ScanHandler) ScanStatus(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	jsonOK(w, map[string]any{
		"running": h.running,
		"scanned": h.scanned,
		"total":   h.total,
		"done":    h.done,
	})
}
