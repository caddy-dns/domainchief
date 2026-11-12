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
		propagation_delay 60s
	}
}
```

We recommend setting [`propagation_delay`](https://caddyserver.com/docs/caddyfile/directives/tls#propagation_delay). Domain Chief serves hosted DNS from nameservers in many locations, and a new challenge record reaches all of them shortly after it reaches the ones closest to you. Caddy only checks the nameservers it can reach before asking the certificate authority to validate, so without a delay the certificate authority can still query a location that does not have the record yet.

You can also set the provider for all sites. The global option does not accept `propagation_delay`, so prefer the `tls` directive above:

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
			},
			"propagation_delay": "60s"
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

### `timed out waiting for record to fully propagate`

Domain Chief publishes challenge records to the hosted DNS nameservers asynchronously, which usually takes less than a minute. Caddy waits up to 2 minutes by default. Raise the limit with the [`propagation_timeout`](https://caddyserver.com/docs/caddyfile/directives/tls#propagation_timeout) option if validation times out:

```caddyfile
tls {
	dns domainchief {env.DOMAINCHIEF_API_TOKEN}
	propagation_timeout 5m
}
```

### `NXDOMAIN looking up TXT for _acme-challenge…` or `No TXT record found`

The certificate authority queried a nameserver location that did not have the challenge record yet. Set or raise [`propagation_delay`](https://caddyserver.com/docs/caddyfile/directives/tls#propagation_delay) so Caddy waits before validation starts:

```caddyfile
tls {
	dns domainchief {env.DOMAINCHIEF_API_TOKEN}
	propagation_delay 2m
}
```

The certificate authority also caches a missing record for up to 5 minutes, so wait at least that long before retrying a failed validation for the same name.

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
