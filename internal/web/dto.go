package web

import "choinhaccli/internal/library"

type TracksResponse struct {
	Tracks []library.Track `json:"tracks"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

type SearchResponse struct {
	Tracks  []library.Track   `json:"tracks"`
	Albums  []library.Album   `json:"albums"`
	Artists []library.Artist  `json:"artists"`
}

type ScanStatus struct {
	Running  bool `json:"running"`
	Scanned  int  `json:"scanned"`
	Total    int  `json:"total"`
	Done     bool `json:"done"`
}
