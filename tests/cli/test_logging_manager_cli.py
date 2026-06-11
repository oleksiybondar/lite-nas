"""System CLI test suite for logging-manager JSON workflows."""

import json
import re
import shlex
import time
from dataclasses import dataclass

import pytest
from constants import (
    SECURITY_LOGGING_MANAGER_CLI_BINARY,
    SECURITY_LOGGING_MANAGER_LOGIN,
    SYSTEM_LOGGING_MANAGER_CLI_BINARY,
    SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
)
from hyperiontf import CLIClient, expect
from schemas.system_logging_manager import EVENT_ITEM_SCHEMA, EVENT_LIST_SCHEMA
from test_data.logging_manager_cli import (
    GET_COMMANDS,
    OCCURRENCE_VALUE_TEXT,
    OCCURRENCE_VALUE_TYPE,
    UNSUPPORTED_FILTER_ARGUMENT,
    UNSUPPORTED_FILTER_ERROR,
    UPDATED_MESSAGE,
    UPDATED_STATUS,
    acknowledge_event_payload,
    create_event_payload,
    create_occurrence_payload,
    mute_event_payload,
    update_event_state_payload,
)


@dataclass(frozen=True)
class LoggingManagerCLIContext:
    """Carry fixture and command context for shared logging-manager CLI test variants."""

    fixture_name: str
    cli_binary: str
    actor_login: str


pytestmark = [
    pytest.mark.SystemLoggingManager,
    pytest.mark.cli,
    pytest.mark.parametrize(
        "logging_manager_cli_context",
        [
            pytest.param(
                LoggingManagerCLIContext(
                    fixture_name="operator_cli_client",
                    cli_binary=SYSTEM_LOGGING_MANAGER_CLI_BINARY,
                    actor_login=SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
                ),
                id="system-logging-manager-cli",
            ),
            pytest.param(
                LoggingManagerCLIContext(
                    fixture_name="security_cli_client",
                    cli_binary=SECURITY_LOGGING_MANAGER_CLI_BINARY,
                    actor_login=SECURITY_LOGGING_MANAGER_LOGIN,
                ),
                id="security-logging-manager-cli",
            ),
        ],
        indirect=True,
    ),
]


class JSONOutputEmptyError(AssertionError):
    """Raised when a command expected to return JSON produces empty output."""

    def __init__(self, output: str) -> None:
        super().__init__(f"Expected JSON output, got empty output: {output!r}")


class JSONOutputInvalidError(AssertionError):
    """Raised when a command expected to return JSON produces non-JSON output."""

    def __init__(self, output: str) -> None:
        super().__init__(f"Expected JSON output, got non-JSON payload: {output!r}")


class EventNotReturnedError(AssertionError):
    """Raised when getEvent does not return the requested event."""

    def __init__(self, event_id: str) -> None:
        super().__init__(f"Event {event_id!r} was not returned by getEvent.")


ANSI_ESCAPE_RE = re.compile(r"\x1b\[[0-?]*[ -/]*[@-~]")


def execute_logging_manager_json(
    operator_cli_client: CLIClient,
    cli_binary: str,
    command: str,
    *,
    expect_success: bool = True,
) -> object:
    """Run get-command in JSON mode and parse JSON output."""
    # Print a separator newline first so command echo and JSON payload do not merge.
    operator_cli_client.execute(f"{cli_binary} {command} --json")
    if expect_success:
        operator_cli_client.assert_exit_code(0)
        if not operator_cli_client.output.strip():
            return {}
        return output_to_json(operator_cli_client)

    expect(operator_cli_client.exit_code).not_to_be(0)
    if not operator_cli_client.output.strip():
        return {}
    return operator_cli_client.output


def execute_logging_manager_mutation(
    operator_cli_client: CLIClient,
    cli_binary: str,
    command: str,
) -> None:
    """Run mutation command without JSON mode and require successful exit."""
    operator_cli_client.execute(f"{cli_binary} {command}")
    operator_cli_client.assert_exit_code(0)


