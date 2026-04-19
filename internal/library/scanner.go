package library

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
)

type Progress struct {
	Scanned int
	Total   int
	Done    bool
	Err     error
}

var audioExts = map[string]string{
	".flac": "flac",
	".wav":  "wav",
	".mp3":  "mp3",
}

// ScanAsync walks root and upserts tracks into store. Returns a progress channel.
// coverDir is the directory where extracted cover art is cached; pass "" to skip extraction.
func ScanAsync(root string, store *Store, coverDir string) <-chan Progress {
	ch := make(chan Progress, 32)
	go func() {
		defer close(ch)

		var files []string
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Printf("scan: skipping %s: %v", path, err)
				if d != nil && d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			if d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if _, ok := audioExts[ext]; ok {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			ch <- Progress{Err: err, Done: true}
			return
		}

		total := len(files)
		for i, path := range files {
			if err := scanFile(path, store, coverDir); err != nil {
				log.Printf("scan skip %s: %v", path, err)
			}
			ch <- Progress{Scanned: i + 1, Total: total}
		}
		ch <- Progress{Scanned: total, Total: total, Done: true}
	}()
	return ch
}

func scanFile(path string, store *Store, coverDir string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	mtime := info.ModTime().Unix()

	stored, err := store.GetTrackMtime(path)
	if err != nil {
		return err
	}
	if stored == mtime {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	format := audioExts[ext]

	meta, err := tag.ReadFrom(f)

	title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	artist := "Unknown Artist"
	album := "Unknown Album"
	trackNum := 0
	duration := 0

	if err == nil {
		if v := meta.Title(); v != "" {
			title = v
		}
		if v := meta.Artist(); v != "" {
			artist = v
		}
		if v := meta.Album(); v != "" {
			album = v
		}
		n, _ := meta.Track()
		trackNum = n
	}

	if err := store.UpsertTrack(Track{
		Title:    title,
		Album:    album,
		Artist:   artist,
		TrackNum: trackNum,
		Duration: duration,
		Format:   format,
		FilePath: path,
		Mtime:    mtime,
	}); err != nil {
		return err
	}

	// Extract cover art once per album (skipped if coverDir is empty)
	if coverDir != "" {
		albumID, err := store.GetAlbumIDForTrack(path)
		if err == nil && albumID > 0 {
			if coverPath := ExtractCover(albumID, path, coverDir); coverPath != "" {
				_ = store.UpdateAlbumCover(albumID, coverPath)
			}
		}
	}

	return nil
}
