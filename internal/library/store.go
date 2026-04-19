package library

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func OpenStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("foreign_keys: %w", err)
	}
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// --- artists ---

func (s *Store) upsertArtist(name string) (int64, error) {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO artists(name) VALUES(?)`, name)
	if err != nil {
		return 0, err
	}
	var id int64
	return id, s.db.QueryRow(`SELECT id FROM artists WHERE name=?`, name).Scan(&id)
}

// --- albums ---

func (s *Store) upsertAlbum(title string, artistID int64, year int) (int64, error) {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO albums(title, artist_id, year) VALUES(?,?,?)`,
		title, artistID, year,
	)
	if err != nil {
		return 0, err
	}
	var id int64
	return id, s.db.QueryRow(
		`SELECT id FROM albums WHERE title=? AND artist_id=?`, title, artistID,
	).Scan(&id)
}

// --- tracks ---

func (s *Store) UpsertTrack(t Track) error {
	artistID, err := s.upsertArtist(t.Artist)
	if err != nil {
		return err
	}
	albumID, err := s.upsertAlbum(t.Album, artistID, 0)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
		INSERT INTO tracks(title, album_id, artist_id, track_num, duration, format, file_path, mtime)
		VALUES(?,?,?,?,?,?,?,?)
		ON CONFLICT(file_path) DO UPDATE SET
			title=excluded.title, album_id=excluded.album_id, artist_id=excluded.artist_id,
			track_num=excluded.track_num, duration=excluded.duration,
			format=excluded.format, mtime=excluded.mtime
	`, t.Title, albumID, artistID, t.TrackNum, t.Duration, t.Format, t.FilePath, t.Mtime)
	return err
}

func (s *Store) GetAlbumIDForTrack(filePath string) (int64, error) {
	var id int64
	err := s.db.QueryRow(
		`SELECT t.album_id FROM tracks t WHERE t.file_path=?`, filePath,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

func (s *Store) UpdateAlbumCover(id int64, coverPath string) error {
	_, err := s.db.Exec(`UPDATE albums SET cover_path=? WHERE id=?`, coverPath, id)
	return err
}

func (s *Store) GetTrackMtime(filePath string) (int64, error) {
	var mtime int64
	err := s.db.QueryRow(`SELECT mtime FROM tracks WHERE file_path=?`, filePath).Scan(&mtime)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return mtime, err
}

func (s *Store) GetTrack(id int64) (Track, error) {
	row := s.db.QueryRow(`
		SELECT t.id, t.title, t.album_id, al.title, t.artist_id, ar.name,
		       t.track_num, t.duration, t.format, t.file_path, t.mtime
		FROM tracks t
		JOIN albums al ON al.id = t.album_id
		JOIN artists ar ON ar.id = t.artist_id
		WHERE t.id=?`, id)
	return scanTrack(row)
}

func (s *Store) ListTracks(limit, offset int) ([]Track, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.title, t.album_id, al.title, t.artist_id, ar.name,
		       t.track_num, t.duration, t.format, t.file_path, t.mtime
		FROM tracks t
		JOIN albums al ON al.id = t.album_id
		JOIN artists ar ON ar.id = t.artist_id
		ORDER BY ar.name, al.title, t.track_num
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTracks(rows)
}

func (s *Store) ListAlbumTracks(albumID int64) ([]Track, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.title, t.album_id, al.title, t.artist_id, ar.name,
		       t.track_num, t.duration, t.format, t.file_path, t.mtime
		FROM tracks t
		JOIN albums al ON al.id = t.album_id
		JOIN artists ar ON ar.id = t.artist_id
		WHERE t.album_id=?
		ORDER BY t.track_num`, albumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTracks(rows)
}

