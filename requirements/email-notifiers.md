# Email Notifier Requirements

- **email-notifiers/FR-001:** System and security email notifiers consume
  validated alerts from their respective logging managers, queue them through
  bounded local workers, render packaged HTML templates, and deliver through
  the configured local SMTP transport. Delivery failures are logged and remain
  observable; external delivery requires an operator-configured relay.
