# LiteNAS


LiteNAS is a project intent for a lightweight, secure self-hosted platform for Linux with ZFS, monitoring, and configuration validation. It is conceived as a modular system built around small services to support extensible automation, security, and media capabilities.


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

LiteNAS started from a real home lab setup built to provide a secure and usable NAS on top of Linux and ZFS. That underlying setup exists today as a CLI-managed system. LiteNAS is the project to evolve those ideas and operational experience into a lightweight, extensible platform for managing storage, system configuration, and services in a more consistent and automated way.

Rather than being defined as a single-purpose NAS solution, LiteNAS is intended as a modular system composed of small, focused services. The project direction is to bring together storage management, monitoring, security validation, and automation into a unified self-hosted platform.

The planned architecture follows a microservices-oriented model with event-driven communication, allowing components to remain loosely coupled while still working together as a cohesive system. A web-based interface is envisioned as the bridge between internal services and external access, making the platform usable without direct terminal interaction for routine tasks.

At this stage, the repository primarily captures project intent, architectural direction, and design ideas. Implementation work is expected to be introduced incrementally as the project takes shape.

## Goals & Philosophy

LiteNAS is built around a set of practical design principles:

- **Lightweight over complex**  
  Focus on minimal overhead and efficient resource usage, making it suitable for low-power systems as well as larger environments.

- **Secure by default**  
  Reduce attack surface, enforce least privilege, and validate system configuration continuously.

- **Self-hosted first**  
  Full control over data, services, and infrastructure without reliance on external platforms.

- **Modular and extensible**  
  Built as a collection of independent services that can evolve and scale over time.

- **Automation and validation**  
  System configuration should be verifiable, reproducible, and enforceable through policy-based validation.

- **Security aligned with best practices**  
  Design decisions are guided by established security principles and industry-standard frameworks, with a focus on practical, real-world applicability.

- **Practical over theoretical**  
  Features are driven by real usage needs rather than upfront design assumptions.

## Architecture Overview

LiteNAS is intended to evolve as a modular, event-driven platform composed of small, focused services rather than a single monolithic application. The goal is to keep the system lightweight, maintainable, and adaptable as new requirements emerge.

At its foundation, LiteNAS is planned to build on Linux and ZFS to provide core storage and system capabilities. Around that foundation, the platform is expected to grow with dedicated components responsible for areas such as monitoring, resource supervision, security checks, configuration validation, automation, and auxiliary services.

The overall design favors loose coupling between components. Rather than concentrating all logic in one place, LiteNAS is intended to use service boundaries and event-driven communication so that responsibilities remain separated and the system can evolve incrementally over time.

A web-based HMI is planned as the main user-facing entry point for administration and day-to-day interaction. Its role is not to replace the internal service model, but to provide a controlled interface to it, making the platform more usable without depending on direct terminal access for routine operations.

The current reference environment is a working home lab deployment managed primarily through the CLI, but the architectural direction is intentionally broader. While Raspberry Pi hardware is one practical target, the platform is meant to remain portable across Linux-based systems and not depend on a single hardware profile.

At a high level, LiteNAS can be viewed as consisting of:

- **System layer**  
  Linux, ZFS, storage devices, networking, and host-level configuration.

- **Service layer**  
  Focused components for monitoring, validation, security functions, automation, and auxiliary capabilities.

- **Messaging layer**  
  Event-driven coordination between components, supporting separation of concerns and future extensibility.

- **Interface layer**  
  User-facing access through a web-based HMI and other controlled entry points.

This architecture is expected to mature gradually as the project develops, with implementation details shaped by real operational needs rather than fixed upfront assumptions.

## Monorepo Structure

LiteNAS is intended to be organized as a monorepo so that core components, services, and shared logic can evolve together.

The exact structure has not yet been implemented and is expected to take shape as the project develops. Individual services and modules are expected to include their own documentation, with more detailed READMEs provided within subprojects where needed.

## Core Features

LiteNAS is defined as an evolving platform, with capabilities expected to be added incrementally based on real-world usage. The following areas represent the intended core focus of the system:

- **ZFS-based storage management**  
  Reliable storage built on top of ZFS, with a focus on data integrity, flexibility, and efficient use of resources.

- **Monitoring and resource supervision**  
  Continuous insight into system state, including CPU, memory, storage, and service-level metrics.

- **Configuration validation**  
  Policy-based validation of system configuration, allowing detection of misconfigurations and drift from expected states.

