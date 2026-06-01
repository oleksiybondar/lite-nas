"""System infrastructure test suite for installed LiteNAS package dependencies."""

import pytest
from constants import DEPENDENCY_PACKAGES
from hyperiontf import CLIClient

DEPENDENCY_PACKAGE_MARKS = {
    "acl": pytest.mark.ACL,
    "aide": pytest.mark.AIDE,
    "apparmor": pytest.mark.AppArmor,
    "postfix": pytest.mark.Postfix,
    "zfsutils-linux": pytest.mark.ZFS,
    "nginx": pytest.mark.Nginx,
    "nats-server": pytest.mark.NATS,
}

DEPENDENCY_PACKAGE_CASES = [
    pytest.param(package, marks=DEPENDENCY_PACKAGE_MARKS[package], id=package)
    for package in DEPENDENCY_PACKAGES
]


def assert_apt_package_is_installed(cli_client: CLIClient, package: str) -> None:
    """Verify that apt lists the requested package as installed."""
    cli_client.execute(f"apt list --installed {package} -o APT::Color=0 2>/dev/null", timeout=120)
    cli_client.assert_output_contains(f"{package}/")


@pytest.mark.Dependency
@pytest.mark.infra
@pytest.mark.parametrize("package", DEPENDENCY_PACKAGE_CASES)
def test_runtime_dependency_package_is_installed(
    cli_client: CLIClient,
    package: str,
) -> None:
    """Test case: LiteNAS runtime dependency package is installed.

    Preparation:
    - The LiteNAS Debian package has been installed with apt dependency resolution.

    Action:
    - Query apt for the parametrized installed dependency package entry.

    Expected result:
    - Apt lists the dependency package by name.
    """
    assert_apt_package_is_installed(cli_client, package)


@pytest.mark.Dependency
@pytest.mark.infra
def test_lite_nas_package_is_installed_via_apt(cli_client: CLIClient) -> None:
    """Test case: LiteNAS Debian package is visible as installed to apt.

    Preparation:
    - The LiteNAS Debian package has been installed on the target system.

    Action:
    - Query apt for the installed LiteNAS package entry.

    Expected result:
    - Apt reports the `lite-nas` package in its installed package list.
    """
    assert_apt_package_is_installed(cli_client, "lite-nas")
