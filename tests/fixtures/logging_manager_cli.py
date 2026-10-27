"""Reusable Hyperion helpers for logging-manager CLI workflows."""

import json
import re

from hyperiontf import CLIClient, expect

ANSI_ESCAPE_RE = re.compile(r"\x1b\[[0-?]*[ -/]*[@-~]")


class JSONOutputEmptyError(AssertionError):
    """Raised when a command expected to return JSON produces empty output."""

    def __init__(self, output: str) -> None:
        super().__init__(f"Expected JSON output, got empty output: {output!r}")


class JSONOutputInvalidError(AssertionError):
    """Raised when a command expected to return JSON produces non-JSON output."""

    def __init__(self, output: str) -> None:
        super().__init__(f"Expected JSON output, got non-JSON payload: {output!r}")


def execute_logging_manager_json(
    cli_client: CLIClient,
    cli_binary: str,
    command: str,
    *,
    expect_success: bool = True,
) -> object:
    """Run one manager CLI command in JSON mode and parse its output."""
    cli_client.execute(f"{cli_binary} {command} --json")
    if expect_success:
        cli_client.assert_exit_code(0)
        if not cli_client.output.strip():
            return {}
        return output_to_json(cli_client)

    expect(cli_client.exit_code).not_to_be(0)
    if not cli_client.output.strip():
        return {}
    return cli_client.output


def output_to_json(cli_client: CLIClient) -> object:
    """Extract and parse a JSON value from PTY output and prompt noise."""
    output = cli_client.output
    normalized_output = normalize_wrapped_json(strip_terminal_control_sequences(output))
    json_fragment = extract_json_fragment(normalized_output)
    if not json_fragment.strip():
        raise JSONOutputEmptyError(output)
    return parse_json_fragment(json_fragment, output)


def extract_json_fragment(output: str) -> str:
    """Extract the widest valid JSON value from shell output."""
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
    return best_fragment or output


def strip_terminal_control_sequences(output: str) -> str:
    """Remove terminal control sequences that can occur in PTY output."""
    return ANSI_ESCAPE_RE.sub("", output).replace("\r", "")


def normalize_wrapped_json(raw_json: str) -> str:
    """Remove PTY-inserted line breaks that can split JSON tokens or values."""
    return raw_json.replace("\r", "").replace("\n", "")


def parse_json_fragment(normalized_json: str, original_output: str) -> object:
    """Parse JSON while tolerating trailing PTY prompt artifacts."""
    try:
        return json.loads(normalized_json)
    except json.JSONDecodeError:
        for closing in ("]", "}"):
            end = normalized_json.rfind(closing)
            if end < 0:
                continue
            try:
                return json.loads(normalized_json[: end + 1])
            except json.JSONDecodeError:
                continue
    raise JSONOutputInvalidError(original_output)


def as_event_items(payload: object) -> list[dict[str, object]]:
    """Return dictionary items from a manager event-list response."""
    if not isinstance(payload, list):
        return []
    return [dict(item) for item in payload if isinstance(item, dict)]
