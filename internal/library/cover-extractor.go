package library

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dhowden/tag"
)

var coverFileNames = []string{
	"cover.jpg", "cover.png",
	"folder.jpg", "folder.png",
	"album.jpg", "album.png",
}

// ExtractCover finds cover art for albumID using trackPath as the first track hint.
// Tries embedded art first, then image files in the track's directory.
// Saves to coverDir/{albumID}.jpg and returns the path, or "" if none found.
func ExtractCover(albumID int64, trackPath, coverDir string) string {
	destPath := filepath.Join(coverDir, fmt.Sprintf("%d.jpg", albumID))

	if _, err := os.Stat(destPath); err == nil {
		return destPath // already extracted
	}

	if err := os.MkdirAll(coverDir, 0755); err != nil {
		return ""
	}

	if data := embeddedArtData(trackPath); len(data) > 0 {
		if err := os.WriteFile(destPath, data, 0644); err == nil {
			return destPath
		}
	}

	dir := filepath.Dir(trackPath)
	for _, name := range coverFileNames {
		if data, err := os.ReadFile(filepath.Join(dir, name)); err == nil {
			if err := os.WriteFile(destPath, data, 0644); err == nil {
				return destPath
			}
		}
	}

	return ""
}

func embeddedArtData(trackPath string) []byte {
	f, err := os.Open(trackPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	meta, err := tag.ReadFrom(f)
	if err != nil {
		return nil
	}
	pic := meta.Picture()
	if pic == nil {
		return nil
	}
	return pic.Data
}