def output_to_json(operator_cli_client: CLIClient) -> object:
    output = operator_cli_client.output
    json_fragment = extract_json_fragment(strip_terminal_control_sequences(output))
    if not json_fragment.strip():
        raise JSONOutputEmptyError(output)
    normalized_json = normalize_wrapped_json(json_fragment)
    return parse_json_fragment(normalized_json, output)


def extract_json_fragment(output: str) -> str:
    """Extract the widest valid JSON payload from shell output with prompt noise."""
    decoder = json.JSONDecoder()
    best_fragment = ""
    best_length = -1

    for index, char in enumerate(output):
        if char not in "[{":
            continue
        try:
            _, end = decoder.raw_decode(output[index:])
        except json.JSONDecodeError:
            continue
        if end > best_length:
            best_fragment = output[index : index + end]
            best_length = end

    if best_fragment:
        return best_fragment
    return output


def strip_terminal_control_sequences(output: str) -> str:
    """Remove ANSI/control sequences that can appear in interactive PTY output."""
    cleaned = ANSI_ESCAPE_RE.sub("", output)
    return cleaned.replace("\r", "")


def normalize_wrapped_json(raw_json: str) -> str:
    """Remove PTY hard-wrap newlines that may appear inside JSON string values."""
    normalized: list[str] = []
    in_string = False
    escaped = False
    for char in raw_json:
        if escaped:
            normalized.append(char)
            escaped = False
            continue
        if char == "\\":
            normalized.append(char)
            escaped = True
            continue
        if char == '"':
            normalized.append(char)
            in_string = not in_string
            continue
        if in_string and char in ("\n", "\r"):
            continue
        normalized.append(char)
    return "".join(normalized)


def parse_json_fragment(normalized_json: str, original_output: str) -> object:
    """Parse JSON fragment with tolerant tail trimming for shell/prompt artifacts."""
    try:
        return json.loads(normalized_json)
    except json.JSONDecodeError:
        closing_candidates = ["]", "}"]
        for closing in closing_candidates:
            end = normalized_json.rfind(closing)
            if end < 0:
                continue
            candidate = normalized_json[: end + 1]
            try:
                return json.loads(candidate)
            except json.JSONDecodeError:
                continue
    raise JSONOutputInvalidError(original_output)


def as_event_items(payload: object) -> list[dict[str, object]]:
    """Convert JSON payload into a validated list of event dictionaries."""
    if not isinstance(payload, list):
        return []
    items: list[dict[str, object]] = []
    for item in payload:
        if isinstance(item, dict):
            items.append(dict(item))
    return items


def compact_json(data: dict[str, object]) -> str:
    """Serialize payload into one-line JSON safe for CLI --data argument."""
    return json.dumps(data, separators=(",", ":"))


def get_event_by_id(
    operator_cli_client: CLIClient,
    cli_binary: str,
    event_id: str,
) -> dict[str, object]:
    """Load one event by ID using the dedicated getEvent command."""
    payload = execute_logging_manager_json(
        operator_cli_client,
        cli_binary,
        f"--cmd getEvent --eventID {event_id}",
    )
    items = as_event_items(payload)
    if not items:
        raise EventNotReturnedError(event_id)
    return items[0]


