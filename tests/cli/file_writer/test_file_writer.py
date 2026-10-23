"""Installed-system CLI tests for the user-scoped file writer."""

import os
import shlex
from collections.abc import Generator

import pytest
from constants import FILE_WRITER_BINARY
from hyperiontf import CLIClient, expect

pytestmark = [
    pytest.mark.FileWriter,
    pytest.mark.cli,
    pytest.mark.usefixtures("writer_stage"),
]
ZFS_WRITER_TEST_ROOT = os.getenv("LITENAS_FILE_WRITER_ZFS_TEST_DIR", "")


@pytest.fixture
def writer_stage(testsudo_cli_client: CLIClient) -> Generator[None, None, None]:
    """Create an owned stage in /home and expose its identity to the shell session."""
    client = testsudo_cli_client
    client.execute('writer_test_dir=$(mktemp -d "$HOME/.file-writer-test.XXXXXX")')
    client.assert_exit_code(0)
    client.execute(f'{FILE_WRITER_BINARY} create --target "$writer_test_dir/target.upload"')
    client.assert_exit_code(0)
    client.execute(
        'writer_identity_args="--expected-device '
        '$(stat -c %d "$writer_test_dir/target.upload") '
        '--expected-inode $(stat -c %i "$writer_test_dir/target.upload") '
        '--expected-uid $(stat -c %u "$writer_test_dir/target.upload")"'
    )
    client.assert_exit_code(0)
    yield
    client.execute('rm -rf -- "$writer_test_dir"')


# Requirements: file-writer/FR-005.
@pytest.mark.parametrize("arguments", ["", " --help"], ids=["no-arguments", "help-flag"])
def test_writer_displays_help(testsudo_cli_client: CLIClient, arguments: str) -> None:
    """Test case: the installed writer explains its operations.

    Preparation:
    - The writer is installed on a LiteNAS host.

    Action:
    - Invoke it without arguments or with the help flag.

    Expected result:
    - It succeeds and prints readable operation usage.
    """
    testsudo_cli_client.execute(f"{FILE_WRITER_BINARY}{arguments}")
    testsudo_cli_client.assert_exit_code(0)
    testsudo_cli_client.assert_output_contains("Usage:")


# Requirements: file-writer/FR-001.
def test_create_rejects_existing_stage(testsudo_cli_client: CLIClient) -> None:
    """Test case: exclusive creation refuses a conflicting stage.

    Preparation:
    - The writer created an owned upload stage.

    Action:
    - Request creation at the same path again.

    Expected result:
    - The writer rejects the conflict.
    """
    testsudo_cli_client.execute(
        f'{FILE_WRITER_BINARY} create --target "$writer_test_dir/target.upload"'
    )
    expect(testsudo_cli_client.exit_code).not_to_be(0)


# Requirements: file-writer/FR-002.
def test_write_binary_chunk(testsudo_cli_client: CLIClient) -> None:
    """Test case: write raw bytes at the committed offset.

    Preparation:
    - An owned stage and its recorded identity exist.

    Action:
    - Pipe three binary bytes into the writer.

    Expected result:
    - The stage contains those exact bytes.
    """
    testsudo_cli_client.execute(
        f'printf "\\000\\377A" | {FILE_WRITER_BINARY} write '
        '--target "$writer_test_dir/target.upload" $writer_identity_args --offset 0 --length 3'
    )
    testsudo_cli_client.assert_exit_code(0)
    testsudo_cli_client.execute('printf "\\000\\377A" | cmp - "$writer_test_dir/target.upload"')
    testsudo_cli_client.assert_exit_code(0)


# Requirements: file-writer/FR-002.
def test_retry_replaces_partial_tail(testsudo_cli_client: CLIClient) -> None:
    """Test case: a retry replaces an uncommitted stage tail.

    Preparation:
    - The stage contains a committed prefix and an incomplete tail.

    Action:
    - Write a replacement chunk at the committed offset.

    Expected result:
    - The stage contains only the prefix and replacement bytes.
    """
    testsudo_cli_client.execute('printf "ABCxy" > "$writer_test_dir/target.upload"')
    testsudo_cli_client.execute(
        f'printf "DEFG" | {FILE_WRITER_BINARY} write '
        '--target "$writer_test_dir/target.upload" $writer_identity_args --offset 3 --length 4'
    )
    testsudo_cli_client.assert_exit_code(0)
    testsudo_cli_client.execute('printf "ABCDEFG" | cmp - "$writer_test_dir/target.upload"')
    testsudo_cli_client.assert_exit_code(0)


