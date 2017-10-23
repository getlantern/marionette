package marionette

import (
	"strings"
)

const (
	PartyClient = "client"
	PartyServer = "server"
)

// StripFormatVersion removes any version specified on a format.
func StripFormatVersion(format string) string {
	if i := strings.Index(format, ":"); i != -1 {
		return format[:i]
	}
	return format
}
