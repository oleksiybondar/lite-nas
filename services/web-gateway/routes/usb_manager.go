package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"
)

// mountUSBManagerRouter registers authenticated USBGuard management endpoints.
func mountUSBManagerRouter(api huma.API, controllers modules.Controllers, authentication middlewares.AuthenticationOptions) {
	group := huma.NewGroup(api, "/usbguard")
	group.UseSimpleModifier(func(op *huma.Operation) { op.Tags = []string{"usbguard-management"} })
	group.UseMiddleware(middlewares.RequireAuthentication(api, authentication))
	group.UseMiddleware(middlewares.RequireSecurity(api))
	manager := controllers.USBManager
	huma.Register(group, managementOperation("usbguard-devices-list", http.MethodGet, "/devices", "List USBGuard devices", "Returns the current USBGuard device inventory. Requires a security-equivalent access token."), manager.List)
	huma.Register(group, managementOperation("usbguard-device-allow", http.MethodPost, "/devices/{device_id}/allow", "Allow USBGuard device", "Allows the specified USBGuard device. Set permanent to persist the decision in policy. Requires a security-equivalent access token."), manager.Allow)
	huma.Register(group, managementOperation("usbguard-device-block", http.MethodPost, "/devices/{device_id}/block", "Block USBGuard device", "Blocks the specified USBGuard device. Set permanent to persist the decision in policy. Requires a security-equivalent access token."), manager.Block)
}
