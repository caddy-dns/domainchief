package domainchief

import (
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/libdns/domainchief"
)

func TestUnmarshalCaddyfile(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		token   string
		baseURL string
		wantErr bool
	}{
		{name: "inline token", input: `domainchief ctt_token`, token: "ctt_token"},
		{name: "block token", input: "domainchief {\n api_token ctt_token\n}", token: "ctt_token"},
		{name: "base url", input: "domainchief ctt_token {\n base_url https://example.test/api/v1\n}", token: "ctt_token", baseURL: "https://example.test/api/v1"},
		{name: "placeholder", input: `domainchief {env.DOMAINCHIEF_API_TOKEN}`, token: "{env.DOMAINCHIEF_API_TOKEN}"},
		{name: "missing token", input: `domainchief`, wantErr: true},
		{name: "token set twice", input: "domainchief ctt_token {\n api_token ctt_other\n}", wantErr: true},
		{name: "extra argument", input: `domainchief ctt_token extra`, wantErr: true},
		{name: "unknown subdirective", input: "domainchief {\n api_key ctt_token\n}", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := Provider{new(domainchief.Provider)}

			err := provider.UnmarshalCaddyfile(caddyfile.NewTestDispenser(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}

				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if provider.APIToken != tt.token || provider.BaseURL != tt.baseURL {
				t.Errorf("unexpected config token=%q base_url=%q", provider.APIToken, provider.BaseURL)
			}
		})
	}
}

func TestProvisionResolvesEnvironmentPlaceholders(t *testing.T) {
	t.Setenv("DOMAINCHIEF_API_TOKEN", "ctt_from_env")

	provider := Provider{&domainchief.Provider{APIToken: "{env.DOMAINCHIEF_API_TOKEN}"}}

	if err := provider.Provision(caddy.Context{}); err != nil {
		t.Fatal(err)
	}

	if provider.APIToken != "ctt_from_env" {
		t.Errorf("expected the placeholder to be replaced, got %q", provider.APIToken)
	}
}

func TestProvisionFailsWhenTheTokenResolvesToEmpty(t *testing.T) {
	provider := Provider{&domainchief.Provider{APIToken: "{env.DOMAINCHIEF_UNSET_TOKEN}"}}

	if err := provider.Provision(caddy.Context{}); err == nil {
		t.Fatal("expected an error")
	}
}
