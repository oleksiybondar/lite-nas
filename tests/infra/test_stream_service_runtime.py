"""Installed Stream Service account and transport boundary checks."""

import pytest
from hyperiontf import CLIClient

pytestmark = [pytest.mark.infra, pytest.mark.StreamService]


# Requirements: stream-service-svc/OR-001.
def test_stream_service_runs_as_dedicated_user(cli_client: CLIClient) -> None:
    """Test case: systemd starts Stream Service under its dedicated Unix account.

    Preparation:
    - The LiteNAS package has installed and started Stream Service.

    Action:
    - Read the systemd User property for its unit.

    Expected result:
    - The runtime account is the non-root stream-service account.
    """
    cli_client.execute("systemctl show -p User --value lite-nas-stream-service.service")
    cli_client.assert_output_contains("lite-nas-stream-service")


# Requirements: stream-service-svc/OR-001.
def test_stream_service_has_storage_read_group(cli_client: CLIClient) -> None:
    """Test case: the dedicated account joins the storage read group.

    Preparation:
    - Stream Service deployment has created its account and storage group.

    Action:
    - Query the account's supplementary groups.

    Expected result:
    - The group list includes `lite-nas-fs-rd`.
    """
    cli_client.execute("id -nG lite-nas-stream-service")
    cli_client.assert_output_contains("lite-nas-fs-rd")


# Requirements: stream-service-svc/FR-005 and OR-002.
def test_host_ffmpeg_provides_h264_encoder(cli_client: CLIClient) -> None:
    """Test case: the installed OS codec package supports browser video output.

    Preparation:
    - The LiteNAS package has installed its ffmpeg dependency.

    Action:
    - List available ffmpeg encoders.

    Expected result:
    - The host supplies the libx264 encoder used by Stream Service.
    """
    cli_client.execute("ffmpeg -hide_banner -encoders 2>/dev/null")
    cli_client.assert_output_contains("libx264")


# Requirements: stream-service-svc/FR-004 and OR-002.
def test_host_libheif_provides_heic_decoder(cli_client: CLIClient) -> None:
    """Test case: installed OS packages can decode tiled HEIC stills.

    Preparation:
    - The LiteNAS package has installed libheif-examples and its HEVC plugin.

    Action:
    - List available libheif decoders.

    Expected result:
    - The host supplies the libde265 decoder used for HEIC previews.
    """
    cli_client.execute("heif-dec --list-decoders")
    cli_client.assert_output_contains("libde265")
