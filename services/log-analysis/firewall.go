package main

import (
	"net/netip"
	"strconv"
	"strings"

	"lite-nas/shared/contracts/networksecurity"
)

// parseFirewallAttempt extracts one blocked inbound TCP attempt from a trusted
// kernel record. Other UFW actions and protocols are intentionally ignored.
func parseFirewallAttempt(record logRecord) (networksecurity.FirewallAttempt, bool) {
	if record.Format != formatKernel || !record.KernelTrusted || !strings.HasPrefix(record.Message, "[UFW BLOCK] ") {
		return networksecurity.FirewallAttempt{}, false
	}
	fields := make(map[string]string)
	for _, token := range strings.Fields(record.Message) {
		key, value, ok := strings.Cut(token, "=")
		if ok {
			fields[key] = value
		}
	}
	if fields["IN"] == "" || fields["PROTO"] != "TCP" {
		return networksecurity.FirewallAttempt{}, false
	}
	source, sourceErr := netip.ParseAddr(fields["SRC"])
	destination, destinationErr := netip.ParseAddr(fields["DST"])
	port, portErr := strconv.ParseUint(fields["DPT"], 10, 16)
	if sourceErr != nil || destinationErr != nil || portErr != nil || port == 0 ||
		source.IsUnspecified() || source.IsMulticast() || destination.IsUnspecified() || destination.IsMulticast() {
		return networksecurity.FirewallAttempt{}, false
	}
	return networksecurity.FirewallAttempt{
		Timestamp: record.At, SourceIP: source.Unmap().String(), DestinationIP: destination.Unmap().String(),
		DestinationPort: uint16(port), Protocol: "TCP",
	}, true
}
