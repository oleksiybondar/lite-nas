"""Shared constants for LiteNAS Python system tests."""

import os

API_BASE_URL: str = os.environ.get("LITENAS_API_URL", "http://localhost")

CREDENTIALS: dict[str, str] = {
    "login": os.environ.get("LITENAS_API_LOGIN", "testuser"),
    "password": os.environ.get("LITENAS_API_PASSWORD", "testpassword"),
}

DEPENDENCY_PACKAGES: list[str] = [
    "zfsutils-linux",
    "nginx",
    "nats-server",
]

SYSTEMD_SERVICES: list[str] = [
    "lite-nas-auth",
    "lite-nas-web-gateway",
    "lite-nas-system-metrics",
    "nginx",
    "nats-server",
]
