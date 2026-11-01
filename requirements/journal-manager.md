# Journal Manager requirements

- `journal-manager/FR-001`: Return bounded chunks of the whole host system journal through an authenticated NATS
  RPC. Support an optional UTC time range, literal message search, systemd unit, priority, and an opaque cursor for
  loading further entries in newest-first order. Return the journal cursor, timestamp, message, unit, priority,
  identifier, and process ID where available; tolerate missing or non-text fields.
- `journal-manager/FR-002`: Expose a vacuum action that rotates active journals and retains the user-selected number
  of days. Report command failure, and publish a security alert identifying the authenticated actor and retention
  period when vacuum succeeds. A publication failure must be observable and must not be reported as an unqualified
  success.
- `journal-manager/SR-001`: Require a valid LiteNAS token with the same security-equivalent roles as Session Manager
  for every RPC and gateway route. Place the browser page under Security with the same role gate.
- `journal-manager/SR-002`: Validate request size, JSON shape, dates, time order, search text, unit, priority,
  cursor, chunk size, and retention days before command execution. Invoke fixed executables without a shell; bound
  command time and output size.
- `journal-manager/IR-001`: Run as a dedicated service account with read-only journal access for queries. Use
  narrowly scoped sudoers privileges for vacuum only, a required AppArmor profile, bounded command work, and NATS
  authorization for only the query, vacuum, reply, and security-alert subjects.
- `journal-manager/IR-002`: Integrate with the existing build, local deployment, certificate rotation, Debian
  package, postinstall, backup configuration, and CI scripts.
- `journal-manager/FR-003`: Add a Security Journal UI similar to Sessions. Show journal entries in a modal with Load
  more, literal search highlighting, structured journal fields and a plain-message fallback, filters, and a
  confirmed vacuum action. The systemd service-unit picker lists units with entries matching the current message,
  priority, and time filters. Localize all UI text in English, Ukrainian, and Danish.
- `journal-manager/IR-003`: Cover command validation and parsing with Go tests, browser-facing routes and UI
  behavior with existing test conventions, and installed service behavior with category-marked HyperionTF system
  tests.
- `journal-manager/BR-001`: Expose Journal Manager as a Backup Restore file target for its packaged configuration.
  Restoring that target requests a restart of `lite-nas-journal-manager.service`. The shared systemd and AppArmor
  targets cover its packaged unit and profile; archived journal records and the privileged sudoers policy are outside
  this file target.
