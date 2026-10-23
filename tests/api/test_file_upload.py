"""Installed-system File Service upload checks through Nginx's /f mount."""

import hashlib
import shlex
from collections.abc import Generator
from http import HTTPStatus
from pwd import getpwnam
from urllib.parse import urlsplit
from uuid import uuid4

import pytest
from constants import (
    API_BASE_URL,
    CREDENTIALS,
    SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
)
from hyperiontf import CLIClient, RESTClient, expect

from .conftest import ACCESS_COOKIE

pytestmark = [pytest.mark.api, pytest.mark.FileService, pytest.mark.Nginx]

CONTENT = b"file-service-system-test\n"
ORIGIN = f"{urlsplit(API_BASE_URL).scheme}://{urlsplit(API_BASE_URL).netloc}"
UPLOAD_ENDPOINT = "/f/proprietary/upload"

CRITICAL_UPLOAD_DIRECTORIES = [
    pytest.param("/etc", id="system-configuration"),
    pytest.param("/etc/lite-nas", id="service-configuration"),
    pytest.param("/etc/lite-nas/certificates/transport", id="security-certificates"),
    pytest.param("/var/log/lite-nas", id="service-logs"),
    pytest.param("/var/lib/lite-nas", id="service-state"),
    pytest.param("/usr", id="installed-binaries"),
]


@pytest.fixture
def upload_target(operator_cli_client: CLIClient) -> Generator[str, None, None]:
    """Reserve a unique operator-home path and remove its upload artifacts later."""
    path = f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/.lite-nas-upload-{uuid4().hex}"
    try:
        yield path
    finally:
        operator_cli_client.execute(f"rm -f -- {shlex.quote(path)} {shlex.quote(path + '.upload')}")


def multipart_body(destination: str, *, file_first: bool = False) -> tuple[bytes, str]:
    """Encode the documented destination and file parts in the selected order."""
    boundary = f"lite-nas-{uuid4().hex}"
    destination_part = (
        f'--{boundary}\r\nContent-Disposition: form-data; name="destination"\r\n\r\n'
        f"{destination}\r\n"
    ).encode()
    file_part = (
        (
            f"--{boundary}\r\n"
            'Content-Disposition: form-data; name="file"; filename="sample.txt"\r\n'
            "Content-Type: text/plain\r\n\r\n"
        ).encode()
        + CONTENT
        + b"\r\n"
    )
    parts = (file_part, destination_part) if file_first else (destination_part, file_part)
    return b"".join(parts) + f"--{boundary}--\r\n".encode(), boundary


def upload_conventional(api_client: RESTClient, cookie: str, destination: str) -> None:
    """Publish the small fixture file through the destination-first multipart route."""
    body, boundary = multipart_body(destination)
    response = api_client.post(
        path="/f/upload",
        payload=body,
        headers={
            "Content-Type": f"multipart/form-data; boundary={boundary}",
            "Origin": ORIGIN,
        },
        cookies={ACCESS_COOKIE: cookie},
    )
    if response.status != HTTPStatus.CREATED:
        pytest.fail(f"conventional upload failed: HTTP {response.status}")


def start_upload(
    api_client: RESTClient, cookie: str, destination: str, content: bytes
) -> tuple[str, int]:
    """Create a resumable upload and return its ID and committed chunk size."""
    response = api_client.post(
        path=UPLOAD_ENDPOINT,
        payload={
            "destination": destination,
            "filename": "sample.txt",
            "size": len(content),
            "sha256": hashlib.sha256(content).hexdigest(),
        },
        headers={"Origin": ORIGIN, "Idempotency-Key": uuid4().hex},
        cookies={ACCESS_COOKIE: cookie},
    )
    if response.status != HTTPStatus.CREATED:
        pytest.fail(f"upload initialization failed: HTTP {response.status}")
    return str(response.body["upload_id"]), int(response.body["chunk_size"])


def cancel_upload(api_client: RESTClient, cookie: str, upload_id: str) -> None:
    """Cancel an incomplete upload to release its stage and metadata."""
    api_client.delete(
        path=f"{UPLOAD_ENDPOINT}/{upload_id}",
        headers={"Origin": ORIGIN},
        cookies={ACCESS_COOKIE: cookie},
    )


def configured_chunk_size(api_client: RESTClient, cookie: str, destination: str) -> int:
    """Probe the effective chunk size without assuming the installed setting."""
    upload_id, chunk_size = start_upload(api_client, cookie, destination, b"")
    cancel_upload(api_client, cookie, upload_id)
    return chunk_size


def varied_content(chunk_size: int) -> bytes:
    """Produce more than two chunks of nonrepeating printable test data."""
    return hashlib.shake_256(uuid4().bytes).hexdigest(chunk_size + 97).encode()


