package routes

import (
	"net/http"

	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"
	"lite-nas/shared/apireference"
	"lite-nas/shared/browserauth"

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
	mountManagementRouters(api, controllerModule, authentication)
	mountBackupRestoreRouter(api, controllerModule, authentication)
	mountMetricsRouters(api, controllerModule, authentication)
	root.Mount("/api", apiRouter)
	mountIndexRouter(root, controllerModule)

	return browserauth.SameOrigin(root)
}

func apiConfig(serviceName string, version string) huma.Config {
	config := huma.DefaultConfig(serviceName, version)
	config.Servers = []*huma.Server{{URL: "/api"}}
	config.DocsPath = ""
	return config
}

func mountDocsRoute(apiRouter chi.Router, config huma.Config) {
	apiRouter.Get("/docs", apireference.Handler(docsTitle(config), docsOpenAPIPath(config)))
}

func docsTitle(config huma.Config) string {
	if config.Info != nil && config.Info.Title != "" {
		return config.Info.Title + " Reference"
	}
	return "API Reference"
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
	mountProcessMetricsRouter(api, controllerModule, authentication)
	mountServiceMetricsRouter(api, controllerModule, authentication)
	mountDiskMetricsRouter(api, controllerModule, authentication)
	mountNetworkMetricsRouter(api, controllerModule, authentication)
	mountZFSMetricsRouter(api, controllerModule, authentication)
}
