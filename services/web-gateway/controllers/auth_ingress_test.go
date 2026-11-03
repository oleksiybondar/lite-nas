package controllers

import "testing"

// TestTunnelIngressNeverSuppliesSessionClientIP ensures Cloudflare's loopback
// origin cannot turn an unverified browser header into mitigation evidence.
func TestTunnelIngressNeverSuppliesSessionClientIP(t *testing.T) {
	if got := verifiedDirectClientIP("tunnel", "198.51.100.10", "192.0.2.10"); got != "" {
		t.Fatalf("tunnel client IP = %q", got)
	}
}

// TestLoopbackListenerCannotSupplyDirectClientIP excludes local HTTPS and
// virtual recovery interfaces from the physical-ingress evidence path.
func TestLoopbackListenerCannotSupplyDirectClientIP(t *testing.T) {
	if got := verifiedDirectClientIP("direct-https", "198.51.100.10", "127.0.0.1"); got != "" {
		t.Fatalf("loopback listener client IP = %q", got)
	}
}
