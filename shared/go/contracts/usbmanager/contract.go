// Package usbmanager defines authenticated USBGuard device management RPCs.
package usbmanager

const (
	ListSubject  = "usb-manager.rpc.list"
	AllowSubject = "usb-manager.rpc.allow"
	BlockSubject = "usb-manager.rpc.block"
)

// ListRequest authenticates a request for the current USBGuard device list.
type ListRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
}

// DeviceRequest identifies a USBGuard device and optional permanent policy update.
type DeviceRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	DeviceID    string `json:"device_id" validate:"required,max=10"`
	Permanent   bool   `json:"permanent"`
}

// ListResponse contains the usbguard CLI device inventory.
type ListResponse struct {
	Devices string `json:"devices"`
}

// ActionResponse confirms that USBGuard completed the requested action.
type ActionResponse struct {
	Completed bool `json:"completed"`
}
