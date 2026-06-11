# Email Notifiers Setup

## Purpose

LiteNAS provides two email notifier services:

- `lite-nas-system-email-notifier`
- `lite-nas-security-email-notifier`

Both services subscribe to alert events over NATS, render HTML email from a
local template, and hand mail off to local Postfix on `127.0.0.1:25`.

External delivery is owned by Postfix. For real internet delivery, the target
machine must be configured with an authenticated upstream SMTP relay such as
Mailgun, Amazon SES, SendGrid, or Google Workspace SMTP relay.

## Files Installed by LiteNAS

Service config files:

- `/etc/lite-nas/system-email-notifier.conf`
- `/etc/lite-nas/security-email-notifier.conf`

Template directories:

- `/etc/lite-nas/system-email-notifier/alert.html`
- `/etc/lite-nas/security-email-notifier/alert.html`

Postfix managed config:

- `/etc/postfix/main.cf`
- `/etc/postfix/master.cf`

Postfix relay authentication file:

- `/etc/postfix/postfix.d/authentication.conf`

The relay authentication file is intentionally local-only:

- deploy creates it from a placeholder if it does not exist
- later deploys do not overwrite it
- credentials must not be committed to git

## Service Config Shape

Both notifier services use the same config structure.

Example `system-email-notifier.conf`:

```ini
[messaging]
url=tls://127.0.0.1:4222
client_name=system-email-notifier
ca=/etc/lite-nas/certificates/transport/root-ca.crt
cert=/etc/lite-nas/certificates/transport/lite-nas-sys-email-notifier/client.crt
key=/etc/lite-nas/certificates/transport/lite-nas-sys-email-notifier/client.key
timeout=5s

[email]
to=ops@example.com
cc=oncall@example.com
from=system-alert-notifier@example-sender-domain.tld
subject_prefix=[LiteNAS]

[smtp]
host=127.0.0.1
port=25
timeout=10s
helo=localhost

[logging]
level=warn
format=rfc5424
output=file
file_path=/var/log/lite-nas/system-email-notifier.log
```

Example `security-email-notifier.conf`:

```ini
[email]
to=security@example.com
cc=
from=security-alert-notifier@example-sender-domain.tld
subject_prefix=[LiteNAS]
```

Behavior:

- if both `to` and `cc` are empty, the notifier skips delivery
- `from` should use a sender domain that is verified with the SMTP relay
- the notifier always sends to local Postfix, not directly to the external
  provider

## Postfix Relay Authentication

The local machine must contain `/etc/postfix/postfix.d/authentication.conf`.

Template:

```sh
postfix_relay_host=
postfix_relay_port=
postfix_relay_username=
postfix_relay_password=
postfix_relay_tls_level=encrypt
```

Example for a generic authenticated relay:

```sh
postfix_relay_host=smtp.example-provider.tld
postfix_relay_port=587
postfix_relay_username=mailer@example-sender-domain.tld
postfix_relay_password=replace-with-real-password
postfix_relay_tls_level=encrypt
```

The deploy flow translates this file into Postfix runtime settings:

- `relayhost`
- `smtp_sasl_auth_enable`
- `smtp_sasl_password_maps`
- `smtp_sasl_security_options`
- `smtp_sasl_tls_security_options`

and generates:

- `/etc/postfix/sasl_passwd`
- `/etc/postfix/sasl_passwd.db`

## Local Setup Flow

1. Install or deploy LiteNAS runtime:

```bash
sudo ./scripts/deploy-all.sh
```

1. Edit notifier recipient config:

```bash
sudoedit /etc/lite-nas/system-email-notifier.conf
sudoedit /etc/lite-nas/security-email-notifier.conf
```

1. Create or update Postfix relay authentication:

```bash
sudo install -d -m 0755 /etc/postfix/postfix.d
sudoedit /etc/postfix/postfix.d/authentication.conf
```

1. Redeploy managed runtime config and restart services:

```bash
sudo ./scripts/deploy-all.sh
sudo systemctl restart lite-nas-system-email-notifier
sudo systemctl restart lite-nas-security-email-notifier
```

1. Verify effective Postfix relay configuration:

```bash
sudo postconf -n | rg 'relayhost|smtp_sasl|smtp_tls'
```

## Package Install Flow

When installed from the LiteNAS Debian package:

1. install the package
2. edit:
   - `/etc/lite-nas/system-email-notifier.conf`
   - `/etc/lite-nas/security-email-notifier.conf`
   - `/etc/postfix/postfix.d/authentication.conf`
3. run the packaged runtime deploy flow again or restart Postfix and notifiers

If the package post-install already created
`/etc/postfix/postfix.d/authentication.conf`, only replace the placeholder
values. Do not delete the file.

## Testing from CLI

### Verify Postfix Listener

```bash
ss -ltn sport = :25
ps -ef | rg 'postfix|master|smtpd'
```

### Verify Notifier Service State

```bash
systemctl status lite-nas-system-email-notifier --no-pager
systemctl status lite-nas-security-email-notifier --no-pager
journalctl -u lite-nas-system-email-notifier -n 50 --no-pager
journalctl -u lite-nas-security-email-notifier -n 50 --no-pager
```

### Send a Test System Alert

```bash
sudo system-logging-manager-cli \
  --cmd createEvent \
  --data "{\"event_id\":\"sysram_00009999\",\"category\":\"system.metrics.mem.used\",\"severity\":\"warning\",\"priority\":2,\"created_at\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",\"source\":\"system-metrics\",\"message\":\"RAM usage is above threshold\",\"trigger_value\":\"93\"}"
```

### Inspect Mail Flow

```bash
sudo tail -n 100 /var/log/mail.log
postqueue -p
```

## Capture Mode for Local Rendering Tests

For rendering and integration tests without real outbound delivery:

```bash
sudo ./scripts/postfix/configure-test-capture.sh --mode capture
```

Then send a test alert and inspect:

```bash
cat /var/tmp/lite-nas-postfix-test-mail.log
```

To restore normal Postfix behavior:

```bash
sudo ./scripts/postfix/configure-test-capture.sh --mode normal
```

## Operational Notes

- Gmail and other mailbox providers commonly reject unauthenticated direct mail.
- Direct delivery from this host without relay authentication is not a reliable
  production model.
- The sender address in notifier config should match a verified relay domain.
- Relay credentials are machine-local secrets and must not be stored in the
  repository.