def prepare_event_lifecycle_data(
    operator_cli_client: CLIClient,
    cli_binary: str,
    actor_login: str,
    event_id: str,
) -> None:
    """Prepare one event with occurrence and lifecycle mutations for focused assertions."""
    create_event_data = compact_json(create_event_payload(event_id))
    execute_logging_manager_mutation(
        operator_cli_client,
        cli_binary,
        f"--cmd createEvent --data {shlex.quote(create_event_data)}",
    )

    create_occurrence_data = compact_json(create_occurrence_payload())
    execute_logging_manager_mutation(
        operator_cli_client,
        cli_binary,
        (
            f"--cmd createOccurrence --eventID {event_id} "
            f"--data {shlex.quote(create_occurrence_data)}"
        ),
    )

    acknowledge_data = compact_json(acknowledge_event_payload(event_id, actor_login))
    execute_logging_manager_mutation(
        operator_cli_client,
        cli_binary,
        f"--cmd acknowledgeEvent --data {shlex.quote(acknowledge_data)}",
    )

    status_data = compact_json(update_event_state_payload(event_id))
    execute_logging_manager_mutation(
        operator_cli_client,
        cli_binary,
        f"--cmd updateEventState --data {shlex.quote(status_data)}",
    )

    mute_data = compact_json(mute_event_payload(event_id, actor_login))
    execute_logging_manager_mutation(
        operator_cli_client,
        cli_binary,
        f"--cmd muteEvent --data {shlex.quote(mute_data)}",
    )


def build_test_event_id() -> str:
    """Build a dynamic event ID that follows logging-manager validation rules."""
    now_ns = time.time_ns()
    prefix = f"t{(now_ns // 100_000_000) % 1_000_000_000:09d}"
    suffix = f"{now_ns % 100_000_000:08d}"
    return f"{prefix}_{suffix}"


@pytest.fixture
def logging_manager_cli_context(
    request: pytest.FixtureRequest,
) -> tuple[CLIClient, str, str]:
    """Resolve the parametrized CLI fixture and its binary/login execution context."""
    context = request.param
    cli_client = request.getfixturevalue(context.fixture_name)
    return cli_client, context.cli_binary, context.actor_login


@pytest.fixture
def prepared_event(
    logging_manager_cli_context: tuple[CLIClient, str, str],
) -> dict[str, object]:
    """Create prepared lifecycle data and return the resulting event snapshot."""
    logging_manager_cli_client, cli_binary, actor_login = logging_manager_cli_context
    event_id = build_test_event_id()
    prepare_event_lifecycle_data(logging_manager_cli_client, cli_binary, actor_login, event_id)
    return get_event_by_id(logging_manager_cli_client, cli_binary, event_id)


@pytest.mark.parametrize(
    "command",
    [pytest.param(command, id=case_id) for command, case_id in GET_COMMANDS],
)
def test_logging_manager_get_commands_return_event_list_json_schema(
    logging_manager_cli_context: tuple[CLIClient, str, str],
    command: str,
) -> None:
    """Test case: get-commands return JSON arrays that match the event list schema.

    Preparation:
    - The system-logging-manager service and its NATS dependency are running.
    - The operator test user has CLI access permissions.

    Action:
    - Execute the parametrized get command in JSON mode as the operator user.

    Expected result:
    - The command output validates against the expected event-list JSON schema.
    """
    logging_manager_cli_client, cli_binary, _ = logging_manager_cli_context
    payload = execute_logging_manager_json(logging_manager_cli_client, cli_binary, command)
    expect(payload).to_match_schema(EVENT_LIST_SCHEMA)


def test_logging_manager_get_alerts_items_match_event_item_schema(
    logging_manager_cli_context: tuple[CLIClient, str, str],
) -> None:
    """Test case: getAlerts item format matches event item JSON schema.

    Preparation:
    - The system-logging-manager service is running with available events.

    Action:
    - Execute `getAlerts` in JSON mode and read the first item.

    Expected result:
    - The first item validates against the shared event-item schema.
    """
    logging_manager_cli_client, cli_binary, _ = logging_manager_cli_context
    payload = execute_logging_manager_json(
        logging_manager_cli_client,
        cli_binary,
        "--cmd getAlerts --page 1 --pageSize 1",
    )
    items = as_event_items(payload)
    if not items:
        pytest.skip("No events available to validate a single getAlerts item schema.")
    expect(items[0]).to_match_schema(EVENT_ITEM_SCHEMA)


