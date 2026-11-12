package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"
)

// mountJournalRouter adds security-gated journal browsing and vacuum routes.
func mountJournalRouter(api huma.API, controllers modules.Controllers, authentication middlewares.AuthenticationOptions) {
	group := huma.NewGroup(api, "/journal")
	group.UseSimpleModifier(func(op *huma.Operation) { op.Tags = []string{"journal-management"} })
	group.UseMiddleware(middlewares.RequireAuthentication(api, authentication))
	group.UseMiddleware(middlewares.RequireSecurity(api))
	huma.Register(group, managementOperation("journal-query", http.MethodGet, "", "Query host journal", "Return one newest-first journal chunk."), controllers.Journal.Query)
	huma.Register(group, managementOperation("journal-units", http.MethodGet, "/units", "List matching journal units", "Return service units with entries matching the non-unit filters."), controllers.Journal.Units)
	huma.Register(group, managementOperation("journal-vacuum", http.MethodPost, "/vacuum", "Vacuum host journal", "Rotate active files and retain the selected number of days."), controllers.Journal.Vacuum)
}
