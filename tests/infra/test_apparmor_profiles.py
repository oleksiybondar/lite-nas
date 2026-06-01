"""System infrastructure test suite for deployed AppArmor policy profiles."""

import json
import shlex
from typing import TypeGuard

import pytest
from hyperiontf import CLIClient, expect

APPARMOR_PROFILE_CASES = [
    pytest.param("/usr/libexec/lite-nas/auth-service", id="auth-service-profile"),
    pytest.param("/usr/libexec/lite-nas/rbac-service", id="rbac-service-profile"),
    pytest.param("/usr/libexec/lite-nas/resources-monitor", id="resources-monitor-profile"),
    pytest.param("/usr/libexec/lite-nas/system-metrics", id="system-metrics-profile"),
    pytest.param("/usr/libexec/lite-nas/zfs-metrics", id="zfs-metrics-profile"),
    pytest.param(
        "/usr/libexec/lite-nas/system-logging-manager", id="system-logging-manager-profile"
    ),
    pytest.param(
        "/usr/libexec/lite-nas/security-logging-manager", id="security-logging-manager-profile"
    ),
    pytest.param("/usr/libexec/lite-nas/system-email-notifier", id="system-email-notifier-profile"),
    pytest.param(
        "/usr/libexec/lite-nas/security-email-notifier", id="security-email-notifier-profile"
    ),
    pytest.param("/usr/libexec/lite-nas/web-gateway", id="web-gateway-profile"),
    pytest.param("/usr/libexec/lite-nas/system-metrics-cli", id="system-metrics-cli-profile"),
    pytest.param("/usr/libexec/lite-nas/zfs-metrics-cli", id="zfs-metrics-cli-profile"),
    pytest.param(
        "/usr/libexec/lite-nas/system-logging-manager-cli", id="system-logging-manager-cli-profile"
    ),
    pytest.param(
        "/usr/libexec/lite-nas/security-logging-manager-cli",
        id="security-logging-manager-cli-profile",
    ),
    pytest.param("/usr/sbin/nginx", id="nginx-profile"),
    pytest.param("/usr/lib/postfix/sbin/*", id="postfix-profile"),
]

APPARMOR_WRAPPED_STATUS_SCHEMA: dict[str, object] = {
    "type": "object",
    "required": ["profiles"],
    "properties": {
        "version": {"type": ["string", "number"]},
        "profiles": {
            "type": "object",
            "additionalProperties": {"type": "string"},
        },
    },
    "additionalProperties": True,
}

APPARMOR_BARE_STATUS_SCHEMA: dict[str, object] = {
    "type": "object",
    "additionalProperties": {"type": "string"},
}

APPARMOR_STATUS_SCHEMA: dict[str, object] = {
    "anyOf": [APPARMOR_WRAPPED_STATUS_SCHEMA, APPARMOR_BARE_STATUS_SCHEMA],
}


class ProfileNotFoundError(AssertionError):
    """Raised when a filtered aa-status payload does not contain the expected profile key."""

    def __init__(self, profile_name: str, available_profiles: list[str]) -> None:
        super().__init__(
            f"Expected profile {profile_name!r} in aa-status JSON, got keys: "
            f"{available_profiles!r}"
        )


class ProfileModeMismatchError(AssertionError):
    """Raised when a filtered aa-status profile does not report enforce mode."""

    def __init__(self, profile_name: str, actual_mode: str) -> None:
        super().__init__(f"Expected profile {profile_name!r} in enforce mode, got {actual_mode!r}")


class JSONPayloadNotFoundError(AssertionError):
    """Raised when aa-status output does not include parseable JSON payload."""

    def __init__(self, output: str) -> None:
        super().__init__(f"aa-status did not return parseable JSON output: {output!r}")


class InvalidProfileStatesMappingError(AssertionError):
    """Raised when aa-status JSON does not expose a string-to-string profile mapping."""

    def __init__(self, value: object) -> None:
        super().__init__(f"Expected string profile mapping, got {value!r}")


def assert_profile_count(
    testsudo_cli_client: CLIClient,
    profile_name: str,
    *,
    enforce_mode_only: bool,
) -> None:
    """Verify filtered aa-status JSON output contains the expected profile state."""
    mode_filter = "--filter.mode=enforce " if enforce_mode_only else ""
    testsudo_cli_client.execute(
        "sudo -n aa-status "
        "--show=profiles "
        f"{mode_filter}"
        f"--filter.profiles={shlex.quote(profile_name)} "
        "--json --quiet"
    )
    testsudo_cli_client.assert_exit_code(0)
    payload = extract_json_payload(testsudo_cli_client.output)
    profile_states = extract_profile_states(payload)
    if profile_name not in profile_states:
        raise ProfileNotFoundError(profile_name, sorted(profile_states.keys()))
    if enforce_mode_only and profile_states[profile_name] != "enforce":
        raise ProfileModeMismatchError(profile_name, profile_states[profile_name])