def send_chunk(
    api_client: RESTClient, cookie: str, upload_id: str, content: bytes, offset: int
) -> None:
    """Commit one ordered chunk or fail the fixture flow at its exact offset."""
    response = api_client.patch(
        path=f"{UPLOAD_ENDPOINT}/{upload_id}",
        payload=content,
        headers={
            "Origin": ORIGIN,
            "Upload-Offset": str(offset),
            "Content-Type": "application/octet-stream",
        },
        cookies={ACCESS_COOKIE: cookie},
    )
    if response.status != HTTPStatus.NO_CONTENT:
        pytest.fail(f"chunk at offset {offset} failed: HTTP {response.status}")


def complete_upload(api_client: RESTClient, cookie: str, upload_id: str) -> None:
    """Publish a fully committed upload or fail the fixture flow."""
    response = api_client.put(
        path=f"{UPLOAD_ENDPOINT}/{upload_id}/completion",
        headers={"Origin": ORIGIN},
        cookies={ACCESS_COOKIE: cookie},
    )
    if response.status != HTTPStatus.OK:
        pytest.fail(f"upload completion failed: HTTP {response.status}")


# Requirements: file-service-svc/FR-001 and file-service-svc/FR-005.
def test_anonymous_upload_is_denied(api_client: RESTClient, upload_target: str) -> None:
    """Test case: upload creation requires an authenticated session.

    Preparation:
    - A unique destination is available in the operator's home.

    Action:
    1. POST a small multipart file to /f/upload without an access token.

    Expected result:
    - The service rejects the request with 401 before processing the destination.
    """
    body, boundary = multipart_body(upload_target)
    response = api_client.post(
        path="/f/upload",
        payload=body,
        headers={"Content-Type": f"multipart/form-data; boundary={boundary}"},
    )
    expect(response.status).to_be(401)


# Requirements: file-service-svc/FR-004 and file-service-svc/FR-005.
def test_conventional_upload_publishes_file(
    api_client: RESTClient, operator_cookie: str, upload_target: str
) -> None:
    """Test case: a destination-first multipart upload publishes its bytes.

    Preparation:
    - The operator is logged in and owns the destination directory.

    Action:
    1. POST the destination field followed by a small text file to /f/upload.
    2. GET the published destination through /f.

    Expected result:
    - The published file contains the uploaded bytes.
    """
    upload_conventional(api_client, operator_cookie, upload_target)
    downloaded = api_client.get(path=f"/f{upload_target}", cookies={ACCESS_COOKIE: operator_cookie})
    expect(downloaded.raw_body).to_be(CONTENT.decode())


# Requirements: file-service-svc/FR-004 and file-service-svc/TR-002.
@pytest.mark.parametrize("upload_mode", ["conventional", "proprietary"])
def test_uploaded_file_has_operator_permissions(
    api_client: RESTClient,
    operator_cli_client: CLIClient,
    operator_cookie: str,
    upload_target: str,
    upload_mode: str,
) -> None:
    """Test case: either upload protocol publishes with the operator's private permissions.

    Preparation:
    - The operator is authenticated and has a unique destination in their home.

    Action:
    1. Upload a small file through the selected conventional or proprietary route.
    2. Inspect the published file's Unix UID, GID, and mode with stat.

    Expected result:
    - The UID and GID match the operator's Unix identity and the mode is 0600.
    """
    if upload_mode == "conventional":
        upload_conventional(api_client, operator_cookie, upload_target)
    else:
        upload_id, chunk_size = start_upload(api_client, operator_cookie, upload_target, CONTENT)
        finished = False
        try:
            for offset in range(0, len(CONTENT), chunk_size):
                send_chunk(
                    api_client,
                    operator_cookie,
                    upload_id,
                    CONTENT[offset : offset + chunk_size],
                    offset,
                )
            complete_upload(api_client, operator_cookie, upload_id)
            finished = True
        finally:
            if not finished:
                cancel_upload(api_client, operator_cookie, upload_id)

    operator_cli_client.execute(f"stat -c '%u:%g:%a' -- {shlex.quote(upload_target)}")
    operator_cli_client.assert_exit_code(0)
    operator = getpwnam(SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN)
    expect(operator_cli_client.output.splitlines()[-1]).to_be(
        f"{operator.pw_uid}:{operator.pw_gid}:600"
    )