- **Security and hardening**  
  Emphasis on reducing attack surface, enforcing least privilege, and maintaining a secure baseline configuration.

- **Event-driven service coordination**  
  Internal components are designed to communicate through an event-driven model, enabling loose coupling and extensibility.

- **Web-based interface (HMI)**  
  A user-facing interface for interacting with the system without relying on direct terminal access.

- **Remote access (VPN)**  
  Secure access to the platform through controlled network entry points.

- **Media services (planned)**  
  Support for media storage and browsing, including a gallery-style experience for personal content.

## Configuration Validation Engine

A central component in the LiteNAS design is the configuration validation engine, intended to provide a consistent and automated way to verify system state against expected policies.

The engine is intended to operate on structured rules, allowing system configuration to be described, validated, and reasoned about in a reproducible way. Rather than relying on manual checks or ad-hoc scripts, LiteNAS aims to treat configuration validation as a first-class capability once implemented.

### Rule-Based Validation

Validation is based on declarative rules that define expected system state. These rules may evaluate:

- Configuration files
- Command outputs
- System parameters
- Service states

Each rule is designed to be both machine-readable and human-understandable, enabling transparency and maintainability.

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

In addition to validation, the engine is intended to support remediation workflows. Where possible, rules may define how to adjust the system toward a compliant state.

The goal is to move from passive validation toward active configuration management, while maintaining transparency and control.

### Integration with the Platform

The validation engine is designed to integrate with other components of LiteNAS:

- Monitoring services can trigger validation checks
- Event-driven workflows can react to validation results
- The web interface can present validation status and insights to the user

This allows configuration validation to become part of the overall system lifecycle rather than a standalone tool.

### Design Approach

The validation engine is planned to be developed incrementally, with a focus on:

- Simplicity of rule definition
- Clear and interpretable results
- Extensibility for new validation types
- Practical applicability in real environments

The long-term goal is to provide a flexible foundation for enforcing system configuration policies in a lightweight and transparent way.

## Event-Driven Design

LiteNAS is designed around an event-driven approach to enable loose coupling between components and support incremental system evolution.

Instead of relying on tightly integrated control flows, components are intended to communicate through events that represent changes in system state, actions, or requests. This allows services to operate independently while still participating in coordinated workflows.

### Decoupled Components

By using an event-driven model, individual services do not need direct knowledge of each other. This reduces dependencies and makes it easier to:

- Add or remove services without impacting the entire system
- Evolve components independently
- Isolate failures and limit their impact

### Asynchronous Workflows

Events enable asynchronous processing, allowing the system to react to changes without blocking execution. This is particularly useful for:

- Monitoring and alerting
- Configuration validation
- Automation tasks
- Background processing

### Messaging Layer

LiteNAS is expected to use a lightweight messaging system to facilitate communication between components. The focus is on simplicity, reliability, and minimal operational overhead.

The messaging layer acts as the backbone of the platform, enabling services to publish and consume events without tight coupling.

### Integration with Validation and Monitoring

The event-driven model integrates naturally with other parts of the system:

- Monitoring components can emit events when system state changes
- The validation engine can react to events and evaluate configuration
- Automation workflows can be triggered based on validation results or system conditions

This creates a feedback loop where system state, validation, and automation continuously interact.

### Design Approach

The event-driven architecture is introduced gradually, with a focus on:

- Keeping communication patterns simple and understandable
- Avoiding unnecessary complexity
- Ensuring that the system remains observable and debuggable

The goal is to enable flexibility and extensibility without sacrificing clarity or control.

## Technology Choices

LiteNAS is intended to favor simple, well-understood technologies that provide strong performance, low operational overhead, and good long-term maintainability.

The focus is on choosing tools that align with the goals of lightweight operation, clear architecture, and ease of development.

### Go

Go is the intended primary implementation language due to its balance between simplicity, performance, and practicality.

- **Lightweight and efficient**  
  Well-suited for system-level tooling and services with minimal resource overhead.

- **Simple concurrency model**  
  Built-in primitives make it easier to implement concurrent and event-driven components.

- **Maintainability**  
  Clear syntax and minimal language complexity help keep codebases understandable over time.

- **Practical ecosystem**  
  Strong standard library and good support for networking, system interaction, and service development.

Go also provides a relatively smooth transition path from scripting languages such as Python, while offering better control over performance and deployment.

### NATS

