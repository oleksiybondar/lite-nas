# USB Manager requirements

- `usb-manager/FR-001`: List USBGuard devices through authenticated NATS RPC.
- `usb-manager/FR-002`: Allow or block a USBGuard device by its numeric device ID; callers may request a
  persistent policy change.
- `usb-manager/SR-001`: Require a valid LiteNAS access token carrying `lite-nas-security`, `admin`, or
  `sudo` for every RPC.
- `usb-manager/SR-002`: Reject malformed JSON, unknown fields, oversized requests, and invalid device IDs
  before invoking USBGuard.
- `usb-manager/IR-001`: Run as a dedicated service account with a bounded command worker. Invoke only the
  `/usr/bin/usbguard` executable through restricted sudoers rules and without a shell.
- `usb-manager/IR-002`: Require the service's AppArmor profile and grant access only to service
  configuration, certificates, logs, sudo policy, and USBGuard IPC.
- `usb-manager/IR-003`: Include the service in local build, deploy, test, CI, and Debian packaging
  workflows.
- `usb-manager/BR-001`: Back up and restore USBGuard policy files as a backup-restore target. Runtime
  device authorization state is not part of that target.
