package provider

import "testing"

func TestRelativeName(t *testing.T) {
	cases := []struct{ api, domain, want string }{
		{"example.com", "example.com", ""},
		{"example.com.", "example.com", ""},
		{"www.example.com", "example.com", "www"},
		{"*.dev.example.com", "example.com", "*.dev"},
		{"_dmarc.example.com", "example.com", "_dmarc"},
		{"www", "example.com", "www"},
		{"status-service.dev.example.com", "example.com", "status-service.dev"},
	}
	for _, c := range cases {
		if got := relativeName(c.api, c.domain); got != c.want {
			t.Errorf("relativeName(%q, %q) = %q, want %q", c.api, c.domain, got, c.want)
		}
	}
}
