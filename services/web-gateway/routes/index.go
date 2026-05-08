package routes

import (
	"lite-nas/services/web-gateway/modules"

	"github.com/go-chi/chi/v5"
)

// mountIndexRouter mounts the browser entrypoint and SPA fallback routes on the
// root router.
func mountIndexRouter(root chi.Router, controllerModule modules.Controllers) {
	controller := controllerModule.Static
	root.Get("/", controller.ServeIndex)
	root.Get("/favicon.ico", controller.ServeFavicon)
	root.Get("/*", controller.ServeIndex)
}
