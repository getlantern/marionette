package mar

import (
	"path"
	"sort"
	"strings"
)

//go:generate go-bindata -ignore .go -o mar.gen.go -pkg mar ./...

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

// Formats returns a list of available formats.
func Formats() []string {
	var formats []string

	names := AssetNames()
	sort.Strings(names)

	for _, name := range names {
		// Ignore files outside 'formats' directory.
		if !strings.HasPrefix(name, "formats/") {
			continue
		}

		// Remove subdir and extension.
		name = strings.TrimPrefix(name, "formats/")
		name = strings.TrimSuffix(name, ".mar")

		// Move version to the end.
		segments := strings.SplitN(name, "/", 2)
		format := segments[1] + ":" + segments[0]

		// Add to format list.
		formats = append(formats, format)
	}
	return formats
}
