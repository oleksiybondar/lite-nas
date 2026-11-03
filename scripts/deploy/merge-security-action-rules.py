#!/usr/bin/env python3
"""Merge reviewed Security Automation action rules into an installed rule file.

The deployment keeps administrator changes and updates only known packaged
shapes from before action enrollment. A host-side backup is created once.
"""

import json
import os
import shutil
import sys
import tempfile
from pathlib import Path


def indexed(entries: list[dict]) -> dict[str, dict]:
    """Index validated rule objects by their stable configured identifier."""
    result: dict[str, dict] = {}
    for entry in entries:
        key = entry.get("id")
        if not isinstance(key, str) or key in result:
            raise ValueError("rule file has a missing or duplicate ID")
        result[key] = entry
    return result


def merge(installed: dict, packaged: dict, kind: str) -> bool:
    """Add new IDs and change only the known older action-rule definitions."""
    changed = False
    if kind == "log-analysis":
        additions = {
            "instances": ("web_api_account_source_retry", "sudo_policy_event"),
            "rules": ("pam_web_api_password_guessing",),
        }
    elif kind == "security-automation":
        additions = {
            "rules": (
                "web_password_guessing",
                "web_https_source_block",
                "ssh_probe_then_root_login",
                "sudo_denial_then_group_addition",
                "network_scan_sequence",
            ),
        }
    else:
        raise ValueError("unknown rule file kind")

    for section, names in additions.items():
        current = indexed(installed[section])
        desired = indexed(packaged[section])
        for name in names:
            if name not in current:
                installed[section].append(desired[name])
                changed = True

    current_rules = indexed(installed["rules"])
    desired_rules = indexed(packaged["rules"])
    if kind == "log-analysis":
        for name, current in current_rules.items():
            desired = desired_rules.get(name)
            if desired is None:
                continue
            for field in ("name", "explanation", "possible_reactions"):
                if field not in current:
                    current[field] = desired[field]
                    changed = True
        sudo = current_rules.get("sudo_policy_violation")
        if sudo is not None and sudo.get("instance") == "privilege_event":
            sudo["instance"] = "sudo_policy_event"
            changed = True
        current_instances = indexed(installed["instances"])
        usb = current_instances.get("usb_device_event")
        if usb is not None and usb.get("cooldown") == "1m":
            usb["cooldown"] = "1s"
            changed = True
    else:
        for name, old_key in (
            ("repeated_sudo_denial", "username"),
            ("usb_suspicious_activity", "device_id"),
        ):
            current = current_rules.get(name)
            desired = desired_rules[name]
            if current is None or current.get("correlate_by") != old_key or current.get("action") != desired["action"]:
                continue
            for field in ("name", "description", "fsm", "correlate_by", "steps", "explanation"):
                current[field] = desired[field]
            changed = True
        older_ip_explanations = {
            "ssh_brute_force": "Detects SSH authentication failures from one source address within five minutes. Intended protection is a temporary source IP block, but host blocking and restoration are not enrolled yet.",
            "account_enumeration": "Counts SSH account probes from one source address within five minutes. Intended protection is a temporary source IP block, but host blocking and restoration are not enrolled yet.",
            "network_reconnaissance": "Counts multiple network alerts from one source address within five minutes. Intended protection is a temporary source IP block, but host blocking and restoration are not enrolled yet.",
            "recon_then_ssh": "Matches a network scan followed by SSH failures from the same address within ten minutes. Intended protection is a temporary source IP block, but host blocking and restoration are not enrolled yet.",
        }
        for name, old_explanation in older_ip_explanations.items():
            current = current_rules.get(name)
            if current is not None and current.get("explanation") == old_explanation:
                current["explanation"] = desired_rules[name]["explanation"]
                changed = True
        previous_ip_explanations = {
            "ssh_brute_force": "Detects SSH failures from one source address within five minutes. An operator can request a 30-minute direct HTTPS block for that IP, but it does not block SSH. Automatic blocking is not enrolled.",
            "account_enumeration": "Counts SSH account probes from one source address within five minutes. An operator can request a 30-minute direct HTTPS block for that IP, but it does not block SSH. Automatic blocking is not enrolled.",
            "recon_then_ssh": "Matches a network scan followed by SSH failures from one address within ten minutes. An operator can request a 30-minute direct HTTPS block for that IP, but it does not block SSH. Automatic blocking is not enrolled.",
            "network_reconnaissance": "Counts network alerts from one source address within five minutes. An operator can request a 30-minute direct HTTPS block on verified physical interfaces. Automatic blocking is not enrolled until ingress provenance is verified.",
            "web_https_source_block": "A verified direct HTTPS source is blocked on physical interfaces for 30 minutes. Cloudflare Tunnel and Tailscale traffic are outside this UFW rule.",
        }
        for name, old_explanation in previous_ip_explanations.items():
            current = current_rules.get(name)
            if current is not None and current.get("explanation") == old_explanation:
                current["explanation"] = desired_rules[name]["explanation"]
                changed = True
        old_ssh_step = {
            "type": "security",
            "id_prefix": "secssh",
            "message": {"regex": "from (?P<requester_ip_address>[0-9a-fA-F:.]+)"},
        }
        for name, step_index in (
            ("ssh_brute_force", 0),
            ("account_enumeration", 0),
            ("recon_then_ssh", 1),
        ):
            current = current_rules.get(name)
            if current is not None and len(current.get("steps", [])) > step_index and current["steps"][step_index] == old_ssh_step:
                current["steps"][step_index] = desired_rules[name]["steps"][step_index]
                changed = True
    return changed


def main() -> int:
    """Write one atomic merge, retaining ownership and a one-time backup."""
    if len(sys.argv) != 4:
        raise ValueError("usage: merge-security-action-rules.py KIND INSTALLED PACKAGED")
    kind, installed_path, packaged_path = sys.argv[1:]
    installed_file = Path(installed_path)
    packaged_file = Path(packaged_path)
    installed = json.loads(installed_file.read_text(encoding="utf-8"))
    packaged = json.loads(packaged_file.read_text(encoding="utf-8"))
    if not merge(installed, packaged, kind):
        return 0
    backup = installed_file.with_name(installed_file.name + ".pre-action-enrollment")
    if not backup.exists():
        shutil.copy2(installed_file, backup)
    metadata = installed_file.stat()
    descriptor, temporary_path = tempfile.mkstemp(prefix=".automation-rules-", dir=installed_file.parent)
    try:
        with os.fdopen(descriptor, "w", encoding="utf-8") as stream:
            json.dump(installed, stream, indent=2, ensure_ascii=False)
            stream.write("\n")
            stream.flush()
            os.fsync(stream.fileno())
        os.chown(temporary_path, metadata.st_uid, metadata.st_gid)
        os.chmod(temporary_path, metadata.st_mode)
        os.replace(temporary_path, installed_file)
        directory_fd = os.open(installed_file.parent, os.O_RDONLY)
        try:
            os.fsync(directory_fd)
        finally:
            os.close(directory_fd)
    finally:
        if os.path.exists(temporary_path):
            os.unlink(temporary_path)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
