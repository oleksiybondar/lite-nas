# Moments skeleton

Moments currently contains browser authentication, an empty home page, the
shared user dropdown, and shared profile/appearance settings. Its appearance
preference applies to the entire page in both embedded and standalone views.
Album and media APIs are not part of this first increment.

## Local build and deployment

Build the two artifacts with `./scripts/build-moments-binary.sh` and
`./scripts/build-moments.sh`. `./scripts/deploy-moments.sh` builds and installs
both, the loopback service, and the Nginx configuration. The script requires
root privileges and accepts `--binary`, `--assets-source`, `--no-start`, and
`--skip-bootstrap` for an existing prepared host.

After deployment, open `https://moments.localhost/`,
`https://localhost/apps/moments/`, or the admin panel's **Applications → Moments**
page at `/applications/moments`. The admin panel embeds the same-origin
`/apps/moments/` Nginx mount in an iframe. A manual command such as
`curl --resolve moments.localhost:443:127.0.0.1` can resolve the local host
without DNS changes. The local self-signed certificate includes
`moments.localhost` after rotation. Moments uses its own `moments.crt` and
`moments.key`; adding Moments no longer replaces the admin host certificate.
Browsers must trust each self-signed certificate separately for direct local
HTTPS access.

With empty `cookie_domain` settings, each host has a host-only session and
requires its own sign-in. To share a session across real subdomains, set the
same parent DNS domain in `/etc/lite-nas/web-gateway.conf` and
`/etc/lite-nas/moments.conf`, then restart both services. For example, an
admin host `nas.example.com` and Moments host `moments.example.com` use
`cookie_domain=example.com`. Use HTTPS on both browser origins.

## Cloudflare tunnel

Publish a Moments hostname through a Cloudflare tunnel to
`http://127.0.0.1:80`, preserving the external hostname in the Host header
and forwarding `X-Forwarded-Proto: https`.
The standalone Nginx server matches `moments.<domain>` and keeps the origin
listener on loopback. Add the corresponding DNS record to the tunnel. The
browser reaches `/api/`, `/f/`, `/d/`, and `/ls/` through the same origin, so
no cross-origin token transfer is needed. For direct HTTPS without a tunnel,
install a certificate covering the production hostname and adjust the Nginx
TLS certificate paths.

The existing File Service snippet handles uploads under `/f/`, along with
source files under `/f/`, downloads under `/d/`, and listings under `/ls/`.
The skeleton UI does not expose those media workflows yet.