def test_logging_manager_acknowledge_sets_acknowledged_true(
    prepared_event: dict[str, object],
) -> None:
    """Test case: acknowledge mutation marks event as acknowledged.

    Preparation:
    - Prepare one event and apply lifecycle mutations via CLI helper fixture.

    Action:
    - Read the prepared event snapshot from `getAlerts`.

    Expected result:
    - `Acknowledged` is `true` for the prepared event.
    """
    expect(prepared_event.get("Acknowledged")).to_be(True)


def test_logging_manager_mute_sets_muted_true(
    prepared_event: dict[str, object],
) -> None:
    """Test case: mute mutation marks event as muted.

    Preparation:
    - Prepare one event and apply lifecycle mutations via CLI helper fixture.

    Action:
    - Read the prepared event snapshot from `getAlerts`.

    Expected result:
    - `Muted` is `true` for the prepared event.
    """
    expect(prepared_event.get("Muted")).to_be(True)


def test_logging_manager_status_update_sets_failure_status(
    prepared_event: dict[str, object],
) -> None:
    """Test case: state mutation stores failure status.

    Preparation:
    - Prepare one event and apply lifecycle mutations via CLI helper fixture.

    Action:
    - Read the prepared event snapshot from `getAlerts`.

    Expected result:
    - `Status` equals `failure` for the prepared event.
    """
    expect(prepared_event.get("Status")).to_be(UPDATED_STATUS)


def test_logging_manager_status_update_sets_message(
    prepared_event: dict[str, object],
) -> None:
    """Test case: state mutation stores updated message.

    Preparation:
    - Prepare one event and apply lifecycle mutations via CLI helper fixture.

    Action:
    - Read the prepared event snapshot from `getAlerts`.

    Expected result:
    - `Message` equals `operator escalated` for the prepared event.
    """
    expect(prepared_event.get("Message")).to_be(UPDATED_MESSAGE)


def test_logging_manager_occurrence_sets_last_value_type(
    prepared_event: dict[str, object],
) -> None:
    """Test case: occurrence creation stores last value type.

    Preparation:
    - Prepare one event and apply lifecycle mutations via CLI helper fixture.

    Action:
    - Read the prepared event snapshot from `getAlerts`.

    Expected result:
    - `LastValueType` equals `text` for the prepared event.
    """
    expect(prepared_event.get("LastValueType")).to_be(OCCURRENCE_VALUE_TYPE)


def test_logging_manager_occurrence_sets_last_value_text(
    prepared_event: dict[str, object],
) -> None:
    """Test case: occurrence creation stores last value text.

    Preparation:
    - Prepare one event and apply lifecycle mutations via CLI helper fixture.

    Action:
    - Read the prepared event snapshot from `getAlerts`.

    Expected result:
    - `LastValueText` equals `threshold crossed` for the prepared event.
    """
    expect(prepared_event.get("LastValueText")).to_be(OCCURRENCE_VALUE_TEXT)


def test_logging_manager_get_commands_reject_unknown_filter_flags(
    logging_manager_cli_context: tuple[CLIClient, str, str],
) -> None:
    """Test case: getAlerts validation rejects unsupported filter flags.

    Preparation:
    - The system-logging-manager CLI is installed and executable.
    - The operator test user has CLI access permissions.

    Action:
    - Execute `getAlerts` with a filter-like flag not supported by this CLI.

    Expected result:
    - The command fails with argument-validation error.
    """
    logging_manager_cli_client, cli_binary, _ = logging_manager_cli_context
    execute_logging_manager_json(
        logging_manager_cli_client,
        cli_binary,
        f"--cmd getAlerts --page 1 {UNSUPPORTED_FILTER_ARGUMENT}",
        expect_success=False,
    )
    logging_manager_cli_client.assert_output_contains(UNSUPPORTED_FILTER_ERROR)
