# Session Manager requirements

- `session-manager/FR-001`: List all loginctl sessions by authenticated NATS RPC, returning identity,
  seat, terminal, class, and the `show-session` properties used to distinguish graphical, SSH, local
  terminal, system manager, and unknown sessions.
- `session-manager/FR-002`: Expose terminate, lock, and unlock RPCs for any session ID, including sessions owned
  by other users.
- `session-manager/SR-001`: Require a valid LiteNAS access token carrying `lite-nas-security`, `admin`, or `sudo`
  role for every RPC, including list.
- `session-manager/SR-002`: Reject malformed JSON, unknown fields, oversized requests, invalid session IDs, and
  option-shaped operands before invoking loginctl.
- `session-manager/IR-001`: Use one dedicated service account, a restricted sudoers entry, and a bounded worker
  queue for command execution. Invoke loginctl without a shell.
- `session-manager/IR-002`: Include the service in local build, deploy, test, and CI workflows.
- `session-manager/SR-003`: Run as a dedicated service user under a required AppArmor profile that grants only its
  own configuration, certificates, log, sudo policy access, and loginctl communication.
