package web

import (
	"bytes"
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"m3m/internal/config"
)

//go:embed all:ui/dist
var staticFS embed.FS

// GetFileSystem returns the filesystem for static files
func GetFileSystem() (http.FileSystem, error) {
	subFS, err := fs.Sub(staticFS, "ui/dist")
	if err != nil {
		return nil, err
	}
	return http.FS(subFS), nil
}

// GetIndexHTML returns index.html with injected config
func GetIndexHTML(cfg *config.Config) ([]byte, error) {
	indexBytes, err := staticFS.ReadFile("ui/dist/index.html")
	if err != nil {
		return nil, err
	}

	// Inject configuration into HTML
	configScript := `<script>
window.__APP_CONFIG__ = {
    apiURL: "` + cfg.Server.URI + `"
};
</script>`

	htmlContent := string(indexBytes)
	headEndIndex := strings.Index(htmlContent, "</head>")
	if headEndIndex != -1 {
		var buf bytes.Buffer
		buf.WriteString(htmlContent[:headEndIndex])
		buf.WriteString(configScript)
		buf.WriteString(htmlContent[headEndIndex:])
		return buf.Bytes(), nil
	}

	return indexBytes, nil
}

// HasUI checks if the embedded UI files exist
func HasUI() bool {
	_, err := staticFS.ReadFile("ui/dist/index.html")
	return err == nil
}
