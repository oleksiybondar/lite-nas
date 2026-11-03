// Package apireference renders the shared Scalar API reference page.
package apireference

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
)

// Handler serves the shared API reference for a service-owned document URL.
func Handler(title, documentURL string) http.HandlerFunc {
	page := Page(title, documentURL)
	return func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write([]byte(page))
	}
}

// Page returns a themed API reference for a service-owned OpenAPI URL.
func Page(title, documentURL string) string {
	urlJSON, _ := json.Marshal(documentURL)
	return fmt.Sprintf(pageTemplate, html.EscapeString(title), urlJSON)
}

const pageTemplate = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="referrer" content="same-origin" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>%s</title>
  </head>
  <body>
    <div id="api-reference"></div>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.76.3"></script>
    <script>
      Scalar.createApiReference('#api-reference', {
        url: %s,
        theme: 'moon',
        darkMode: true,
        withDefaultFonts: false,
        showDeveloperTools: 'never',
        customFetch: (input, init) => fetch(input, { ...init, credentials: 'same-origin' })
      });
    </script>
  </body>
</html>`
