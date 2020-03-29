package admin

import (
	"net/http"
	"strings"
)

// extensions is the set of known extensions that we support.
var extensions = map[string]string{
	"apple-app-site-association": "application/json",
	".svg":                       "image/svg+xml",
	".ttf":                       "application/x-font-truetype",
	".otf":                       "application/x-font-opentype",
	".woff":                      "application/font-woff",
	".woff2":                     "application/font-woff2",
	".eot":                       "application/vnd.ms-fontobject",
	".sfnt":                      "application/font-sfnt",
	".jpg":                       "image/jpeg",
	".jpeg":                      "image/jpeg",
	".gif":                       "image/gif",
	".png":                       "image/png",
	".js":                        "application/javascript",
	".css":                       "text/css",
	".html":                      "text/html",
	".ico":                       "image/x-icon",
	".json":                      "application/json",
	".zip":                       "application/zip",
	".py":                        "text/x-python",
}

// detectTypeFromExtension detects the type of file based on its extension.
// This is sufficient for our needs here since we control all files going in, basically
// all of which will have known extensions. Consider carefully whether that applies if you
// are using it outside this package.
// The empty string is returned if no extension is known.
func detectTypeFromExtension(filename string) string {
	for extension, mimetype := range extensions {
		if strings.HasSuffix(filename, extension) {
			return mimetype
		}
	}
	log.Warningf("No MIME type detected for filename %s", filename)
	return ""
}

// writeContentType writes a Content-Type header into the given writer.
func writeContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}