# Requirements: file-writer/FR-002.
def test_short_input_restores_committed_prefix(testsudo_cli_client: CLIClient) -> None:
    """Test case: incomplete input does not leave a partial append.

    Preparation:
    - The stage contains a committed three-byte prefix.

    Action:
    - Supply one byte while declaring two bytes.

    Expected result:
    - The writer fails and the committed prefix remains intact.
    """
    testsudo_cli_client.execute('printf "ABC" > "$writer_test_dir/target.upload"')
    testsudo_cli_client.execute(
        f'printf "D" | {FILE_WRITER_BINARY} write '
        '--target "$writer_test_dir/target.upload" $writer_identity_args --offset 3 --length 2'
    )
    expect(testsudo_cli_client.exit_code).not_to_be(0)
    testsudo_cli_client.execute('printf "ABC" | cmp - "$writer_test_dir/target.upload"')
    testsudo_cli_client.assert_exit_code(0)


# Requirements: file-writer/FR-003.
def test_wrong_stage_identity_is_rejected(testsudo_cli_client: CLIClient) -> None:
    """Test case: mismatched stage identity blocks cleanup.

    Preparation:
    - An owned stage and its actual identity exist.

    Action:
    - Request cleanup with UID zero as the expected owner.

    Expected result:
    - The writer rejects the request and preserves the stage.
    """
    testsudo_cli_client.execute(
        f'{FILE_WRITER_BINARY} cleanup --target "$writer_test_dir/target.upload" '
        '--expected-device "$(stat -c %d "$writer_test_dir/target.upload")" '
        '--expected-inode "$(stat -c %i "$writer_test_dir/target.upload")" --expected-uid 0'
    )
    expect(testsudo_cli_client.exit_code).not_to_be(0)
    testsudo_cli_client.execute('test -f "$writer_test_dir/target.upload"')
    testsudo_cli_client.assert_exit_code(0)


# Requirements: file-writer/FR-004.
def test_checksum_reports_staged_bytes(testsudo_cli_client: CLIClient) -> None:
    """Test case: the writer reports the stage SHA-256 digest.

    Preparation:
    - An owned stage contains three known bytes.

    Action:
    - Request a staged checksum.

    Expected result:
    - The reported digest matches those bytes.
    """
    testsudo_cli_client.execute('printf "ABC" > "$writer_test_dir/target.upload"')
    testsudo_cli_client.execute(
        f'{FILE_WRITER_BINARY} checksum --target "$writer_test_dir/target.upload" '
        "$writer_identity_args"
    )
    testsudo_cli_client.assert_exit_code(0)
    testsudo_cli_client.assert_output_contains(
        '"sha256":"b5d4045c3f466fa91fe2cc6abe79232a1a57cdf104f7a26e716e0a1e2789df78"'
    )


# Requirements: file-writer/FR-006.
def test_publish_new_destination(testsudo_cli_client: CLIClient) -> None:
    """Test case: publication exposes the stage as the completed file.

    Preparation:
    - An owned stage contains a complete upload and the destination is absent.

    Action:
    - Publish with the absent-destination expectation.

    Expected result:
    - The completed file contains the uploaded bytes.
    """
    testsudo_cli_client.execute('printf "ABC" > "$writer_test_dir/target.upload"')
    testsudo_cli_client.execute(
        f'{FILE_WRITER_BINARY} publish --target "$writer_test_dir/target.upload" '
        "$writer_identity_args --destination-absent"
    )
    testsudo_cli_client.assert_exit_code(0)
    testsudo_cli_client.execute('printf "ABC" | cmp - "$writer_test_dir/target"')
    testsudo_cli_client.assert_exit_code(0)