def extract_json_payload(output: str) -> object:
    """Parse the widest JSON document from aa-status output with possible shell noise."""
    decoder = json.JSONDecoder()
    best_payload: object | None = None
    best_length = -1

    for index, char in enumerate(output):
        if char not in "[{":
            continue
        try:
            parsed, end = decoder.raw_decode(output[index:])
        except json.JSONDecodeError:
            continue
        if end > best_length:
            best_payload = parsed
            best_length = end

    if best_payload is None:
        raise JSONPayloadNotFoundError(output)
    return best_payload


def extract_profile_states(payload: object) -> dict[str, str]:
    """Validate aa-status payload shape and return profile->mode mapping."""
    expect(payload).to_match_schema(APPARMOR_STATUS_SCHEMA)

    if isinstance(payload, dict) and "profiles" in payload:
        expect(payload).to_match_schema(APPARMOR_WRAPPED_STATUS_SCHEMA)
        profiles = payload["profiles"]
        if is_profile_states_mapping(profiles):
            return dict(profiles)
        raise InvalidProfileStatesMappingError(profiles)

    expect(payload).to_match_schema(APPARMOR_BARE_STATUS_SCHEMA)
    if is_profile_states_mapping(payload):
        return dict(payload)
    raise InvalidProfileStatesMappingError(payload)


def is_profile_states_mapping(value: object) -> TypeGuard[dict[str, str]]:
    """Report whether a decoded JSON value is a profile-to-mode string mapping."""
    if not isinstance(value, dict):
        return False

    return all(isinstance(key, str) and isinstance(item, str) for key, item in value.items())


@pytest.mark.infra
@pytest.mark.AppArmor
def test_apparmor_reports_enabled(testsudo_cli_client: CLIClient) -> None:
    """Test case: AppArmor kernel feature is enabled on the system.

    Preparation:
    - The LiteNAS package has been deployed on a host with AppArmor runtime installed.
    - The restricted sudo test user can execute `aa-status` without a password.

    Action:
    - Run `aa-status --enabled` via sudo.

    Expected result:
    - The command exits successfully, indicating AppArmor is enabled.
    """
    testsudo_cli_client.execute("sudo -n aa-status --enabled")
    testsudo_cli_client.assert_exit_code(0)


@pytest.mark.infra
@pytest.mark.AppArmor
def test_apparmor_service_is_active(cli_client: CLIClient) -> None:
    """Test case: AppArmor userspace service is active.

    Preparation:
    - The LiteNAS package has been deployed on a host where AppArmor service is managed by systemd.

    Action:
    - Query systemd active state for `apparmor`.

    Expected result:
    - The AppArmor service reports `active`.
    """
    cli_client.execute("systemctl is-active apparmor")
    cli_client.assert_output_contains("active")


@pytest.mark.infra
@pytest.mark.AppArmor
@pytest.mark.parametrize("profile_name", APPARMOR_PROFILE_CASES)
def test_apparmor_profile_is_loaded(
    testsudo_cli_client: CLIClient,
    profile_name: str,
) -> None:
    """Test case: expected AppArmor profile is loaded.

    Preparation:
    - The LiteNAS package has been deployed with AppArmor profile files installed.
    - The restricted sudo test user can execute `aa-status` without a password.

    Action:
    - Query `aa-status` for the profile count using an exact profile filter.

    Expected result:
    - Exactly one matching loaded profile is present.
    """
    assert_profile_count(testsudo_cli_client, profile_name, enforce_mode_only=False)


@pytest.mark.infra
@pytest.mark.AppArmor
@pytest.mark.parametrize("profile_name", APPARMOR_PROFILE_CASES)
def test_apparmor_profile_is_in_enforce_mode(
    testsudo_cli_client: CLIClient,
    profile_name: str,
) -> None:
    """Test case: expected AppArmor profile is loaded in enforce mode.

    Preparation:
    - The LiteNAS package has been deployed with AppArmor profile files installed.
    - The restricted sudo test user can execute `aa-status` without a password.

    Action:
    - Query `aa-status` for the profile count using both profile and enforce-mode filters.

    Expected result:
    - Exactly one matching profile is reported in enforce mode.
    """
    assert_profile_count(testsudo_cli_client, profile_name, enforce_mode_only=True)
