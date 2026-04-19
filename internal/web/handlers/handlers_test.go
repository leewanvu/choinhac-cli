package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"choinhaccli/internal/library"
)

// helper to create test store
func newTestStore(t *testing.T) *library.Store {
	store, err := library.OpenStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("OpenStore failed: %v", err)
	}
	return store
}

// helper to create test track
func createTestTrack() library.Track {
	return library.Track{
		Title:    "Test Song",
		Artist:   "Test Artist",
		Album:    "Test Album",
		TrackNum: 1,
		Duration: 180,
		Format:   "flac",
		FilePath: "/music/test.flac",
		Mtime:    1000000,
	}
}

// --- LibraryHandler tests ---

func TestListTracksHandler(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// insert test track
	track := createTestTrack()
	if err := store.UpsertTrack(track); err != nil {
		t.Fatalf("UpsertTrack failed: %v", err)
	}

	handler := NewLibraryHandler(store)
	req := httptest.NewRequest("GET", "/api/library/tracks", nil)
	w := httptest.NewRecorder()

	handler.ListTracks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	// parse response
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// verify "tracks" key exists
	if _, ok := resp["tracks"]; !ok {
		t.Fatal("response missing 'tracks' key")
	}

	tracks := resp["tracks"].([]any)
	if len(tracks) != 1 {
		t.Errorf("expected 1 track, got %d", len(tracks))
	}
}

func TestListTracksHandlerEmpty(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	handler := NewLibraryHandler(store)
	req := httptest.NewRequest("GET", "/api/library/tracks", nil)
	w := httptest.NewRecorder()

	handler.ListTracks(w, req)

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	tracks := resp["tracks"].([]any)
	if len(tracks) != 0 {
		t.Errorf("expected 0 tracks, got %d", len(tracks))
	}
}

func TestListTracksHandlerWithLimit(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// insert 5 tracks
	for i := 1; i <= 5; i++ {
		track := createTestTrack()
		track.Title = "Song " + string(rune('0'+i))
		track.FilePath = "/music/song" + string(rune('0'+i)) + ".flac"
		store.UpsertTrack(track)
	}

	handler := NewLibraryHandler(store)
	req := httptest.NewRequest("GET", "/api/library/tracks?limit=2", nil)
	w := httptest.NewRecorder()

	handler.ListTracks(w, req)

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	tracks := resp["tracks"].([]any)
	if len(tracks) != 2 {
		t.Errorf("expected 2 tracks with limit=2, got %d", len(tracks))
	}

	if resp["limit"].(float64) != 2 {
		t.Errorf("expected limit 2 in response")
	}
}

func TestListTracksHandlerLimitValidation(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	handler := NewLibraryHandler(store)

	tests := []struct {
		query     string
		expectMax int
	}{
		{"?limit=-1", 50},      // negative becomes default
		{"?limit=0", 50},       // zero becomes default
		{"?limit=50", 50},      // valid limit
		{"?limit=201", 200},    // over max capped at 200
		{"?limit=invalid", 50}, // invalid becomes default
	}

	for _, test := range tests {
		req := httptest.NewRequest("GET", "/api/library/tracks"+test.query, nil)
		w := httptest.NewRecorder()

		handler.ListTracks(w, req)

		var resp map[string]any
		json.NewDecoder(w.Body).Decode(&resp)

		if int(resp["limit"].(float64)) != test.expectMax {
			t.Errorf("query %s: expected limit %d, got %d", test.query, test.expectMax, int(resp["limit"].(float64)))
		}
	}
}

func TestListTracksHandlerWithOffset(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// insert 3 tracks
	for i := 1; i <= 3; i++ {
		track := createTestTrack()
		track.Title = "Song " + string(rune('0'+i))
		track.FilePath = "/music/song" + string(rune('0'+i)) + ".flac"
		store.UpsertTrack(track)
	}

	handler := NewLibraryHandler(store)
	req := httptest.NewRequest("GET", "/api/library/tracks?limit=10&offset=1", nil)
	w := httptest.NewRecorder()

	handler.ListTracks(w, req)

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	if int(resp["offset"].(float64)) != 1 {
		t.Errorf("expected offset 1 in response")
	}
}

