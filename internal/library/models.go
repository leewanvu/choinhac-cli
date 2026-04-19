package library

type Artist struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Album struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	ArtistID int64  `json:"artist_id"`
	Artist   string `json:"artist"`
	Year     int    `json:"year"`
	CoverPath string `json:"cover_path,omitempty"`
}

type Playlist struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Track struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	AlbumID   int64  `json:"album_id"`
	Album     string `json:"album"`
	ArtistID  int64  `json:"artist_id"`
	Artist    string `json:"artist"`
	TrackNum  int    `json:"track_num"`
	Duration  int    `json:"duration"` // seconds
	Format    string `json:"format"`   // flac, wav, mp3
	FilePath  string `json:"file_path"`
	Mtime     int64  `json:"mtime"`
}
