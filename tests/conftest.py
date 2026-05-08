"""Pytest plugin registration for LiteNAS system-test fixtures."""

pytest_plugins = (
    "fixtures.api_client",
    "fixtures.cli_client",
)
