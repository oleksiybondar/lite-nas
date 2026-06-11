"""Shared CLI payload and expectation data for logging-manager system tests."""

from typing import Any

DEFAULT_CATEGORY = "disk_health"
DEFAULT_SEVERITY = "warning"
DEFAULT_PRIORITY = 2
DEFAULT_SOURCE = "system-tests"

OCCURRENCE_TIMESTAMP = "2026-05-13T10:00:00Z"
OCCURRENCE_VALUE_TYPE = "text"
OCCURRENCE_VALUE_TEXT = "threshold crossed"

ACKNOWLEDGED_AT = "2026-05-13T10:01:00Z"
MUTED_AT = "2026-05-13T10:02:00Z"

UPDATED_STATUS = "failure"
UPDATED_MESSAGE = "operator escalated"

UNSUPPORTED_FILTER_ARGUMENT = "--filter category:eq:disk_health"
UNSUPPORTED_FILTER_ERROR = "unknown argument: --filter"

GET_COMMANDS = (
    ("--cmd getAlerts --page 1 --pageSize 1", "get-alerts"),
    ("--cmd getActiveEvents --page 1 --pageSize 1", "get-active-events"),
    (
        "--cmd getActiveUnacknowledgedEvents --page 1 --pageSize 1",
        "get-active-unacknowledged-events",
    ),
)


def create_event_payload(event_id: str) -> dict[str, Any]:
    """Build createEvent payload for logging-manager CLI mutation flows."""
    return {
        "event_id": event_id,
        "category": DEFAULT_CATEGORY,
        "severity": DEFAULT_SEVERITY,
        "priority": DEFAULT_PRIORITY,
        "source": DEFAULT_SOURCE,
    }


def create_occurrence_payload() -> dict[str, Any]:
    """Build createOccurrence payload for logging-manager CLI mutation flows."""
    return {
        "timestamp": OCCURRENCE_TIMESTAMP,
        "value_type": OCCURRENCE_VALUE_TYPE,
        "value_text": OCCURRENCE_VALUE_TEXT,
    }


def acknowledge_event_payload(event_id: str, operator_login: str) -> dict[str, Any]:
    """Build acknowledgeEvent payload for logging-manager CLI mutation flows."""
    return {
        "event_id": event_id,
        "acknowledged_by": operator_login,
        "acknowledged_at": ACKNOWLEDGED_AT,
    }


def update_event_state_payload(event_id: str) -> dict[str, Any]:
    """Build updateEventState payload for logging-manager CLI mutation flows."""
    return {
        "event_id": event_id,
        "status": UPDATED_STATUS,
        "message": UPDATED_MESSAGE,
    }


def mute_event_payload(event_id: str, operator_login: str) -> dict[str, Any]:
    """Build muteEvent payload for logging-manager CLI mutation flows."""
    return {
        "event_id": event_id,
        "muted_by": operator_login,
        "muted_at": MUTED_AT,
    }
