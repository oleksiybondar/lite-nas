# Security Policy Exceptions

## Purpose

This document records intentional deviations from generic security benchmark
recommendations when those recommendations do not fit the LiteNAS deployment
model.

The goal is to make exceptions explicit, reviewable, and traceable instead of
allowing them to remain undocumented behavior.

## How To Record An Exception

Each exception should capture:

- the benchmark or recommendation being considered
- the LiteNAS decision
- the reason the generic recommendation does not fit
- the risk introduced by deviating
- the compensating controls or operator expectations
- the deployment scope where the exception applies

## Initial Exception Register

### USB Mass Storage Support

- Recommendation: Generic hardening guidance may recommend disabling or
  blacklisting removable storage support such as the `usb-storage` kernel
  module.
- LiteNAS decision: Keep `usb-storage` enabled and available.
- Rationale: LiteNAS targets NAS deployments, including Raspberry Pi systems
  that depend on USB-attached mass storage as a primary storage interface.
- Risk introduced: Allowing USB mass storage support increases the removable
  media attack surface and makes it easier to attach untrusted storage devices.
- Compensating controls: Use `usbguard`, apply operator review of attached
  devices, restrict physical access where possible, and rely on integrity and
  malware-monitoring controls as additional layers.
- Applies to: LiteNAS NAS hosts that use USB-attached storage, especially
  Raspberry Pi deployments.

## Review Expectations

Policy exceptions should be reviewed whenever:

- a new hardening control is introduced
- a benchmark recommendation is evaluated for adoption
- the target hardware model changes
- the LiteNAS threat model changes
- compensating controls become stronger or weaker