# Requirements: file-service-svc/FR-005.
def test_conventional_upload_requires_destination_first(
    api_client: RESTClient, operator_cookie: str, upload_target: str
) -> None:
    """Test case: file bytes cannot precede destination authorization.

    Preparation:
    - The operator is logged in and has an available destination.

    Action:
    1. POST the file part before the destination field to /f/upload.

    Expected result:
    - The response reports destination_required_first.
    """
    body, boundary = multipart_body(upload_target, file_first=True)
    response = api_client.post(
        path="/f/upload",
        payload=body,
        headers={
            "Content-Type": f"multipart/form-data; boundary={boundary}",
            "Origin": ORIGIN,
        },
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    expect(response.body["error"]["code"]).to_be("destination_required_first")


# Requirements: file-service-svc/FR-005.
def test_resumable_upload_reports_initial_progress(
    api_client: RESTClient, operator_cookie: str, upload_target: str
) -> None:
    """Test case: a created upload reports its starting progress.

    Preparation:
    - The operator is logged in and has an available destination.

    Action:
    1. POST metadata to create a proprietary upload resource.
    2. GET the new resource before sending any file content.

    Expected result:
    - The resource reports zero acknowledged bytes.
    """
    upload_id, _ = start_upload(api_client, operator_cookie, upload_target, CONTENT)
    try:
        response = api_client.get(
            path=f"{UPLOAD_ENDPOINT}/{upload_id}",
            cookies={ACCESS_COOKIE: operator_cookie},
        )
        expect(response.body["acknowledged_bytes"]).to_be(0)
    finally:
        cancel_upload(api_client, operator_cookie, upload_id)


# Requirements: file-service-svc/FR-005.
def test_sequential_chunks_publish_exact_content(
    api_client: RESTClient, operator_cookie: str, upload_target: str
) -> None:
    """Test case: sequential chunks publish a varied multi-chunk file.

    Preparation:
    - The operator owns the destination directory and the effective chunk size is known.
    - The generated content spans more than two chunks and varies throughout.

    Action:
    1. Create a proprietary upload for the generated content.
    2. PATCH each chunk in increasing offset order, including the short final chunk.
    3. PUT completion and GET the published file.

    Expected result:
    - The published file exactly matches the generated content.
    """
    chunk_size = configured_chunk_size(api_client, operator_cookie, upload_target)
    content = varied_content(chunk_size)
    upload_id, _ = start_upload(api_client, operator_cookie, upload_target, content)
    finished = False
    try:
        for offset in range(0, len(content), chunk_size):
            send_chunk(
                api_client,
                operator_cookie,
                upload_id,
                content[offset : offset + chunk_size],
                offset,
            )
        complete_upload(api_client, operator_cookie, upload_id)
        finished = True
        downloaded = api_client.get(
            path=f"/f{upload_target}", cookies={ACCESS_COOKIE: operator_cookie}
        )
        expect(downloaded.raw_body).to_be(content.decode())
    finally:
        if not finished:
            cancel_upload(api_client, operator_cookie, upload_id)


# Requirements: file-service-svc/FR-005.
def test_incomplete_upload_cannot_complete(
    api_client: RESTClient, operator_cookie: str, upload_target: str
) -> None:
    """Test case: a partially written upload cannot be published.

    Preparation:
    - The operator initializes an upload containing more than two chunks.

    Action:
    1. PATCH only the first full chunk.
    2. PUT completion while the remaining chunks are missing.

    Expected result:
    - The response reports upload_incomplete; the upload is not published.
    """
    chunk_size = configured_chunk_size(api_client, operator_cookie, upload_target)
    content = varied_content(chunk_size)
    upload_id, _ = start_upload(api_client, operator_cookie, upload_target, content)
    try:
        send_chunk(api_client, operator_cookie, upload_id, content[:chunk_size], 0)
        response = api_client.put(
            path=f"{UPLOAD_ENDPOINT}/{upload_id}/completion",
            headers={"Origin": ORIGIN},
            cookies={ACCESS_COOKIE: operator_cookie},
        )
        expect(response.body["error"]["code"]).to_be("upload_incomplete")
    finally:
        cancel_upload(api_client, operator_cookie, upload_id)


# Requirements: file-service-svc/FR-005.
def test_paused_upload_resumes_from_committed_offset(
    api_client: RESTClient, operator_cookie: str, upload_target: str
) -> None:
    """Test case: a paused upload resumes without resending its first chunk.

    Preparation:
    - The operator has a new upload for a varied multi-chunk file.

    Action:
    1. PATCH the first chunk, then stop sending content.
    2. GET progress and resume PATCH requests at the acknowledged offset.
    3. PUT completion and GET the published file.

    Expected result:
    - The published file exactly matches the original content.
    """
    chunk_size = configured_chunk_size(api_client, operator_cookie, upload_target)
    content = varied_content(chunk_size)
    upload_id, _ = start_upload(api_client, operator_cookie, upload_target, content)
    finished = False
    try:
        send_chunk(api_client, operator_cookie, upload_id, content[:chunk_size], 0)
        progress = api_client.get(
            path=f"{UPLOAD_ENDPOINT}/{upload_id}",
            cookies={ACCESS_COOKIE: operator_cookie},
        )
        offset = int(progress.body["acknowledged_bytes"])
        if offset != chunk_size:
            pytest.fail(f"resumed offset is {offset}, expected {chunk_size}")
        for next_offset in range(offset, len(content), chunk_size):
            send_chunk(
                api_client,
                operator_cookie,
                upload_id,
                content[next_offset : next_offset + chunk_size],
                next_offset,
            )
        complete_upload(api_client, operator_cookie, upload_id)
        finished = True
        downloaded = api_client.get(
            path=f"/f{upload_target}", cookies={ACCESS_COOKIE: operator_cookie}
        )
        expect(downloaded.raw_body).to_be(content.decode())
    finally:
        if not finished:
            cancel_upload(api_client, operator_cookie, upload_id)


def denied_upload_status(api_client: RESTClient, cookie: str, directory: str) -> int:
    """Attempt initialization at a unique critical path and cancel any unexpected stage."""
    destination = f"{directory}/.lite-nas-denied-{uuid4().hex}"
    response = api_client.post(
        path=UPLOAD_ENDPOINT,
        payload={
            "destination": destination,
            "filename": "sample.txt",
            "size": len(CONTENT),
            "sha256": hashlib.sha256(CONTENT).hexdigest(),
        },
        headers={"Origin": ORIGIN, "Idempotency-Key": uuid4().hex},
        cookies={ACCESS_COOKIE: cookie},
    )
    if response.status == HTTPStatus.CREATED:
        cancel_upload(api_client, cookie, str(response.body["upload_id"]))
    return int(response.status)


# Requirements: file-service-svc/FR-004.
def test_operator_cannot_upload_into_another_users_home(
    api_client: RESTClient, operator_cookie: str
) -> None:
    """Test case: an operator cannot create a file in another user's home.

    Preparation:
    - The operator is authenticated and the standard test user's home exists.

    Action:
    1. POST proprietary upload metadata for a unique file in that home.

    Expected result:
    - File Service denies creation with 403.
    """
    status = denied_upload_status(api_client, operator_cookie, f"/home/{CREDENTIALS['login']}")
    expect(status).to_be(HTTPStatus.FORBIDDEN)


# Requirements: file-service-svc/FR-004.
@pytest.mark.parametrize("directory", CRITICAL_UPLOAD_DIRECTORIES)
def test_operator_cannot_upload_into_critical_path(
    api_client: RESTClient, operator_cookie: str, directory: str
) -> None:
    """Test case: an operator cannot upload into a critical system tree.

    Preparation:
    - The operator is authenticated and the installed critical directory exists.

    Action:
    1. POST proprietary upload metadata for a unique path in the critical tree.

    Expected result:
    - File Service denies creation with 403.
    """
    status = denied_upload_status(api_client, operator_cookie, directory)
    expect(status).to_be(HTTPStatus.FORBIDDEN)


# Requirements: file-service-svc/FR-004.
@pytest.mark.parametrize("directory", CRITICAL_UPLOAD_DIRECTORIES)
def test_admin_cannot_upload_into_critical_path(
    api_client: RESTClient, admin_cookie: str, directory: str
) -> None:
    """Test case: admin read access does not permit uploads to critical paths.

    Preparation:
    - The configured admin is authenticated and can read the installed critical files.

    Action:
    1. POST proprietary upload metadata for a unique path in the critical tree.

    Expected result:
    - File Service denies creation with 403 despite the admin's read access.
    """
    status = denied_upload_status(api_client, admin_cookie, directory)
    expect(status).to_be(HTTPStatus.FORBIDDEN)


# Requirements: file-service-svc/FR-001 and file-service-svc/FR-005.
@pytest.mark.parametrize(
    "method",
    ["GET", "HEAD", "PATCH", "DELETE", "PUT"],
)
def test_other_user_cannot_access_resumable_upload(
    api_client: RESTClient,
    operator_cookie: str,
    testsudo_cookie: str,
    upload_target: str,
    method: str,
) -> None:
    """Test case: an upload ID does not grant another user access to the upload.

    Preparation:
    - The operator creates a resumable upload with a known ID.
    - A second account has its own valid session.

    Action:
    - The second account sends the selected progress or mutation request.

    Expected result:
    - The upload resource remains hidden from the second account.
    """
    upload_id, _ = start_upload(api_client, operator_cookie, upload_target, CONTENT)
    try:
        endpoint = f"{UPLOAD_ENDPOINT}/{upload_id}"
        if method == "PUT":
            endpoint += "/completion"
        response = api_client.request(
            method=method,
            path=endpoint,
            payload=b"x" if method == "PATCH" else None,
            headers={"Origin": ORIGIN, "Upload-Offset": "0"},
            cookies={ACCESS_COOKIE: testsudo_cookie},
        ).execute()
        expect(response.status).to_be(HTTPStatus.NOT_FOUND)
    finally:
        cancel_upload(api_client, operator_cookie, upload_id)
