"""Shared constants for LiteNAS Python system tests."""

import os

API_BASE_URL: str = os.environ.get("LITENAS_API_URL", "https://localhost")
UI_BASE_URL: str = os.environ.get("LITENAS_UI_URL", "https://localhost")

UI_BROWSER_CAPS: dict[str, object] = {
    "automation": os.environ.get("LITENAS_UI_AUTOMATION", "playwright"),
    "browser": os.environ.get("LITENAS_UI_BROWSER", "chrome"),
    "headless": os.environ.get("LITENAS_UI_HEADLESS", "true").lower() == "true",
}

CREDENTIALS: dict[str, str] = {
    "login": os.environ.get("LITENAS_API_LOGIN", "testuser"),
    "password": os.environ.get("LITENAS_API_PASSWORD", "testpassword"),
}

DEPENDENCY_PACKAGES: list[str] = [
    "acl",
    "aide",
    "apparmor",
    "postfix",
    "sudo",
    "zfsutils-linux",
    "nginx",
    "nats-server",
]

SYSTEM_METRICS_CLI_BINARY: str = os.environ.get(
    "LITENAS_SYSTEM_METRICS_CLI_BINARY",
    "/usr/bin/system-metrics-cli",
)
SYSTEM_LOGGING_MANAGER_CLI_BINARY: str = os.environ.get(
    "LITENAS_SYSTEM_LOGGING_MANAGER_CLI_BINARY",
    "/usr/bin/system-logging-manager-cli",
)
SECURITY_LOGGING_MANAGER_CLI_BINARY: str = os.environ.get(
    "LITENAS_SECURITY_LOGGING_MANAGER_CLI_BINARY",
    "/usr/bin/security-logging-manager-cli",
)
SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN: str = os.environ.get(
    "LITENAS_OPERATOR_LOGIN",
    "testoperator",
)
SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD: str = os.environ.get(
    "LITENAS_OPERATOR_PASSWORD",
    os.environ.get("LITENAS_TEST_PASSWORD", "testpassword"),
)
SECURITY_LOGGING_MANAGER_LOGIN: str = os.environ.get(
    "LITENAS_SECURITY_LOGIN",
    "testsecurity",
)
SECURITY_LOGGING_MANAGER_PASSWORD: str = os.environ.get(
    "LITENAS_SECURITY_PASSWORD",
    os.environ.get("LITENAS_TEST_PASSWORD", "testpassword"),
)
TEST_SUDO_LOGIN: str = os.environ.get("LITENAS_TESTSUDO_LOGIN", "testsudouser")
TEST_SUDO_PASSWORD: str = os.environ.get(
    "LITENAS_TESTSUDO_PASSWORD",
    os.environ.get("LITENAS_TEST_PASSWORD", "testpassword"),
)

SYSTEMD_SERVICES: list[str] = [
    "lite-nas-auth",
    "lite-nas-web-gateway",
    "lite-nas-system-metrics",
    "lite-nas-system-logging-manager",
    "lite-nas-security-logging-manager",
    "lite-nas-system-email-notifier",
    "lite-nas-security-email-notifier",
    "postfix",
    "apparmor",
    "nginx",
    "nats-server",
]
