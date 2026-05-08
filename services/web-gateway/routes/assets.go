package routes

import (
	"lite-nas/services/web-gateway/modules"

	"github.com/go-chi/chi/v5"
)

// mountAssetsRouter mounts the packaged static asset routes under /assets.
func mountAssetsRouter(root chi.Router, controllerModule modules.Controllers) {
	assetsRouter := chi.NewRouter()
	controller := controllerModule.Static
	assetsRouter.Get("/index.css", controller.ServeIndexCSS)
	assetsRouter.Get("/index.js", controller.ServeIndexJS)

	root.Mount("/assets", assetsRouter)
}
