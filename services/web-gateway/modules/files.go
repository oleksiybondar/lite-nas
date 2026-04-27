package modules

import (
	"path/filepath"

	"lite-nas/services/web-gateway/controllers"
	sharedfileio "lite-nas/shared/fileio"
)

const (
	indexHTMLName  = "index.html"
	indexCSSName   = "index.css"
	indexJSName    = "index.js"
	faviconICOName = "favicon.ico"
)

// Files groups packaged file readers owned by the gateway runtime.
//
// The fields are populated once during startup and are expected to be treated
// as logically read-only after construction.
type Files struct {
	Static controllers.StaticFiles
}

// NewFilesModule opens the frontend assets served by the gateway.
//
// Parameters:
//   - assetRoot: directory containing the packaged or development frontend assets
//
// It resolves each reader from the selected asset directory so the static
// controller can serve explicit files without depending on directory-backed
// static file serving.
func NewFilesModule(assetRoot string) (Files, error) {
	indexHTML, err := sharedfileio.NewFileReader(filepath.Join(assetRoot, indexHTMLName))
	if err != nil {
		return Files{}, err
	}

	indexCSS, err := sharedfileio.NewFileReader(filepath.Join(assetRoot, indexCSSName))
	if err != nil {
		return Files{}, err
	}

	indexJS, err := sharedfileio.NewFileReader(filepath.Join(assetRoot, indexJSName))
	if err != nil {
		return Files{}, err
	}

	favicon, err := sharedfileio.NewFileReader(filepath.Join(assetRoot, faviconICOName))
	if err != nil {
		return Files{}, err
	}

	return Files{
		Static: controllers.StaticFiles{
			IndexHTML: indexHTML,
			IndexCSS:  indexCSS,
			IndexJS:   indexJS,
			Favicon:   favicon,
		},
	}, nil
}
