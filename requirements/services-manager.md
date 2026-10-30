# Services Manager requirements

- `services-manager/FR-001`: List loaded and installed systemd system service units by authenticated NATS RPC,
  returning unit name, load, active and substate, description, and enablement where available.
- `services-manager/FR-002`: Expose start, stop, restart, enable and disable RPCs for any syntactically valid
  system service unit.
- `services-manager/SR-001`: Require a valid LiteNAS access token carrying administrator-equivalent `admin` or
  `sudo` role for every RPC, including list.
- `services-manager/SR-002`: Reject malformed JSON, unknown fields, oversized requests, invalid unit names, and
  option-shaped operands before invoking systemctl.
- `services-manager/IR-001`: Use one dedicated service account, a restricted sudoers entry, and a bounded worker
  queue for command execution. Invoke systemctl without a shell.
- `services-manager/IR-002`: Include the service in local build, deploy, test, and CI workflows.
- `services-manager/SR-003`: Run as a dedicated service user under a required AppArmor profile that grants only its
  own configuration, certificates, log, sudo policy access, and systemctl communication.