func TestListAlbumsHandler(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	track := createTestTrack()
	store.UpsertTrack(track)

	handler := NewLibraryHandler(store)
	req := httptest.NewRequest("GET", "/api/library/albums", nil)
	w := httptest.NewRecorder()

	handler.ListAlbums(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	if _, ok := resp["albums"]; !ok {
		t.Fatal("response missing 'albums' key")
	}
}

func TestListArtistsHandler(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	track := createTestTrack()
	store.UpsertTrack(track)

	handler := NewLibraryHandler(store)
	req := httptest.NewRequest("GET", "/api/library/artists", nil)
	w := httptest.NewRecorder()

	handler.ListArtists(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	if _, ok := resp["artists"]; !ok {
		t.Fatal("response missing 'artists' key")
	}
}

func TestSearchHandlerEmptyQuery(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// insert test data
	track := createTestTrack()
	store.UpsertTrack(track)

	handler := NewLibraryHandler(store)
	req := httptest.NewRequest("GET", "/api/search?q=", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	// should return empty arrays
	tracks := resp["tracks"].([]any)
	albums := resp["albums"].([]any)
	artists := resp["artists"].([]any)

	if len(tracks) != 0 {
		t.Errorf("expected 0 tracks for empty query, got %d", len(tracks))
	}
	if len(albums) != 0 {
		t.Errorf("expected 0 albums for empty query, got %d", len(albums))
	}
	if len(artists) != 0 {
		t.Errorf("expected 0 artists for empty query, got %d", len(artists))
	}
}

func TestSearchHandlerNoQuery(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	track := createTestTrack()
	store.UpsertTrack(track)

	handler := NewLibraryHandler(store)
	req := httptest.NewRequest("GET", "/api/search", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	// should return empty arrays (no query means empty query string)
	tracks := resp["tracks"].([]any)
	if len(tracks) != 0 {
		t.Errorf("expected 0 tracks for no query, got %d", len(tracks))
	}
}

func TestSearchHandlerWithQuery(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// insert test tracks
	tracks := []library.Track{
		{Title: "Midnight Dreams", Artist: "Luna", Album: "Night", TrackNum: 1, Duration: 200, Format: "flac", FilePath: "/music/1.flac", Mtime: 1000},
		{Title: "Solar Wind", Artist: "Sun", Album: "Day", TrackNum: 1, Duration: 180, Format: "flac", FilePath: "/music/2.flac", Mtime: 1000},
	}
	for _, t := range tracks {
		store.UpsertTrack(t)
	}

	handler := NewLibraryHandler(store)
	req := httptest.NewRequest("GET", "/api/search?q=Midnight", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	tracks_resp := resp["tracks"].([]any)
	if len(tracks_resp) != 1 {
		t.Errorf("expected 1 track result, got %d", len(tracks_resp))
	}
}

func TestSearchHandlerError(t *testing.T) {
	store := newTestStore(t)
	store.Close() // close to force error

	handler := NewLibraryHandler(store)
	req := httptest.NewRequest("GET", "/api/search?q=test", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)

	if _, ok := resp["error"]; !ok {
		t.Fatal("response missing 'error' key")
	}
}

// --- PlaylistHandler tests ---

func TestCreatePlaylistHandler(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	handler := NewPlaylistHandler(store)

	body := map[string]string{"name": "My Playlist"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/playlists", bytes.NewReader(payload))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["name"] != "My Playlist" {
		t.Errorf("expected name 'My Playlist', got %v", resp["name"])
	}

	if resp["id"] == nil {
		t.Fatal("response missing 'id'")
	}
}

func TestCreatePlaylistHandlerMissingName(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	handler := NewPlaylistHandler(store)

	body := map[string]string{"name": ""}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/playlists", bytes.NewReader(payload))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["error"] != "name required" {
		t.Errorf("expected error 'name required', got %s", resp["error"])
	}
}

func TestCreatePlaylistHandlerInvalidBody(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	handler := NewPlaylistHandler(store)

	req := httptest.NewRequest("POST", "/api/playlists", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestListPlaylistsHandler(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// create test playlist
	store.CreatePlaylist("Test Playlist")

	handler := NewPlaylistHandler(store)
	req := httptest.NewRequest("GET", "/api/playlists", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	if _, ok := resp["playlists"]; !ok {
		t.Fatal("response missing 'playlists' key")
	}

	playlists := resp["playlists"].([]any)
	if len(playlists) != 1 {
		t.Errorf("expected 1 playlist, got %d", len(playlists))
	}
}

func TestListPlaylistsHandlerEmpty(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	handler := NewPlaylistHandler(store)
	req := httptest.NewRequest("GET", "/api/playlists", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	playlists := resp["playlists"].([]any)
	if len(playlists) != 0 {
		t.Errorf("expected 0 playlists, got %d", len(playlists))
	}
}

func TestRenamePlaylistHandlerLogic(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	playlistID, _ := store.CreatePlaylist("Original")

	// Test through store since handler requires chi router URL param parsing
	err := store.RenamePlaylist(playlistID, "Renamed")
	if err != nil {
		t.Fatalf("RenamePlaylist failed: %v", err)
	}

	playlists, _ := store.ListPlaylists()
	if playlists[0].Name != "Renamed" {
		t.Errorf("expected name 'Renamed', got %s", playlists[0].Name)
	}
}

func TestDeletePlaylistHandlerLogic(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	playlistID, _ := store.CreatePlaylist("ToDelete")

	// Test through store since handler requires chi router URL param parsing
	err := store.DeletePlaylist(playlistID)
	if err != nil {
		t.Fatalf("DeletePlaylist failed: %v", err)
	}

	playlists, _ := store.ListPlaylists()
	if len(playlists) != 0 {
		t.Errorf("expected 0 playlists after delete, got %d", len(playlists))
	}
}

// --- ScanHandler tests ---

func TestScanHandlerStartScan(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	handler := NewScanHandler(store, "/tmp/music", "")

	req := httptest.NewRequest("POST", "/api/scan", nil)
	w := httptest.NewRecorder()

	handler.StartScan(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["started"] != true {
		t.Errorf("expected started=true, got %v", resp["started"])
	}
}

func TestScanHandlerAlreadyRunning(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	handler := NewScanHandler(store, "/tmp/music", "")

	// start first scan
	req1 := httptest.NewRequest("POST", "/api/scan", nil)
	w1 := httptest.NewRecorder()
	handler.StartScan(w1, req1)

	var resp1 map[string]any
	json.NewDecoder(w1.Body).Decode(&resp1)

	if resp1["started"] != true {
		t.Fatal("first scan should start")
	}

	// try to start second scan while first is running
	req2 := httptest.NewRequest("POST", "/api/scan", nil)
	w2 := httptest.NewRecorder()
	handler.StartScan(w2, req2)

	var resp2 map[string]any
	json.NewDecoder(w2.Body).Decode(&resp2)

	if resp2["started"] != false {
		t.Errorf("second scan should not start while one is running, got %v", resp2["started"])
	}

	if resp2["message"] != "scan already running" {
		t.Errorf("expected message 'scan already running', got %v", resp2["message"])
	}
}

func TestScanHandlerStatus(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	handler := NewScanHandler(store, "/tmp/music", "")

	req := httptest.NewRequest("GET", "/api/scan/status", nil)
	w := httptest.NewRecorder()

	handler.ScanStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	if _, ok := resp["running"]; !ok {
		t.Fatal("response missing 'running' key")
	}
	if _, ok := resp["scanned"]; !ok {
		t.Fatal("response missing 'scanned' key")
	}
	if _, ok := resp["total"]; !ok {
		t.Fatal("response missing 'total' key")
	}
	if _, ok := resp["done"]; !ok {
		t.Fatal("response missing 'done' key")
	}

	// initial state should be not running
	if resp["running"] != false {
		t.Errorf("initial state should be not running")
	}
}

func TestScanHandlerStatusAfterStart(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	handler := NewScanHandler(store, "/tmp/music", "")

	// start scan
	req1 := httptest.NewRequest("POST", "/api/scan", nil)
	w1 := httptest.NewRecorder()
	handler.StartScan(w1, req1)

	// check status while running
	req2 := httptest.NewRequest("GET", "/api/scan/status", nil)
	w2 := httptest.NewRecorder()
	handler.ScanStatus(w2, req2)

	var resp map[string]any
	json.NewDecoder(w2.Body).Decode(&resp)

	if resp["running"] != true {
		t.Errorf("expected running=true after StartScan, got %v", resp["running"])
	}
}

// --- Error handling tests ---

func TestJsonErrorFunction(t *testing.T) {
	w := httptest.NewRecorder()
	jsonError(w, "test error", http.StatusBadRequest)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["error"] != "test error" {
		t.Errorf("expected error 'test error', got %s", resp["error"])
	}
}

func TestJsonOKFunction(t *testing.T) {
	w := httptest.NewRecorder()
	jsonOK(w, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}
}

func TestJsonCreatedFunction(t *testing.T) {
	w := httptest.NewRecorder()
	jsonCreated(w, map[string]string{"key": "value"})

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}
}

func TestQueryIntHelper(t *testing.T) {
	tests := []struct {
		query      string
		key        string
		defaultVal int
		expected   int
	}{
		{"limit=50", "limit", -1, 50},
		{"offset=10", "offset", -1, 10},
		{"invalid=abc", "invalid", -1, -1},
		{"limit=-5", "limit", -1, -1}, // negative values return default
		{"", "missing", 42, 42},         // missing key returns default
	}

	for _, test := range tests {
		req := httptest.NewRequest("GET", "/api/test?"+test.query, nil)
		result := queryInt(req, test.key, test.defaultVal)

		if result != test.expected {
			t.Errorf("query %q, key %q: expected %d, got %d", test.query, test.key, test.expected, result)
		}
	}
}