NATS is a planned messaging backbone for event-driven communication between components.

- **Simplicity**  
  Minimal configuration and straightforward operational model.

- **Lightweight**  
  Designed to run efficiently even on constrained systems.

- **High performance**  
  Supports fast, low-latency message exchange between services.

- **Flexible communication patterns**  
  Supports publish/subscribe and request/reply models, enabling a range of interaction styles.

NATS aligns well with the goals of keeping the system decoupled while avoiding the complexity of heavier messaging systems.

## Reference Setup (Home Lab)

LiteNAS is informed by a real, continuously used home setup rather than a purely experimental lab environment.

The original goal of this setup was to provide a low-power, reliable storage system for personal data such as photo and video libraries. It operates today as a shared storage and backup solution, managed primarily through the command line and based on simple mirroring on low-power hardware.

The current setup is based on:

- **Raspberry Pi 5**
- **Powered USB hub**
- **2 × 2TB WD Elements SSD (USB)**
- **Local network with VPN-based remote access (Tailscale)**

In its initial form, the system was accessed via CIFS and managed primarily through the command line. It provided:

- ~25–30 MB/s over network (CIFS)
- Up to ~200 MB/s direct disk writes
- Power consumption below ~30W

This demonstrated that a low-power system can deliver acceptable performance for home storage and backup use cases.

LiteNAS is intended to build on top of this working foundation. The goal is to evolve the setup into a more structured, secure, and reusable platform by introducing:

- Configuration validation instead of manual tuning
- Improved access methods (e.g. NFS alongside CIFS)
- Better usability through a web-based interface
- Stronger security and controlled access patterns

It is important to note that LiteNAS is not limited to this hardware. The platform is designed to be portable across Linux-based systems and can be deployed on more powerful machines or cloud environments if needed.

The Raspberry Pi environment serves as a lower-bound reference, ensuring that the system remains efficient while still being practical for everyday use.

## Security Approach

Security is a core aspect of the LiteNAS design and is treated as an ongoing process rather than a one-time configuration.

The intended system is designed to minimize exposure, enforce least privilege, and continuously validate its configuration to reduce the risk of misconfiguration and unintended access.

### Minimal Exposure

LiteNAS is intended to follow a model where services are not directly exposed to the public internet.

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

Remote access is expected to be handled through secure networking solutions such as Tailscale, allowing services to remain private while still being accessible when needed.

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

The goal is to maintain a balance between strong security practices and system usability, ensuring that the platform remains both safe and practical for everyday use.

## Development Approach

LiteNAS is intended to be developed incrementally, with a focus on building a solid foundation before introducing higher-level functionality.

The development process starts from the system itself. Core infrastructure, security hardening, and baseline configuration are first established and validated manually in the reference environment. This ensures that the platform is grounded in a working and well-understood setup before automation is introduced.

### Foundation First

Rather than implementing complex features upfront, development focuses on small, reusable building blocks:

- Core system interactions
- Basic service components
- Simple and well-defined interfaces

This approach helps ensure that each part of the system remains understandable and reliable as the platform evolves.

### Code Quality and Tooling

Early emphasis is intended for code quality, consistency, and maintainability:

- Linting and formatting rules should be defined from the start
- Automated checks should be integrated into development workflows
- Continuous integration pipelines are expected to enforce standards once implementation begins

A target level of test coverage is expected to be defined to help ensure reliability and reduce the risk of regressions.

### Iterative Development

Features are introduced gradually, based on real usage and practical needs:

- Functionality is added when required, not pre-designed in full
- Components are refined over time rather than rewritten
- Feedback from actual usage guides further development

This reduces the likelihood of large, disruptive refactoring efforts and helps keep the system stable as it grows.

### Balancing Simplicity and Structure

The goal is to maintain a balance between:

- **Simplicity** — avoiding unnecessary abstraction and complexity
- **Structure** — ensuring the system remains organized and extensible

By establishing clear foundations early, LiteNAS aims to support long-term evolution without requiring major architectural changes.

## Subprojects & Documentation

LiteNAS is intended to be structured as a monorepo containing multiple components and services that evolve together.

As implementation begins, each subproject is expected to include its own documentation, with dedicated README files providing more detailed information about functionality, usage, and design decisions where applicable.

Until then, the top-level repository serves primarily as the project definition and architectural starting point.

## License

This project is licensed under the GNU General Public License v3.0 (GPLv3).

See the LICENSE file for details.
