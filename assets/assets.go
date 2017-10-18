package assets

import (
	"path"
)

//go:generate go-bindata -ignore .go -o assets.gen.go -pkg assets ./...

var FormatVersions = []string{"20150701", "20150702"}

// Format returns the contents of the named MAR file.
// If the verison is not specified then latest version is returned.
// Returns nil if the format does not exist.
func Format(name, version string) []byte {
	// Return specific version, if specified.
	if version != "" {
		buf, _ := Asset(path.Join("formats", version, name+".mar"))
		return buf
	}

	// Otherwise iterate over versions from newest to oldest.
	for i := len(FormatVersions) - 1; i >= 0; i-- {
		if buf, _ := Asset(path.Join("formats", FormatVersions[i], name+".mar")); buf != nil {
			return buf
		}
	}

	return nil
}
