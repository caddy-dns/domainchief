#!/usr/bin/env bash
#
# End-to-end test: build Caddy with this module and obtain real certificates from
# Let's Encrypt staging through ACME DNS-01 challenges published on Domain Chief hosted DNS.
#
# Requires go, xcaddy and a token with the domainchief:acme:write scope for a domain that uses hosted DNS:
#
#   DOMAINCHIEF_API_TOKEN=ctt_... DOMAINCHIEF_TEST_ZONE=example.com ./e2e/run.sh
#
set -euo pipefail

cd "$(dirname "$0")/.."

: "${DOMAINCHIEF_API_TOKEN:?DOMAINCHIEF_API_TOKEN must be set}"
: "${DOMAINCHIEF_TEST_ZONE:?DOMAINCHIEF_TEST_ZONE must be set}"

zone="${DOMAINCHIEF_TEST_ZONE%.}"
# A fresh name for every run, so no resolver can answer from an earlier run's cache. The wildcard
# is one level below the host so it does not cover it and both names get their own certificate.
label="caddy-e2e-${GITHUB_RUN_ID:-$(date +%s)}"
host="$label.$zone"
wildcard="*.$label.$zone"
timeout="${E2E_TIMEOUT:-600}"

work="$(mktemp -d)"
export E2E_HOST="$host" E2E_WILDCARD="$wildcard" E2E_STORAGE="$work/storage" E2E_LOG="$work/caddy.log"

echo "Building Caddy with this module"
xcaddy build --output "$work/caddy" --with github.com/caddy-dns/domainchief=.

if ! "$work/caddy" list-modules | grep -qx "dns.providers.domainchief"; then
	echo "The dns.providers.domainchief module is not registered in the Caddy binary."
	exit 1
fi

echo "Obtaining certificates for $host and $wildcard"
"$work/caddy" run --config e2e/Caddyfile --adapter caddyfile &
caddy_pid=$!

stop_caddy() {
	kill "$caddy_pid" 2>/dev/null || true
	wait "$caddy_pid" 2>/dev/null || true
}
trap stop_caddy EXIT

show_log() {
	echo "--- last 40 lines of the Caddy log ---"
	tail -40 "$E2E_LOG" 2>/dev/null || true
}

deadline=$((SECONDS + timeout))

for name in "$host" "$wildcard"; do
	until grep -F '"msg":"certificate obtained successfully"' "$E2E_LOG" 2>/dev/null | grep -qF "\"identifier\":\"$name\""; do
		# Fail on the first failed attempt instead of letting Caddy retry, so slow publishing is not hidden.
		if grep -qF '"msg":"could not get certificate from issuer"' "$E2E_LOG" 2>/dev/null; then
			echo "Caddy could not obtain a certificate:"
			grep -F '"msg":"could not get certificate from issuer"' "$E2E_LOG"
			exit 1
		fi

		if ! kill -0 "$caddy_pid" 2>/dev/null; then
			echo "Caddy stopped unexpectedly."
			show_log
			exit 1
		fi

		if [ "$SECONDS" -ge "$deadline" ]; then
			echo "Timed out after ${timeout}s waiting for a certificate for $name."
			show_log
			exit 1
		fi

		sleep 5
	done

	echo "Obtained a certificate for $name"
done
