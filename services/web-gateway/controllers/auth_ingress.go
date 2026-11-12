package controllers

import (
	"net"
	"net/netip"
	"path/filepath"
	"strings"
)

// verifiedDirectClientIP accepts Nginx's overwritten peer address only when
// HTTPS reached an address assigned to one physical host interface. The
// loopback Tunnel and virtual Tailscale listener addresses are excluded.
func verifiedDirectClientIP(ingress, realIP, localIP string) string {
	if ingress != "direct-https" {
		return ""
	}
	peer, peerErr := netip.ParseAddr(realIP)
	local, localErr := netip.ParseAddr(localIP)
	if peerErr != nil || localErr != nil || !peer.IsGlobalUnicast() || peer.Zone() != "" || peer.Unmap().String() != realIP || !local.IsGlobalUnicast() || local.Zone() != "" || local.Unmap().String() != localIP {
		return ""
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range interfaces {
		device, err := filepath.EvalSymlinks(filepath.Join("/sys/class/net", iface.Name, "device"))
		if err != nil || !strings.HasPrefix(device, "/sys/devices/") || strings.HasPrefix(device, "/sys/devices/virtual/") {
			continue
		}
		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, address := range addresses {
			prefix, err := netip.ParsePrefix(address.String())
			if err == nil && prefix.Addr().Unmap() == local.Unmap() {
				return realIP
			}
		}
	}
	return ""
}
