# LiteNAS

LiteNAS is a lightweight, security-oriented self-hosted platform for Linux that is being built as a
set of small services, apps, and shared modules.

The repository already contains the initial platform skeleton:

- shared Go modules for logging, configuration, messaging, and metrics support
- initial backend service and consumer app slices
- NATS-based internal communication foundations
- Debian packaging, deployment scripts, and CI validation

The broader platform vision still extends beyond what is implemented today. Storage management,
web-facing administration, configuration validation, and further hardening are being added
incrementally on top of this initial working slice.

## Table of Contents

- [Overview](#overview)
- [Goals & Philosophy](#goals--philosophy)
- [Architecture Overview](#architecture-overview)
- [Monorepo Structure](#monorepo-structure)
- [Core Features](#core-features)
- [Configuration Validation Engine](#configuration-validation-engine)
- [Event-Driven Design](#event-driven-design)
- [Technology Choices](#technology-choices)
  - [Go](#go)
  - [NATS](#nats)
- [Reference Setup (Home Lab)](#reference-setup-home-lab)
- [Security Approach](#security-approach)
- [Development Approach](#development-approach)
- [Subprojects & Documentation](#subprojects--documentation)
- [License](#license)

## Overview

LiteNAS started from a real home lab system used for low-power Linux and ZFS-based storage. That
operational baseline still matters, but the repository is no longer only a design sketch.

The project now has an implemented first slice that seeds the platform architecture:

- a shared internal module layer
- an initial backend service
- an initial consumer app
- messaging-based service interaction over NATS
- reproducible packaging and install validation
- CI/CD checks for analysis, build, test, packaging, and package installability

This first slice is intentionally low direct business value. Its purpose is to establish the
platform shape in a reproducible way before broader service, app, and interface layers are added.

## Goals & Philosophy

LiteNAS is built around a set of practical design principles:

- **Lightweight over complex**
  Focus on minimal overhead and efficient resource usage, making it suitable for low-power systems
  as well as larger environments.

- **Secure by default**
  Reduce attack surface, enforce least privilege, and validate system configuration continuously.

- **Self-hosted first**
  Full control over data, services, and infrastructure without reliance on external platforms.

- **Modular and extensible**
  Built as a collection of independent services that can evolve and scale over time.

- **Automation and validation**
  System configuration should be verifiable, reproducible, and enforceable through policy-based
  validation.

- **Security aligned with best practices**
  Design decisions are guided by established security principles and industry-standard frameworks,
  with a focus on practical, real-world applicability.

- **Practical over theoretical**
  Features are driven by real usage needs rather than upfront design assumptions.

## Architecture Overview

LiteNAS is being built as a modular, event-driven platform composed of small, focused services
rather than a single monolithic application.

At a high level, the platform consists of:

- **System layer**
  Linux, ZFS, storage devices, networking, and host-level configuration.

- **Shared module layer**
  Reusable Go packages for logging, configuration, messaging, metrics, and related runtime support.

- **Service layer**
  Focused backend services added incrementally over time.

- **App layer**
  Consumer applications and later browser-facing frontend apps.

- **Messaging layer**
  NATS-based communication used to keep service boundaries explicit and loosely coupled.

- **Interface layer**
  Controlled entrypoints for CLI and later browser-facing administration.

The current implemented architecture is intentionally small, but it already reflects the intended
direction: thin service boundaries, shared foundations, explicit messaging, and reproducible build
and packaging flow.

A dedicated auth authority keeps PAM-backed host authentication, token
issuance, and emergency auth-state control out of the browser gateway.

## Monorepo Structure

LiteNAS is organized as a monorepo so that services, apps, packaging, CI/CD, and shared logic can
evolve together.

The structure is already taking shape around a few main areas:

- `shared/`
  Shared Go modules used by multiple services and apps.

- `services/`
  Backend services and gateway components.

- `apps/`
  Consumer applications and frontend app projects.

- `requirements/`
  Requirement documents used to define services and apps before or alongside implementation.

- `scripts/`
  Build, test, CI, deployment, formatting, packaging, and developer helper scripts.

- `packaging/`
  Debian packaging templates, metadata, and supporting files.

- `.github/`
  CI workflows and reusable composite actions.

## Core Features

LiteNAS is an evolving platform. Some capabilities already exist in initial form, while others are
still planned.

Current implemented focus:

- **Shared runtime and messaging foundations**
  Reusable internal modules for logging, configuration, messaging, and metrics support.

- **Monitoring seed slice**
  An initial service/app slice used to establish service wiring, messaging flow, and packaging.

- **Event-driven service communication**
  Internal communication over NATS using request/reply and event-oriented patterns.

- **Reproducible packaging and deployment**
  Debian packaging, deployment scripts, and install validation for the current platform slice.

- **CI/CD validation**
  Analysis, build, test, package, and installability checks wired into the repository workflow.

Planned expansion areas:

- **Dedicated host auth authority**
  `auth-service` verifies real login-capable users of the managed machine
  through PAM-backed flows, issues short-lived JWT access tokens plus
  refresh-backed sessions, and publishes auth-state events such as lockdown
  transitions over NATS.

- **ZFS-based storage management**
  Reliable storage built on top of ZFS, with a focus on data integrity, flexibility, and efficient
  use of resources.

- **Monitoring and resource supervision**
  Continuous insight into system state, including CPU, memory, storage, and service-level metrics.

- **Configuration validation**
  Policy-based validation of system configuration, allowing detection of misconfigurations and drift
  from expected states.

- **Security and hardening**
  Emphasis on reducing attack surface, enforcing least privilege, and maintaining a secure baseline
  configuration.

- **Event-driven service coordination**
  Internal components are designed to communicate through an event-driven model, enabling loose
  coupling and extensibility.

- **Web-based administration**
  Browser-facing administration built on top of the same service and packaging foundations.

- **Remote access (VPN)**
  Secure access to the platform through controlled network entry points.

- **Media services (planned)**
  Support for media storage and browsing, including a gallery-style experience for personal content.

## Configuration Validation Engine

A central component in the LiteNAS design is the configuration validation engine, intended to
provide a consistent and automated way to verify system state against expected policies.

The engine is intended to operate on structured rules, allowing system configuration to be
described, validated, and reasoned about in a reproducible way. Rather than relying on manual checks
or ad-hoc scripts, LiteNAS aims to treat configuration validation as a first-class capability once
implemented.

### Rule-Based Validation

Validation is based on declarative rules that define expected system state. These rules may
evaluate:

- Configuration files
- Command outputs
- System parameters
- Service states

Each rule is designed to be both machine-readable and human-understandable, enabling transparency
and maintainability.

### Validation States

Each validation produces one of three outcomes:

- **Pass**
  The system matches the expected configuration.

- **Pass with warning**
  The configuration is acceptable but not optimal, often involving trade-offs.

- **Error**
  The configuration does not meet the expected requirements and may require remediation.

This allows the system to distinguish between strict failures and acceptable deviations.

### Context and Rationale

Rules may include additional context describing:

- Why a configuration is considered correct
- What alternatives are acceptable
- Trade-offs between different configurations

This makes validation results more actionable and easier to interpret.

### Remediation (Planned)

In addition to validation, the engine is intended to support remediation workflows. Where possible,
rules may define how to adjust the system toward a compliant state.

The goal is to move from passive validation toward active configuration management, while
maintaining transparency and control.

### Integration with the Platform

The validation engine is designed to integrate with other components of LiteNAS:

- Monitoring services can trigger validation checks
- Event-driven workflows can react to validation results
- The web interface can present validation status and insights to the user

This allows configuration validation to become part of the overall system lifecycle rather than a
standalone tool.

### Design Approach

The validation engine is planned to be developed incrementally, with a focus on:

- Simplicity of rule definition
- Clear and interpretable results
- Extensibility for new validation types
- Practical applicability in real environments

The long-term goal is to provide a flexible foundation for enforcing system configuration policies
in a lightweight and transparent way.

## Event-Driven Design

LiteNAS is designed around an event-driven approach to enable loose coupling between components and
support incremental system evolution.

Instead of relying on tightly integrated control flows, components are intended to communicate
through events that represent changes in system state, actions, or requests. This allows services to
operate independently while still participating in coordinated workflows.

### Decoupled Components

By using an event-driven model, individual services do not need direct knowledge of each other. This
reduces dependencies and makes it easier to:

- Add or remove services without impacting the entire system
- Evolve components independently
- Isolate failures and limit their impact

### Asynchronous Workflows

Events enable asynchronous processing, allowing the system to react to changes without blocking
execution. This is particularly useful for:

- Monitoring and alerting
- Configuration validation
- Automation tasks
- Background processing

### Messaging Layer

LiteNAS is expected to use a lightweight messaging system to facilitate communication between
components. The focus is on simplicity, reliability, and minimal operational overhead.

The messaging layer acts as the backbone of the platform, enabling services to publish and consume
events without tight coupling.

### Integration with Validation and Monitoring

The event-driven model integrates naturally with other parts of the system:

- Monitoring components can emit events when system state changes
- The validation engine can react to events and evaluate configuration
- Automation workflows can be triggered based on validation results or system conditions

This creates a feedback loop where system state, validation, and automation continuously interact.

### Event Design Approach

The event-driven architecture is introduced gradually, with a focus on:

- Keeping communication patterns simple and understandable
- Avoiding unnecessary complexity
- Ensuring that the system remains observable and debuggable

The goal is to enable flexibility and extensibility without sacrificing clarity or control.

## Technology Choices

LiteNAS is intended to favor simple, well-understood technologies that provide strong performance,
low operational overhead, and good long-term maintainability.

The focus is on choosing tools that align with the goals of lightweight operation, clear
architecture, and ease of development.

### Go

Go is the intended primary implementation language due to its balance between simplicity,
performance, and practicality.

- **Lightweight and efficient**
  Well-suited for system-level tooling and services with minimal resource overhead.

- **Simple concurrency model**
  Built-in primitives make it easier to implement concurrent and event-driven components.

- **Maintainability**
  Clear syntax and minimal language complexity help keep codebases understandable over time.

- **Practical ecosystem**
  Strong standard library and good support for networking, system interaction, and service
  development.

Go also provides a relatively smooth transition path from scripting languages such as Python, while
offering better control over performance and deployment.

### NATS

NATS is the selected messaging backbone for event-driven communication between components and is
already used by the initial implemented platform slice.

- **Simplicity**
  Minimal configuration and straightforward operational model.

- **Lightweight**
  Designed to run efficiently even on constrained systems.

- **High performance**
  Supports fast, low-latency message exchange between services.

- **Flexible communication patterns**
  Supports publish/subscribe and request/reply models, enabling a range of interaction styles.

NATS aligns well with the goals of keeping the system decoupled while avoiding the complexity of
heavier messaging systems.

## Reference Setup (Home Lab)

LiteNAS is informed by a real, continuously used home setup rather than a purely experimental lab
environment.

The original goal of this setup was to provide a low-power, reliable storage system for personal
data such as photo and video libraries. It operates today as a shared storage and backup solution,
managed primarily through the command line and based on simple mirroring on low-power hardware.

The current setup is based on:

- **Raspberry Pi 5**
- **Powered USB hub**
- **2 × 2TB WD Elements SSD (USB)**
- **Local network with VPN-based remote access (Tailscale)**

In its initial form, the system was accessed via CIFS and managed primarily through the command
line. It provided:

- ~25–30 MB/s over network (CIFS)
- Up to ~200 MB/s direct disk writes
- Power consumption below ~30W

This demonstrated that a low-power system can deliver acceptable performance for home storage and
backup use cases.

LiteNAS is intended to build on top of this working foundation. The goal is to evolve the setup into
a more structured, secure, and reusable platform by introducing:

- Configuration validation instead of manual tuning
- Improved access methods (e.g. NFS alongside CIFS)
- Better usability through a web-based interface
- Stronger security and controlled access patterns

It is important to note that LiteNAS is not limited to this hardware. The platform is designed to be
portable across Linux-based systems and can be deployed on more powerful machines or cloud
environments if needed.

The Raspberry Pi environment serves as a lower-bound reference, ensuring that the system remains
efficient while still being practical for everyday use.

## Security Approach

Security is a core aspect of the LiteNAS design and is treated as an ongoing process rather than a
one-time configuration.

The intended system is designed to minimize exposure, enforce least privilege, and continuously
validate its configuration to reduce the risk of misconfiguration and unintended access.

### Minimal Exposure

LiteNAS is intended to follow a model where services are not directly exposed to the public
internet.

- External access is restricted to controlled entry points
- Internal services remain isolated within the private network
- Remote access is provided through VPN-based solutions

This approach reduces the attack surface and avoids unnecessary public endpoints.

### Controlled Access

Access to the system is intended to be explicit and limited:

- Authentication is required for all remote access
- Network boundaries are enforced between internal and external components
- Services are exposed only when necessary and in a controlled manner

### VPN-First Connectivity

Remote access is expected to be handled through secure networking solutions such as Tailscale,
allowing services to remain private while still being accessible when needed.

This avoids the need for direct WAN exposure and simplifies the overall security model.

### Configuration Validation

Security is intended to be reinforced through the configuration validation engine:

- System state can be continuously checked against defined policies
- Misconfigurations can be detected early
- Security-related settings can be verified automatically

This helps maintain a consistent and auditable system state over time.

### Principle of Least Privilege

LiteNAS is intended to be designed with the principle of least privilege in mind:

- Services should operate with only the permissions they require
- Access between components should be restricted by default
- Sensitive operations should be isolated where possible

### Incremental Hardening

Security improvements are introduced incrementally as the system evolves:

- Hardening decisions are validated in a real environment
- Changes are driven by practical needs and observed risks
- Complexity is avoided unless it provides clear security benefits

The goal is to maintain a balance between strong security practices and system usability, ensuring
that the platform remains both safe and practical for everyday use.

## Development Approach

LiteNAS is developed incrementally, with a deliberate focus on building a reproducible foundation
before introducing higher-level functionality.

The development process starts from the system itself. Core infrastructure, packaging, service
wiring, and baseline configuration are established first so later product work can build on a
repeatable base instead of ad hoc local setup.

### Foundation First

Rather than implementing complex features upfront, development focuses on small, reusable building
blocks and infrastructure-completing slices:

- Core system interactions
- Basic service and app components
- Simple and well-defined interfaces
- Reproducible packaging and CI/CD flow

This approach helps ensure that each part of the system remains understandable and reliable as the
platform evolves.

### Code Quality and Tooling

The repository already emphasizes code quality, consistency, and maintainability:

- Linting and formatting rules are defined from the start
- Automated checks are integrated into development workflows
- CI pipelines enforce analysis, build, test, and package validation
- Repo-wide duplication checks are enforced for Go and shell code rather than
  only within individual language-module boundaries

A target level of test coverage is used to help ensure reliability and reduce the risk of
regressions.

The repository also treats test duplication as real maintenance debt. Repeated
test setup is expected to move into named helpers, fixture builders, and
`testutil` packages when reuse crosses package, subproject, or module
boundaries.

### Iterative Development

Features are introduced gradually, based on real usage and practical needs:

- Functionality is added when required, not pre-designed in full
- Components are refined over time rather than rewritten
- Feedback from actual usage guides further development

This reduces the likelihood of large, disruptive refactoring efforts and helps keep the system
stable as it grows.

### Balancing Simplicity and Structure

The goal is to maintain a balance between:

- **Simplicity** — avoiding unnecessary abstraction and complexity
- **Structure** — ensuring the system remains organized and extensible

By establishing clear foundations early, LiteNAS aims to support long-term evolution without
requiring major architectural changes.

## Subprojects & Documentation

The top-level repository now acts as both:

- the high-level platform overview
- the entry point to implemented subprojects and planning documents

Useful starting points:

- [`docs/development-notes.md`](docs/development-notes.md)
  Why early slices are intentionally infrastructure-heavy and low immediate business value.

- [`RELEASE_NOTES.md`](RELEASE_NOTES.md)
  Release-level summary of what has already been established and what later slices add.

- [`requirements/system-metrics.md`](requirements/system-metrics.md)
  Example requirements for an initial backend service slice.

- [`requirements/system-metrics-cli.md`](requirements/system-metrics-cli.md)
  Example requirements for an initial consumer app slice.

- [`requirements/web-gateway.md`](requirements/web-gateway.md)
  Requirements for the browser-facing gateway.

- [`requirements/auth-service.md`](requirements/auth-service.md)
  Requirements for the PAM-backed authentication authority.

- [`services/web-gateway/README.md`](services/web-gateway/README.md)
  Architectural role and boundaries of the browser-facing gateway.

- [`services/auth/README.md`](services/auth/README.md)
  Architectural role and boundaries of the auth service.

- [`apps/admin-panel/README.md`](apps/admin-panel/README.md)
  Naming and role of the planned browser application.

## License

This project is licensed under the GNU General Public License v3.0 (GPLv3).

See the LICENSE file for details.
