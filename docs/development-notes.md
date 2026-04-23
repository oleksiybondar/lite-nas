# Development Notes

## Why system-metrics came first

The `system-metrics` service is not the highest business-value capability in the
intended LiteNAS scope. By itself, it delivers only a small part of the overall
platform vision.

It was still chosen as the first implemented service on purpose.

`system-metrics` is simple enough to let the project establish the foundational
building blocks first:

- repository structure
- shared Go modules
- service runtime patterns
- configuration loading
- logging
- messaging integration
- test conventions
- build scripts
- deployment scripts
- packaging
- CI/CD workflow structure

This follows the intended LiteNAS development approach: build the platform from
small, understandable bricks first, and only then expand into higher-value
business logic and broader product capabilities.

The goal is to avoid jumping directly into complex domain behavior before the
base infrastructure is reproducible, testable, and maintainable.

In practice, `system-metrics` acts as an infrastructure-seeding service rather
than a statement of final product priority.
