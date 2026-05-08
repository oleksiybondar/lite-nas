package routes

import (
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"

	"github.com/danielgtaylor/huma/v2"
)

// mountAuthRouter registers the documented auth routes on the shared API.
func mountAuthRouter(
	api huma.API,
	controllerModule modules.Controllers,
	authentication middlewares.AuthenticationOptions,
) {
	api.UseMiddleware(middlewares.ExtractAuthentication(authentication))

	authGroup := huma.NewGroup(api, "/auth")
	authGroup.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = []string{"auth"}
	})

	controller := controllerModule.Auth
	huma.Post(authGroup, "/login", controller.Login)
	huma.Post(authGroup, "/logout", controller.Logout)
	huma.Post(authGroup, "/refresh", controller.Refresh)

	protected := huma.NewGroup(authGroup)
	protected.UseMiddleware(middlewares.RequireAuthentication(api, authentication))
	huma.Get(protected, "/me", controller.Me)
}
