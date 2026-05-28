"""System infrastructure test suite for deployed certificate path permissions."""

import pytest
from hyperiontf import CLIClient

CERT_PERMISSION_CASES = [
    pytest.param(
        "/etc/lite-nas/certificates",
        "root:lite-nas",
        "711",
        False,
        id="certificates-dir",
    ),
    pytest.param(
        "/etc/lite-nas/certificates/transport",
        "root:lite-nas",
        "711",
        False,
        id="transport-dir",
    ),
    pytest.param(
        "/etc/lite-nas/certificates/transport/root-ca.crt",
        "root:root",
        "644",
        False,
        id="transport-root-ca",
    ),
    pytest.param(
        "/etc/lite-nas/certificates/auth",
        "root:lite-nas",
        "750",
        True,
        id="auth-cert-dir",
    ),
    pytest.param(
        "/etc/lite-nas/certificates/auth/token-signing.crt",
        "root:lite-nas",
        "640",
        True,
        id="token-signing-cert",
    ),
    pytest.param(
        "/etc/lite-nas/certificates/auth/token-signing.key",
        "root:root",
        "600",
        True,
        id="token-signing-key",
    ),
]


def assert_path_owner_and_mode(
    cli_client: CLIClient,
    path: str,
    expected_owner_group: str,
    expected_mode: str,
) -> None:
    """Verify a certificate path exists and matches expected owner/group and mode."""
    cli_client.execute(
        f"if test -e '{path}'; then "
        f"actual=\"$(stat -c '%U:%G %a' '{path}')\"; "
        f"if [ \"$actual\" = '{expected_owner_group} {expected_mode}' ]; then "
        f"echo '__OK__ {path}'; "
        "else "
        'echo "__BAD__ $actual"; '
        "fi; "
        "else "
        "echo '__MISSING__'; "
        "fi"
    )
    cli_client.assert_output_contains(f"__OK__ {path}")


def assert_path_not_accessible(cli_client: CLIClient, path: str) -> None:
    """Verify a sensitive path is not discoverable to an unprivileged shell user."""
    cli_client.execute(
        f"if test -e '{path}'; then "
        f"echo '__UNEXPECTED_ACCESS__ {path}'; "
        "else "
        f"echo '__NO_ACCESS__ {path}'; "
        "fi"
    )
    cli_client.assert_output_contains(f"__NO_ACCESS__ {path}")


@pytest.mark.infra
@pytest.mark.parametrize(
    ("path", "owner_group", "mode", "requires_privileged_metadata"),
    CERT_PERMISSION_CASES,
)
def test_certificate_paths_have_expected_permissions(
    cli_client: CLIClient,
    path: str,
    owner_group: str,
    mode: str,
    requires_privileged_metadata: bool,
) -> None:
    """Test case: deployed LiteNAS certificate paths have expected permissions.

    Preparation:
    - The LiteNAS Debian package has been deployed on the target host.
    - Certificate-rotation and config-deployment scripts completed successfully.

    Action:
    - Query owner/group and numeric mode for the parametrized certificate path.

    Expected result:
    - The path exists and matches the expected owner/group and mode.
    """
    if requires_privileged_metadata:
        assert_path_not_accessible(cli_client, path)
        return

    assert_path_owner_and_mode(cli_client, path, owner_group, mode)
