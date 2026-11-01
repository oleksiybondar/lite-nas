// Package networksecurity defines the firewall observation event consumed by
// Network Security. Log Analysis publishes only validated kernel UFW blocks.
package networksecurity

import "time"

// FirewallAttemptSubject carries one blocked inbound TCP connection attempt.
const FirewallAttemptSubject = "network.security.events.firewall_attempt"

// FirewallAttempt is factual evidence, not a security verdict. Source and
// destination addresses are canonical IP strings; DestinationPort is nonzero.
type FirewallAttempt struct {
	Timestamp       time.Time `json:"timestamp" validate:"required"`
	SourceIP        string    `json:"source_ip" validate:"required,ip"`
	DestinationIP   string    `json:"destination_ip" validate:"required,ip"`
	DestinationPort uint16    `json:"destination_port" validate:"required,gt=0"`
	Protocol        string    `json:"protocol" validate:"required,oneof=TCP"`
}