func (s *Store) ListAlbums() ([]Album, error) {
	rows, err := s.db.Query(`
		SELECT al.id, al.title, al.artist_id, ar.name, al.year, al.cover_path
		FROM albums al JOIN artists ar ON ar.id = al.artist_id
		ORDER BY ar.name, al.title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Album
	for rows.Next() {
		var a Album
		if err := rows.Scan(&a.ID, &a.Title, &a.ArtistID, &a.Artist, &a.Year, &a.CoverPath); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) ListArtists() ([]Artist, error) {
	rows, err := s.db.Query(`SELECT id, name FROM artists ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Artist
	for rows.Next() {
		var a Artist
		if err := rows.Scan(&a.ID, &a.Name); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) SearchTracks(q string, limit int) ([]Track, error) {
	like := "%" + q + "%"
	rows, err := s.db.Query(`
		SELECT t.id, t.title, t.album_id, al.title, t.artist_id, ar.name,
		       t.track_num, t.duration, t.format, t.file_path, t.mtime
		FROM tracks t
		JOIN albums al ON al.id = t.album_id
		JOIN artists ar ON ar.id = t.artist_id
		WHERE t.title LIKE ? OR ar.name LIKE ? OR al.title LIKE ?
		LIMIT ?`, like, like, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTracks(rows)
}

func (s *Store) SearchAlbums(q string, limit int) ([]Album, error) {
	like := "%" + q + "%"
	rows, err := s.db.Query(`
		SELECT al.id, al.title, al.artist_id, ar.name, al.year, al.cover_path
		FROM albums al JOIN artists ar ON ar.id = al.artist_id
		WHERE al.title LIKE ? OR ar.name LIKE ?
		LIMIT ?`, like, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Album
	for rows.Next() {
		var a Album
		if err := rows.Scan(&a.ID, &a.Title, &a.ArtistID, &a.Artist, &a.Year, &a.CoverPath); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) SearchArtists(q string, limit int) ([]Artist, error) {
	like := "%" + q + "%"
	rows, err := s.db.Query(`SELECT id, name FROM artists WHERE name LIKE ? LIMIT ?`, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Artist
	for rows.Next() {
		var a Artist
		if err := rows.Scan(&a.ID, &a.Name); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) GetAlbum(id int64) (Album, error) {
	var a Album
	err := s.db.QueryRow(
		`SELECT al.id, al.title, al.artist_id, ar.name, al.year, al.cover_path
		 FROM albums al JOIN artists ar ON ar.id = al.artist_id
		 WHERE al.id=?`, id,
	).Scan(&a.ID, &a.Title, &a.ArtistID, &a.Artist, &a.Year, &a.CoverPath)
	return a, err
}

func (s *Store) GetArtist(id int64) (Artist, error) {
	var a Artist
	err := s.db.QueryRow(`SELECT id, name FROM artists WHERE id=?`, id).Scan(&a.ID, &a.Name)
	return a, err
}

func (s *Store) ListArtistAlbums(artistID int64) ([]Album, error) {
	rows, err := s.db.Query(
		`SELECT al.id, al.title, al.artist_id, ar.name, al.year, al.cover_path
		 FROM albums al JOIN artists ar ON ar.id = al.artist_id
		 WHERE al.artist_id=?
		 ORDER BY al.year, al.title`, artistID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Album
	for rows.Next() {
		var a Album
		if err := rows.Scan(&a.ID, &a.Title, &a.ArtistID, &a.Artist, &a.Year, &a.CoverPath); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// --- playlists ---

func (s *Store) CreatePlaylist(name string) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO playlists(name) VALUES(?)`, name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) RenamePlaylist(id int64, name string) error {
	_, err := s.db.Exec(`UPDATE playlists SET name=? WHERE id=?`, name, id)
	return err
}

func (s *Store) DeletePlaylist(id int64) error {
	_, err := s.db.Exec(`DELETE FROM playlists WHERE id=?`, id)
	return err
}

func (s *Store) ListPlaylists() ([]Playlist, error) {
	rows, err := s.db.Query(`SELECT id, name FROM playlists ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Playlist
	for rows.Next() {
		var p Playlist
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetPlaylistTracks(id int64) ([]Track, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.title, t.album_id, al.title, t.artist_id, ar.name,
		       t.track_num, t.duration, t.format, t.file_path, t.mtime
		FROM playlist_tracks pt
		JOIN tracks t ON t.id = pt.track_id
		JOIN albums al ON al.id = t.album_id
		JOIN artists ar ON ar.id = t.artist_id
		WHERE pt.playlist_id=?
		ORDER BY pt.position`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTracks(rows)
}

func (s *Store) AddTrackToPlaylist(playlistID, trackID int64) error {
	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO playlist_tracks(playlist_id, track_id, position)
		VALUES(?, ?, (SELECT COALESCE(MAX(position),0)+1 FROM playlist_tracks WHERE playlist_id=?))`,
		playlistID, trackID, playlistID)
	return err
}

func (s *Store) RemoveTrackFromPlaylist(playlistID, trackID int64) error {
	_, err := s.db.Exec(`DELETE FROM playlist_tracks WHERE playlist_id=? AND track_id=?`, playlistID, trackID)
	return err
}

func (s *Store) ReorderPlaylist(playlistID int64, orderedTrackIDs []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for pos, tid := range orderedTrackIDs {
		if _, err := tx.Exec(
			`UPDATE playlist_tracks SET position=? WHERE playlist_id=? AND track_id=?`,
			pos, playlistID, tid,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// --- helpers ---

type scanner interface {
	Scan(dest ...any) error
}

func scanTrack(row scanner) (Track, error) {
	var t Track
	err := row.Scan(&t.ID, &t.Title, &t.AlbumID, &t.Album, &t.ArtistID, &t.Artist,
		&t.TrackNum, &t.Duration, &t.Format, &t.FilePath, &t.Mtime)
	return t, err
}

func scanTracks(rows *sql.Rows) ([]Track, error) {
	var out []Track
	for rows.Next() {
		t, err := scanTrack(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
