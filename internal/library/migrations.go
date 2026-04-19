package library

import "database/sql"

const schema = `
CREATE TABLE IF NOT EXISTS artists (
	id   INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS albums (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	title      TEXT NOT NULL,
	artist_id  INTEGER NOT NULL REFERENCES artists(id),
	year       INTEGER DEFAULT 0,
	cover_path TEXT DEFAULT '',
	UNIQUE(title, artist_id)
);

CREATE TABLE IF NOT EXISTS tracks (
	id        INTEGER PRIMARY KEY AUTOINCREMENT,
	title     TEXT NOT NULL,
	album_id  INTEGER NOT NULL REFERENCES albums(id),
	artist_id INTEGER NOT NULL REFERENCES artists(id),
	track_num INTEGER DEFAULT 0,
	duration  INTEGER DEFAULT 0,
	format    TEXT NOT NULL,
	file_path TEXT NOT NULL UNIQUE,
	mtime     INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS playlists (
	id   INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS playlist_tracks (
	playlist_id INTEGER NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
	track_id    INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
	position    INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (playlist_id, track_id)
);

CREATE INDEX IF NOT EXISTS idx_tracks_album   ON tracks(album_id);
CREATE INDEX IF NOT EXISTS idx_tracks_artist  ON tracks(artist_id);
CREATE INDEX IF NOT EXISTS idx_tracks_title   ON tracks(title);
CREATE INDEX IF NOT EXISTS idx_albums_artist  ON albums(artist_id);
`

func migrate(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}
