Domain Chief module for Caddy
===========================

This package contains a DNS provider module for [Caddy](https://github.com/caddyserver/caddy). It can be used to solve ACME DNS-01 challenges and manage DNS records for domains that use [Domain Chief](https://domain.chief.app?ref=caddy-dns-domainchief) hosted DNS.

## Caddy module name

```
dns.providers.domainchief
```

## Installation

Download a Caddy binary that includes this module from the [Caddy download page](https://caddyserver.com/download?package=github.com%2Fcaddy-dns%2Fdomainchief), or build one with [`xcaddy`](https://github.com/caddyserver/xcaddy):

```bash
xcaddy build --with github.com/caddy-dns/domainchief
```

## API token

[Create a team access token](https://account.chief.app/team/select/token/create?scopes=domainchief:acme:write&ref=caddy-dns-domainchief) with only the `domainchief:acme:write` scope, for the team that owns your domains. The domains must use Domain Chief hosted DNS.

The `domainchief:acme:write` scope only allows publishing and deleting TXT records on `_acme-challenge` names. It cannot read your domains or change any other DNS record. Anyone holding the token can still obtain certificates for the domains in your team, so store it like any other secret.

Other Caddy plugins that use the DNS provider to manage records, such as [dynamic DNS](https://github.com/mholt/caddy-dynamicdns), need a token with the `domainchief:dns:write` scope instead.

## Configuration

### Caddyfile

Use the provider for a site:

```caddyfile
example.com, *.example.com {
	tls {
		dns domainchief {env.DOMAINCHIEF_API_TOKEN}
	}
}
```

Domain Chief serves hosted DNS from nameservers in many locations, and a new challenge record reaches the ones closest to you before it reaches the others. Caddy only checks the nameservers it can reach, so the module waits until Domain Chief reports that every nameserver serves the challenge before Caddy asks the certificate authority to validate. This usually takes one to three minutes per certificate, so you don't need to set [`propagation_delay`](https://caddyserver.com/docs/caddyfile/directives/tls#propagation_delay).

You can also set the provider for all sites:

```caddyfile
{
	acme_dns domainchief {env.DOMAINCHIEF_API_TOKEN}
}
```

The block syntax is also supported:

```caddyfile
dns domainchief {
	api_token {env.DOMAINCHIEF_API_TOKEN}
}
```

You can put the token directly in the config instead of the `{env.DOMAINCHIEF_API_TOKEN}` placeholder, but keeping it in an environment variable is more secure.

### JSON

```json
{
	"module": "acme",
	"challenges": {
		"dns": {
			"provider": {
				"name": "domainchief",
				"api_token": "{env.DOMAINCHIEF_API_TOKEN}"
			}
		}
	}
}
```

## Troubleshooting

### `missing API token`

The environment variable holding the token is not set for the Caddy process. When Caddy runs as a systemd service, add it to the service with `systemctl edit caddy`:

```ini
[Service]
Environment=DOMAINCHIEF_API_TOKEN=ctt_...
```

### `ACME challenge … is not served by every nameserver after 10m0s`

Other changes to the same domain's hosted DNS kept delaying the challenge record. The module deletes the record and Caddy retries the certificate later. Avoid changing the domain's DNS records while certificates are being issued.

### `ACME challenge … could not be added to the hosted DNS zone`

Domain Chief could not publish the challenge record. The module deletes the record and Caddy retries the certificate later.

### `dns.not_hosted`

The domain does not use Domain Chief hosted DNS, so the challenge record cannot be published.

## Testing

The unit tests cover the configuration parsing:

```bash
go test ./...
```

The end-to-end test builds Caddy with this module using [`xcaddy`](https://github.com/caddyserver/xcaddy) and obtains two certificates from Let's Encrypt staging, one for a host and one for a wildcard below it. It needs a token with the `domainchief:acme:write` scope and a domain that uses hosted DNS:

```bash
DOMAINCHIEF_API_TOKEN=ctt_... DOMAINCHIEF_TEST_ZONE=example.com ./e2e/run.sh
```

CI runs the end-to-end test on pushes to `master` when the `DOMAINCHIEF_TEST_ZONE` repository variable and `DOMAINCHIEF_API_TOKEN` repository secret are set.

## Details

See the [libdns provider](https://github.com/libdns/domainchief) for how records are managed.
