package tunnel

import (
	"strings"
	"testing"
)

func TestCloudflareConfigYAML(t *testing.T) {
	yml := CloudflareConfigYAML(CloudflareOptions{
		Hostname:   "memory.example.com",
		URL:        "http://127.0.0.1:7432",
		TunnelName: "fortmemory",
	})
	for _, want := range []string{
		"memory.example.com",
		"http://127.0.0.1:7432",
		"tunnel:",
		"ingress:",
	} {
		if !strings.Contains(yml, want) {
			t.Fatalf("missing %q in:\n%s", want, yml)
		}
	}
}

func TestDetectCloudflaredNonEmptyOrEmpty(t *testing.T) {
	// Smoke: should not panic whether installed or not.
	_ = DetectCloudflared()
}
