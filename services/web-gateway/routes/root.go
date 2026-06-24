package routes

import (
	"fmt"
	"net/http"

	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates the browser-facing root router for the web gateway and
// mounts the route slices owned by each gateway area.
//
// Parameters:
//   - serviceName: API title exposed through the generated Huma documents
//   - version: API version exposed through the generated Huma documents
//   - controllerModule: controller dependencies mounted into route slices
//   - authentication: authentication middleware configuration
func NewRouter(
	serviceName string,
	version string,
	controllerModule modules.Controllers,
	authentication middlewares.AuthenticationOptions,
) http.Handler {
	root := chi.NewMux()
	useRootMiddlewares(root)

	apiRouter := chi.NewMux()
	config := apiConfig(serviceName, version)
	api := humachi.New(apiRouter, config)

	mountAssetsRouter(root, controllerModule)
	mountDocsRoute(apiRouter, config)
	mountAuthRouter(api, controllerModule, authentication)
	mountAlertsRouters(api, controllerModule, authentication)
	mountMetricsRouters(api, controllerModule, authentication)
	root.Mount("/api", apiRouter)
	mountIndexRouter(root, controllerModule)

	return root
}

func apiConfig(serviceName string, version string) huma.Config {
	config := huma.DefaultConfig(serviceName, version)
	config.Servers = []*huma.Server{{URL: "/api"}}
	config.DocsPath = ""
	return config
}

func mountDocsRoute(apiRouter chi.Router, config huma.Config) {
	apiRouter.Get("/docs", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write([]byte(buildDocsHTML(config)))
	})
}

const docsHTMLTemplate = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="referrer" content="same-origin" />
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no" />
    <title>%s</title>
    <link href="https://unpkg.com/@stoplight/elements@9.0.0/styles.min.css" rel="stylesheet" />
    <script src="https://unpkg.com/@stoplight/elements@9.0.0/web-components.min.js" integrity="sha256-Tqvw1qE2abI+G6dPQBc5zbeHqfVwGoamETU3/TSpUw4=" crossorigin="anonymous"></script>
    <style>
      html, body {
        height: 100%%;
        margin: 0;
      }

      body,
      elements-api {
        min-height: 100vh;
      }
    </style>
  </head>
  <body>
    <elements-api
      apiDescriptionUrl="%s"
      router="hash"
      layout="sidebar"
      tryItCredentialsPolicy="same-origin"
    ></elements-api>
  </body>
</html>`

func buildDocsHTML(config huma.Config) string {
	return fmt.Sprintf(docsHTMLTemplate, docsTitle(config), docsOpenAPIPath(config))
}

func docsTitle(config huma.Config) string {
	if config.Info != nil && config.Info.Title != "" {
		return config.Info.Title + " Reference"
	}
	return "Elements in HTML"
}

func docsOpenAPIPath(config huma.Config) string {
	if len(config.Servers) == 0 || config.Servers[0] == nil || config.Servers[0].URL == "" {
		return "/api/openapi.yaml"
	}
	return config.Servers[0].URL + config.OpenAPIPath + ".yaml"
}

func useRootMiddlewares(root chi.Router) {
	root.Use(chimiddleware.RequestID)
	root.Use(chimiddleware.RealIP)
	root.Use(chimiddleware.Recoverer)
}

func mountMetricsRouters(
	api huma.API,
	controllerModule modules.Controllers,
	authentication middlewares.AuthenticationOptions,
) {
	mountSystemMetricsRouter(api, controllerModule, authentication)
	mountServiceMetricsRouter(api, controllerModule, authentication)
	mountDiskMetricsRouter(api, controllerModule, authentication)
	mountNetworkMetricsRouter(api, controllerModule, authentication)
	mountZFSMetricsRouter(api, controllerModule, authentication)
}
