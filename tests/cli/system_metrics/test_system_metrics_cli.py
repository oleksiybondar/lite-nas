"""System CLI test suite for the installed system-metrics CLI."""

import pytest
from constants import SYSTEM_METRICS_CLI_BINARY
from hyperiontf import CLIClient

SYSTEM_METRICS_CLI_SECTION_CASES = [
    pytest.param("--cpu", "CPU Load:", id="cpu"),
    pytest.param("--ram", "RAM:", id="ram"),
]


@pytest.mark.SystemMetrics
@pytest.mark.cli
@pytest.mark.parametrize(("flag", "expectation"), SYSTEM_METRICS_CLI_SECTION_CASES)
def test_system_metrics_cli_prints_selected_section(
    cli_client: CLIClient,
    flag: str,
    expectation: str,
) -> None:
    """Test case: system-metrics CLI prints the selected current snapshot section.

    Preparation:
    - The LiteNAS system-metrics service and NATS runtime dependency are running.
    - The installed system-metrics CLI can read its runtime configuration.

    Action:
    - Run the system-metrics CLI with the parametrized section flag.

    Expected result:
    - The command succeeds after producing the selected current-snapshot section.
    """
    cli_client.execute(f"{SYSTEM_METRICS_CLI_BINARY} {flag}")
    cli_client.assert_exit_code(0)
    cli_client.assert_output_contains(expectation)
