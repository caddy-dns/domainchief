// Package domainchief provides a Caddy DNS provider module for Domain Chief hosted DNS.
package domainchief

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/libdns/domainchief"
)

// Provider wraps the Domain Chief libdns provider as a Caddy module.
type Provider struct{ *domainchief.Provider }

func init() {
	caddy.RegisterModule(Provider{})
}

// CaddyModule returns the Caddy module information.
func (Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dns.providers.domainchief",
		New: func() caddy.Module { return &Provider{new(domainchief.Provider)} },
	}
}

// Provision resolves placeholders in the configuration. Implements caddy.Provisioner.
//
// It also makes publishing ACME challenges wait until every nameserver serves them. Caddy only checks the
// nameservers it can reach, which can serve a value before the certificate authority's nameservers do.
func (p *Provider) Provision(_ caddy.Context) error {
	replacer := caddy.NewReplacer()

	p.WaitForPublication = true

	p.APIToken = replacer.ReplaceAll(p.APIToken, "")
	p.BaseURL = replacer.ReplaceAll(p.BaseURL, "")

	if p.APIToken == "" {
		return errors.New("domainchief: missing API token, make sure the environment variable holding it is set")
	}

	return nil
}

// UnmarshalCaddyfile sets up the DNS provider from Caddyfile tokens. Syntax:
//
//	domainchief [<api_token>] {
//	    api_token <api_token>
//	    base_url  <base_url>
//	}
//
// Expansion of placeholders is left to Provision.
func (p *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive name

	if d.NextArg() {
		p.APIToken = d.Val()
	}

	if d.NextArg() {
		return d.ArgErr()
	}

	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "api_token":
			if p.APIToken != "" {
				return d.Err("API token already set")
			}

			if !d.NextArg() {
				return d.ArgErr()
			}

			p.APIToken = d.Val()
		case "base_url":
			if !d.NextArg() {
				return d.ArgErr()
			}

			p.BaseURL = d.Val()
		default:
			return d.Errf("unrecognized subdirective '%s'", d.Val())
		}

		if d.NextArg() {
			return d.ArgErr()
		}
	}

	if p.APIToken == "" {
		return d.Err("missing API token")
	}

	return nil
}

// Interface guards
var (
	_ caddyfile.Unmarshaler = (*Provider)(nil)
	_ caddy.Provisioner     = (*Provider)(nil)
)
