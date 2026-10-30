package config

import (
	"net/url"
	"testing"
)

func TestProviders(t *testing.T) {
	providers := Providers()
	if len(providers) == 0 {
		t.Fatal("no providers configured")
	}

	names := make(map[string]bool)
	for _, p := range providers {
		if p.Name == "" || names[p.Name] {
			t.Errorf("provider name %q is empty or duplicated", p.Name)
		}
		names[p.Name] = true

		if p.FullText && len(p.Hosts) == 0 {
			t.Errorf("%s: FullText providers need Hosts to match article URLs", p.Name)
		}
		if len(p.Feeds) == 0 {
			t.Errorf("%s: no feeds", p.Name)
		}

		feedNames := make(map[string]bool)
		for _, f := range p.Feeds {
			if f.Name == "" || feedNames[f.Name] {
				t.Errorf("%s: feed name %q is empty or duplicated", p.Name, f.Name)
			}
			feedNames[f.Name] = true

			u, err := url.Parse(f.URL)
			if err != nil || u.Scheme != "https" || u.Host == "" {
				t.Errorf("%s/%s: feed URL %q is not a valid https URL", p.Name, f.Name, f.URL)
			}
		}
	}
}

func TestProvidersReturnsCopy(t *testing.T) {
	Providers()[0].Name = "changed"
	if Providers()[0].Name == "changed" {
		t.Error("Providers() exposes the package-level slice")
	}
}
