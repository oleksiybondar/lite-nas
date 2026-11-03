"""Installed backup and restore JSON payload checks through HyperionTF."""

from collections.abc import Generator
from http import HTTPStatus
from uuid import uuid4

import pytest
from hyperiontf import RESTClient, expect

pytestmark = [pytest.mark.api, pytest.mark.WebGateway, pytest.mark.BackupRestore]

APPLICATION = "lite-nas"
BASE = "/api/backup-restore"
COOKIE = "lite-nas-at"
SESSION_ID = "0" * 32


@pytest.fixture
def named_backup(
    api_client: RESTClient, admin_cookie: str
) -> Generator[tuple[str, str, str], None, None]:
    """Create one named LiteNAS archive and delete that exact version after the case."""
    name = f"payload-{uuid4().hex}"
    description = f"system-payload-{uuid4().hex}"
    response = api_client.post(
        path=f"{BASE}/applications/{APPLICATION}/backups",
        payload={"name": name, "description": description},
        cookies={COOKIE: admin_cookie},
    )
    if response.status != HTTPStatus.OK:
        pytest.fail(f"test backup creation failed: HTTP {response.status}")
    version = str(response.body["data"]["version"])
    try:
        yield version, name, description
    finally:
        deleted = api_client.delete(
            path=f"{BASE}/applications/{APPLICATION}/versions/{version}",
            cookies={COOKIE: admin_cookie},
        )
        if deleted.status != HTTPStatus.OK:
            pytest.fail(f"test backup cleanup failed: HTTP {deleted.status}")


# Requirements: backup-restore/FR-001, backup-restore/IR-004.
def test_backup_json_labels_are_stored(
    api_client: RESTClient, admin_cookie: str, named_backup: tuple[str, str, str]
) -> None:
    """Test case: backup JSON labels survive into archive metadata.

    Preparation:
    - An administrator creates a disposable LiteNAS backup with a unique name and description.

    Action:
    - Read that version's archive metadata through the gateway.

    Expected result:
    - The stored name and description equal the submitted JSON values.
    """
    version, name, description = named_backup
    response = api_client.get(
        path=f"{BASE}/applications/{APPLICATION}/versions/{version}",
        cookies={COOKIE: admin_cookie},
    )
    if response.status != HTTPStatus.OK:
        pytest.fail(f"test backup metadata read failed: HTTP {response.status}")
    data = response.body["data"]
    expect((data.get("name"), data.get("description"))).to_be((name, description))


# Requirements: backup-restore/FR-012, backup-restore/IR-004.
def test_restore_all_by_name_skips_unmatched_archives(
    api_client: RESTClient, admin_cookie: str
) -> None:
    """Test case: a valid named restore body reaches the service without changing files.

    Preparation:
    - At least one backup application is configured and a unique unmatched name is chosen.

    Action:
    - Request restore all by that name in merge mode.

    Expected result:
    - Every configured application is reported as skipped.
    """
    response = api_client.post(
        path=f"{BASE}/restore-by-name",
        payload={"name": f"no-match-{uuid4().hex}", "mode": "merge"},
        cookies={COOKIE: admin_cookie},
        request_timeout=60,
    )
    if response.status != HTTPStatus.OK:
        pytest.fail(f"unmatched named restore failed: HTTP {response.status}")
    results = response.body["data"]["results"]
    if not results:
        pytest.fail("named restore returned no configured applications")
    expect({item["status"] for item in results}).to_be({"skipped"})


# Requirements: backup-restore/IR-004.
@pytest.mark.parametrize(
    ("path", "payload"),
    [
        pytest.param(
            f"{BASE}/applications/{APPLICATION}/versions/unknown/restore",
            {"mode": "invalid"},
            id="restore-mode",
        ),
        pytest.param(
            f"{BASE}/applications/{APPLICATION}/restore-by-name",
            {"name": "", "mode": "replace"},
            id="single-restore-name",
        ),
        pytest.param(
            f"{BASE}/restore-by-name",
            {"name": "", "mode": "replace"},
            id="all-restore-name",
        ),
        pytest.param(
            f"{BASE}/merge-sessions/{SESSION_ID}/save",
            {"name": "x" * 129, "description": "note"},
            id="merge-save-name",
        ),
    ],
)
def test_invalid_backup_post_payload_is_rejected(
    api_client: RESTClient, admin_cookie: str, path: str, payload: dict[str, str]
) -> None:
    """Test case: malformed backup and restore JSON is rejected before service work.

    Preparation:
    - An administrator is authenticated; no restore or merge session is started.

    Action:
    - Submit one invalid JSON field to the selected backup POST route.

    Expected result:
    - The gateway returns a request validation error.
    """
    response = api_client.post(path=path, payload=payload, cookies={COOKIE: admin_cookie})
    expect(response.status).to_be(HTTPStatus.UNPROCESSABLE_ENTITY)


# Requirements: backup-restore/IR-004.
def test_invalid_merge_write_payload_is_rejected(api_client: RESTClient, admin_cookie: str) -> None:
    """Test case: an invalid merge content body is rejected before staging.

    Preparation:
    - An administrator is authenticated; the session ID is not active.

    Action:
    - Upload a content chunk with an empty selected path.

    Expected result:
    - The gateway rejects the JSON payload.
    """
    response = api_client.put(
        path=f"{BASE}/merge-sessions/{SESSION_ID}/content",
        payload={"path": "", "offset": 0, "data": "YQ==", "final": False},
        cookies={COOKIE: admin_cookie},
    )
    expect(response.status).to_be(HTTPStatus.UNPROCESSABLE_ENTITY)
