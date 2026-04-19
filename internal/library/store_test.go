package library

import (
	"testing"
	"time"
)

// helper to create an in-memory test store
func newTestStore(t *testing.T) *Store {
	// Use in-memory SQLite with shared cache for concurrent access
	store, err := OpenStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("OpenStore failed: %v", err)
	}
	return store
}

// helper to create test data
func createTestTrack() Track {
	return Track{
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

func TestOpenStore(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestUpsertTrack(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	track := createTestTrack()
	err := store.UpsertTrack(track)
	if err != nil {
		t.Fatalf("UpsertTrack failed: %v", err)
	}

	// verify track was inserted
	tracks, err := store.ListTracks(10, 0)
	if err != nil {
		t.Fatalf("ListTracks failed: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}
	if tracks[0].Title != track.Title {
		t.Errorf("expected title %q, got %q", track.Title, tracks[0].Title)
	}
}

func TestUpsertTrackUpdate(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	track := createTestTrack()
	err := store.UpsertTrack(track)
	if err != nil {
		t.Fatalf("first UpsertTrack failed: %v", err)
	}

	// update same track (by file path)
	track.Title = "Updated Song"
	track.Mtime = 2000000
	err = store.UpsertTrack(track)
	if err != nil {
		t.Fatalf("second UpsertTrack failed: %v", err)
	}

	// verify update worked
	tracks, err := store.ListTracks(10, 0)
	if err != nil {
		t.Fatalf("ListTracks failed: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}
	if tracks[0].Title != "Updated Song" {
		t.Errorf("expected updated title, got %q", tracks[0].Title)
	}
	if tracks[0].Mtime != 2000000 {
		t.Errorf("expected mtime 2000000, got %d", tracks[0].Mtime)
	}
}

func TestListTracks(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// insert 5 tracks
	for i := 1; i <= 5; i++ {
		track := createTestTrack()
		track.Title = "Song " + string(rune('0'+i))
		track.FilePath = "/music/song" + string(rune('0'+i)) + ".flac"
		if err := store.UpsertTrack(track); err != nil {
			t.Fatalf("UpsertTrack failed: %v", err)
		}
	}

	// test limit
	tracks, err := store.ListTracks(2, 0)
	if err != nil {
		t.Fatalf("ListTracks failed: %v", err)
	}
	if len(tracks) != 2 {
		t.Errorf("expected 2 tracks, got %d", len(tracks))
	}

	// test offset
	tracks, err = store.ListTracks(2, 2)
	if err != nil {
		t.Fatalf("ListTracks with offset failed: %v", err)
	}
	if len(tracks) != 2 {
		t.Errorf("expected 2 tracks with offset, got %d", len(tracks))
	}

	// test all
	tracks, err = store.ListTracks(100, 0)
	if err != nil {
		t.Fatalf("ListTracks all failed: %v", err)
	}
	if len(tracks) != 5 {
		t.Errorf("expected 5 tracks total, got %d", len(tracks))
	}
}

func TestListTracksEmpty(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tracks, err := store.ListTracks(10, 0)
	if err != nil {
		t.Fatalf("ListTracks failed: %v", err)
	}
	if len(tracks) != 0 {
		t.Errorf("expected 0 tracks, got %d", len(tracks))
	}
}

func TestSearchTracks(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// insert test data
	testTracks := []Track{
		{Title: "Midnight Dreams", Artist: "Luna", Album: "Night", TrackNum: 1, Duration: 200, Format: "flac", FilePath: "/music/1.flac", Mtime: 1000},
		{Title: "Solar Wind", Artist: "Sun", Album: "Day", TrackNum: 1, Duration: 180, Format: "flac", FilePath: "/music/2.flac", Mtime: 1000},
		{Title: "Lunar Glow", Artist: "Moon", Album: "Night", TrackNum: 1, Duration: 220, Format: "flac", FilePath: "/music/3.flac", Mtime: 1000},
	}
	for _, track := range testTracks {
		if err := store.UpsertTrack(track); err != nil {
			t.Fatalf("UpsertTrack failed: %v", err)
		}
	}

	// search by title
	results, err := store.SearchTracks("Midnight", 10)
	if err != nil {
		t.Fatalf("SearchTracks failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'Midnight', got %d", len(results))
	}

	// search by artist - Luna appears in both title and artist
	results, err = store.SearchTracks("Luna", 10)
	if err != nil {
		t.Fatalf("SearchTracks failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'Luna', got %d", len(results))
	}

	// search by album
	results, err = store.SearchTracks("Night", 10)
	if err != nil {
		t.Fatalf("SearchTracks failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'Night', got %d", len(results))
	}

	// search with limit
	results, err = store.SearchTracks("Night", 1)
	if err != nil {
		t.Fatalf("SearchTracks with limit failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result with limit, got %d", len(results))
	}
}

func TestSearchTracksEmpty(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	results, err := store.SearchTracks("nonexistent", 10)
	if err != nil {
		t.Fatalf("SearchTracks failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestCreatePlaylist(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	id, err := store.CreatePlaylist("My Playlist")
	if err != nil {
		t.Fatalf("CreatePlaylist failed: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}

	// verify playlist exists
	playlists, err := store.ListPlaylists()
	if err != nil {
		t.Fatalf("ListPlaylists failed: %v", err)
	}
	if len(playlists) != 1 {
		t.Fatalf("expected 1 playlist, got %d", len(playlists))
	}
	if playlists[0].Name != "My Playlist" {
		t.Errorf("expected name 'My Playlist', got %q", playlists[0].Name)
	}
}

func TestAddTrackToPlaylist(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// insert track
	track := createTestTrack()
	if err := store.UpsertTrack(track); err != nil {
		t.Fatalf("UpsertTrack failed: %v", err)
	}

	// get track id
	tracks, err := store.ListTracks(1, 0)
	if err != nil || len(tracks) == 0 {
		t.Fatalf("ListTracks failed: %v", err)
	}
	trackID := tracks[0].ID

	// create playlist
	playlistID, err := store.CreatePlaylist("Test")
	if err != nil {
		t.Fatalf("CreatePlaylist failed: %v", err)
	}

	// add track to playlist
	err = store.AddTrackToPlaylist(playlistID, trackID)
	if err != nil {
		t.Fatalf("AddTrackToPlaylist failed: %v", err)
	}

	// verify
	playlistTracks, err := store.GetPlaylistTracks(playlistID)
	if err != nil {
		t.Fatalf("GetPlaylistTracks failed: %v", err)
	}
	if len(playlistTracks) != 1 {
		t.Errorf("expected 1 track in playlist, got %d", len(playlistTracks))
	}
	if playlistTracks[0].ID != trackID {
		t.Errorf("expected track %d, got %d", trackID, playlistTracks[0].ID)
	}
}

func TestReorderPlaylist(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// insert 3 tracks
	var trackIDs []int64
	for i := 1; i <= 3; i++ {
		track := createTestTrack()
		track.Title = "Song " + string(rune('0'+i))
		track.FilePath = "/music/song" + string(rune('0'+i)) + ".flac"
		if err := store.UpsertTrack(track); err != nil {
			t.Fatalf("UpsertTrack failed: %v", err)
		}
	}

	// get track ids
	tracks, err := store.ListTracks(10, 0)
	if err != nil {
		t.Fatalf("ListTracks failed: %v", err)
	}
	for _, track := range tracks {
		trackIDs = append(trackIDs, track.ID)
	}

	// create playlist and add tracks
	playlistID, err := store.CreatePlaylist("Test")
	if err != nil {
		t.Fatalf("CreatePlaylist failed: %v", err)
	}
	for _, tid := range trackIDs {
		if err := store.AddTrackToPlaylist(playlistID, tid); err != nil {
			t.Fatalf("AddTrackToPlaylist failed: %v", err)
		}
	}

	// reverse the order
	reversed := make([]int64, len(trackIDs))
	for i, j := 0, len(trackIDs)-1; i < len(trackIDs); i, j = i+1, j-1 {
		reversed[i] = trackIDs[j]
	}

	// reorder
	if err := store.ReorderPlaylist(playlistID, reversed); err != nil {
		t.Fatalf("ReorderPlaylist failed: %v", err)
	}

	// verify new order
	playlistTracks, err := store.GetPlaylistTracks(playlistID)
	if err != nil {
		t.Fatalf("GetPlaylistTracks failed: %v", err)
	}
	if len(playlistTracks) != 3 {
		t.Fatalf("expected 3 tracks, got %d", len(playlistTracks))
	}
	for i, track := range playlistTracks {
		if track.ID != reversed[i] {
			t.Errorf("position %d: expected track %d, got %d", i, reversed[i], track.ID)
		}
	}
}

func TestGetTrackMtime(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	filePath := "/music/test.flac"
	mtime := int64(1234567890)

	// initially should return 0 (not found)
	result, err := store.GetTrackMtime(filePath)
	if err != nil {
		t.Fatalf("GetTrackMtime failed: %v", err)
	}
	if result != 0 {
		t.Errorf("expected 0 for nonexistent file, got %d", result)
	}

	// insert track
	track := createTestTrack()
	track.FilePath = filePath
	track.Mtime = mtime
	if err := store.UpsertTrack(track); err != nil {
		t.Fatalf("UpsertTrack failed: %v", err)
	}

	// now should return the mtime
	result, err = store.GetTrackMtime(filePath)
	if err != nil {
		t.Fatalf("GetTrackMtime failed: %v", err)
	}
	if result != mtime {
		t.Errorf("expected mtime %d, got %d", mtime, result)
	}
}

func TestGetTrackMtimeIncrementalScan(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// simulate incremental scan behavior
	filePath := "/music/test.flac"

	// first scan
	mtime1 := time.Now().Unix()
	track := createTestTrack()
	track.FilePath = filePath
	track.Mtime = mtime1
	if err := store.UpsertTrack(track); err != nil {
		t.Fatalf("first UpsertTrack failed: %v", err)
	}

	result1, err := store.GetTrackMtime(filePath)
	if err != nil {
		t.Fatalf("first GetTrackMtime failed: %v", err)
	}
	if result1 != mtime1 {
		t.Errorf("first scan: expected %d, got %d", mtime1, result1)
	}

	// second scan (file modified)
	mtime2 := mtime1 + 1000
	track.Mtime = mtime2
	if err := store.UpsertTrack(track); err != nil {
		t.Fatalf("second UpsertTrack failed: %v", err)
	}

	result2, err := store.GetTrackMtime(filePath)
	if err != nil {
		t.Fatalf("second GetTrackMtime failed: %v", err)
	}
	if result2 != mtime2 {
		t.Errorf("second scan: expected %d, got %d", mtime2, result2)
	}
}

func TestCreatePlaylistDuplicate(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	_, err := store.CreatePlaylist("Same")
	if err != nil {
		t.Fatalf("first CreatePlaylist failed: %v", err)
	}

	// second playlist with same name should fail (unique constraint)
	_, err = store.CreatePlaylist("Same")
	if err == nil {
		t.Fatal("expected error for duplicate playlist name")
	}
}

func TestRenamePlaylist(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	id, err := store.CreatePlaylist("Original")
	if err != nil {
		t.Fatalf("CreatePlaylist failed: %v", err)
	}

	err = store.RenamePlaylist(id, "Renamed")
	if err != nil {
		t.Fatalf("RenamePlaylist failed: %v", err)
	}

	playlists, err := store.ListPlaylists()
	if err != nil {
		t.Fatalf("ListPlaylists failed: %v", err)
	}
	if len(playlists) != 1 || playlists[0].Name != "Renamed" {
		t.Errorf("expected name 'Renamed', got %q", playlists[0].Name)
	}
}

func TestDeletePlaylist(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	id, err := store.CreatePlaylist("ToDelete")
	if err != nil {
		t.Fatalf("CreatePlaylist failed: %v", err)
	}

	err = store.DeletePlaylist(id)
	if err != nil {
		t.Fatalf("DeletePlaylist failed: %v", err)
	}

	playlists, err := store.ListPlaylists()
	if err != nil {
		t.Fatalf("ListPlaylists failed: %v", err)
	}
	if len(playlists) != 0 {
		t.Errorf("expected 0 playlists after delete, got %d", len(playlists))
	}
}

func TestRemoveTrackFromPlaylist(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// setup
	track := createTestTrack()
	if err := store.UpsertTrack(track); err != nil {
		t.Fatalf("UpsertTrack failed: %v", err)
	}

	tracks, _ := store.ListTracks(1, 0)
	trackID := tracks[0].ID

	playlistID, _ := store.CreatePlaylist("Test")
	store.AddTrackToPlaylist(playlistID, trackID)

	// remove
	err := store.RemoveTrackFromPlaylist(playlistID, trackID)
	if err != nil {
		t.Fatalf("RemoveTrackFromPlaylist failed: %v", err)
	}

	// verify removed
	playlistTracks, _ := store.GetPlaylistTracks(playlistID)
	if len(playlistTracks) != 0 {
		t.Errorf("expected 0 tracks after removal, got %d", len(playlistTracks))
	}
}

func TestListAlbums(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// insert tracks in different albums
	albums := []struct {
		title  string
		artist string
	}{
		{"Album A", "Artist 1"},
		{"Album B", "Artist 2"},
	}

	for i, a := range albums {
		track := createTestTrack()
		track.Album = a.title
		track.Artist = a.artist
		track.FilePath = "/music/track" + string(rune('0'+i)) + ".flac"
		if err := store.UpsertTrack(track); err != nil {
			t.Fatalf("UpsertTrack failed: %v", err)
		}
	}

	results, err := store.ListAlbums()
	if err != nil {
		t.Fatalf("ListAlbums failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 albums, got %d", len(results))
	}
}

func TestListArtists(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// insert tracks by different artists
	artists := []string{"Artist A", "Artist B", "Artist C"}

	for i, artist := range artists {
		track := createTestTrack()
		track.Artist = artist
		track.FilePath = "/music/track" + string(rune('0'+i)) + ".flac"
		if err := store.UpsertTrack(track); err != nil {
			t.Fatalf("UpsertTrack failed: %v", err)
		}
	}

	results, err := store.ListArtists()
	if err != nil {
		t.Fatalf("ListArtists failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 artists, got %d", len(results))
	}
}

func TestListPlaylistsEmpty(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	playlists, err := store.ListPlaylists()
	if err != nil {
		t.Fatalf("ListPlaylists failed: %v", err)
	}
	if len(playlists) != 0 {
		t.Errorf("expected 0 playlists, got %d", len(playlists))
	}
}

func TestGetPlaylistTracksEmpty(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	playlistID, _ := store.CreatePlaylist("Empty")
	tracks, err := store.GetPlaylistTracks(playlistID)
	if err != nil {
		t.Fatalf("GetPlaylistTracks failed: %v", err)
	}
	if len(tracks) != 0 {
		t.Errorf("expected 0 tracks in empty playlist, got %d", len(tracks))
	}
}

func TestListAlbumTracks(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// insert tracks in same album
	for i := 1; i <= 3; i++ {
		track := createTestTrack()
		track.Album = "Single Album"
		track.TrackNum = i
		track.FilePath = "/music/track" + string(rune('0'+i)) + ".flac"
		if err := store.UpsertTrack(track); err != nil {
			t.Fatalf("UpsertTrack failed: %v", err)
		}
	}

	// get album id
	albums, err := store.ListAlbums()
	if err != nil {
		t.Fatalf("ListAlbums failed: %v", err)
	}
	if len(albums) != 1 {
		t.Fatalf("expected 1 album, got %d", len(albums))
	}

	// get album tracks
	tracks, err := store.ListAlbumTracks(albums[0].ID)
	if err != nil {
		t.Fatalf("ListAlbumTracks failed: %v", err)
	}
	if len(tracks) != 3 {
		t.Errorf("expected 3 tracks, got %d", len(tracks))
	}

	// verify ordered by track number
	for i, track := range tracks {
		if track.TrackNum != i+1 {
			t.Errorf("expected track number %d, got %d", i+1, track.TrackNum)
		}
	}
}

func TestGetTrack(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	track := createTestTrack()
	track.Title = "Specific Track"
	if err := store.UpsertTrack(track); err != nil {
		t.Fatalf("UpsertTrack failed: %v", err)
	}

	// get track id
	tracks, _ := store.ListTracks(1, 0)
	trackID := tracks[0].ID

	// fetch by id
	retrieved, err := store.GetTrack(trackID)
	if err != nil {
		t.Fatalf("GetTrack failed: %v", err)
	}
	if retrieved.Title != "Specific Track" {
		t.Errorf("expected title 'Specific Track', got %q", retrieved.Title)
	}
	if retrieved.ID != trackID {
		t.Errorf("expected id %d, got %d", trackID, retrieved.ID)
	}
}

func TestSearchAlbums(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// insert test data
	albumData := []struct {
		title  string
		artist string
	}{
		{"Midnight Dreams", "Luna"},
		{"Solar Energy", "Sun"},
		{"Lunar Nights", "Moon"},
	}

	for i, a := range albumData {
		track := createTestTrack()
		track.Album = a.title
		track.Artist = a.artist
		track.FilePath = "/music/track" + string(rune('0'+i)) + ".flac"
		if err := store.UpsertTrack(track); err != nil {
			t.Fatalf("UpsertTrack failed: %v", err)
		}
	}

	// search
	results, err := store.SearchAlbums("Lunar", 10)
	if err != nil {
		t.Fatalf("SearchAlbums failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestSearchArtists(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// insert test data
	artists := []string{"Luna Artist", "Sun", "Moon Artist"}
	for i, a := range artists {
		track := createTestTrack()
		track.Artist = a
		track.FilePath = "/music/track" + string(rune('0'+i)) + ".flac"
		if err := store.UpsertTrack(track); err != nil {
			t.Fatalf("UpsertTrack failed: %v", err)
		}
	}

	// search
	results, err := store.SearchArtists("Artist", 10)
	if err != nil {
		t.Fatalf("SearchArtists failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestCloseStore(t *testing.T) {
	store := newTestStore(t)
	err := store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestStoreConcurrentAccess(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// simple concurrent write test
	done := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			track := createTestTrack()
			track.Title = "Song " + string(rune('0'+idx))
			track.FilePath = "/music/track" + string(rune('0'+idx)) + ".flac"
			done <- store.UpsertTrack(track)
		}(i)
	}

	// wait for all goroutines
	for i := 0; i < 5; i++ {
		err := <-done
		if err != nil {
			t.Errorf("concurrent UpsertTrack failed: %v", err)
		}
	}

	// verify all tracks inserted
	tracks, _ := store.ListTracks(10, 0)
	if len(tracks) != 5 {
		t.Errorf("expected 5 tracks after concurrent inserts, got %d", len(tracks))
	}
}