# Requirements: file-writer/FR-006, file-writer/FR-007.
def test_failed_publish_reports_cleanup(testsudo_cli_client: CLIClient) -> None:
    """Test case: a destination conflict cleans the stage.

    Preparation:
    - An owned stage and an existing destination are present.

    Action:
    - Publish while claiming the destination is absent.

    Expected result:
    - The failure reports successful stage cleanup.
    """
    testsudo_cli_client.execute('printf "old" > "$writer_test_dir/target"')
    testsudo_cli_client.execute(
        f'{FILE_WRITER_BINARY} publish --target "$writer_test_dir/target.upload" '
        "$writer_identity_args --destination-absent"
    )
    expect(testsudo_cli_client.exit_code).not_to_be(0)
    testsudo_cli_client.assert_output_contains('"cleanup":"succeeded"')


# Requirements: file-writer/IR-002.
def test_apparmor_denies_stage_creation_outside_write_root(
    testsudo_cli_client: CLIClient,
) -> None:
    """Test case: AppArmor confines the writer to provisioned write roots.

    Preparation:
    - The profile is enforced and /tmp is writable to the test user.

    Action:
    - Ask the writer to create a stage in /tmp.

    Expected result:
    - The writer exits with an access error.
    """
    testsudo_cli_client.execute(
        f'{FILE_WRITER_BINARY} create --target "/tmp/file-writer-denied-$$.upload"'
    )
    expect(testsudo_cli_client.exit_code).not_to_be(0)


# Requirements: file-writer/IR-001.
def test_non_service_user_cannot_sudo_writer(testsudo_cli_client: CLIClient) -> None:
    """Test case: a non-service user cannot run the writer through sudo.

    Preparation:
    - The restricted writer sudoers rule is installed.

    Action:
    - The system-test user requests the fixed writer executable through sudo.

    Expected result:
    - Sudo denies the invocation.
    """
    testsudo_cli_client.execute(
        'sudo -n -u "$USER" -- /usr/libexec/lite-nas/lite-nas-file-writer --help'
    )
    expect(testsudo_cli_client.exit_code).not_to_be(0)


# Requirements: file-writer/IR-002.
@pytest.mark.skipif(not ZFS_WRITER_TEST_ROOT, reason="No writable ZFS test root was configured")
def test_configured_zfs_write_root(testsudo_cli_client: CLIClient) -> None:
    """Test case: a configured ZFS write root permits an upload stage.

    Preparation:
    - A writable ZFS directory and a matching AppArmor local rule are configured.

    Action:
    - Create, write, and publish a stage in that directory.

    Expected result:
    - The completed file contains the uploaded bytes.
    """
    testsudo_cli_client.execute(
        "writer_zfs_dir=$(mktemp -d "
        f"{shlex.quote(ZFS_WRITER_TEST_ROOT)}/.file-writer-test.XXXXXX)"
    )
    testsudo_cli_client.assert_exit_code(0)
    try:
        testsudo_cli_client.execute(
            f'{FILE_WRITER_BINARY} create --target "$writer_zfs_dir/target.upload"'
        )
        testsudo_cli_client.assert_exit_code(0)
        testsudo_cli_client.execute(
            'writer_zfs_identity="--expected-device '
            '$(stat -c %d "$writer_zfs_dir/target.upload") '
            '--expected-inode $(stat -c %i "$writer_zfs_dir/target.upload") '
            '--expected-uid $(stat -c %u "$writer_zfs_dir/target.upload")"'
        )
        testsudo_cli_client.execute(
            f'printf "ZFS" | {FILE_WRITER_BINARY} write '
            '--target "$writer_zfs_dir/target.upload" $writer_zfs_identity --offset 0 --length 3'
        )
        testsudo_cli_client.assert_exit_code(0)
        testsudo_cli_client.execute(
            f'{FILE_WRITER_BINARY} publish --target "$writer_zfs_dir/target.upload" '
            "$writer_zfs_identity --destination-absent"
        )
        testsudo_cli_client.assert_exit_code(0)
        testsudo_cli_client.execute('printf "ZFS" | cmp - "$writer_zfs_dir/target"')
        testsudo_cli_client.assert_exit_code(0)
    finally:
        testsudo_cli_client.execute('rm -rf -- "$writer_zfs_dir"')
