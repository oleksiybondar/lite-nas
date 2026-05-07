"""Shared constants for LiteNAS Python system tests."""

import os

API_BASE_URL: str = os.environ.get("LITENAS_API_URL", "http://localhost")
UI_BASE_URL: str = os.environ.get("LITENAS_UI_URL", "http://localhost:9090")

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
