# Post-Installation Hardening

LiteNAS installs a security baseline, but host hardware, network exposure, and
administrator policy still require review after installation.

## Operator Workflow

1. Verify the installed service, firewall, AppArmor, USBGuard, AIDE, and log
   rotation configuration against the actual host.
2. Initialize or update the AIDE baseline after planned package and policy
   changes, then review subsequent integrity reports.
3. Install OpenSCAP separately if compliance assessment is needed. Select
   benchmark content and a profile appropriate to the host, run the assessment,
   and review each result in its deployment context.
4. Record intentional deviations in [Security Policy
   Exceptions](./security-policy-exceptions.md). Apply approved host changes
   through the appropriate system or LiteNAS configuration and reassess.
5. Review SIEM alerts, malware scan results, backup recovery, and attached
   device policy as ongoing operational work.

OpenSCAP results are compliance evidence for the selected content and profile.
LiteNAS does not provide a separate assessment engine or automatic remediation
workflow. AIDE integrity checks and SIEM detections continue to serve their
own runtime purposes.

See [Preinstalled Hardened Configs](./preinstalled-hardened-configs.md) for the
packaged baseline and [Security Tooling](./security-tooling.md) for tool roles.
