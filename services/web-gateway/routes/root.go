package routes

import (
	"net/http"

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
func NewRouter(
	serviceName string,
	version string,
	controllerModule modules.Controllers,
) http.Handler {
	root := chi.NewMux()
	root.Use(chimiddleware.RequestID)
	root.Use(chimiddleware.RealIP)
	root.Use(chimiddleware.Recoverer)

	api := humachi.New(root, huma.DefaultConfig(serviceName, version))

	mountIndexRouter(root, controllerModule)
	mountAssetsRouter(root, controllerModule)
	mountAuthRouter(api, controllerModule)
	mountSystemMetricsRouter(api, controllerModule)

	return root
}
