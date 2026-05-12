package query

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/enum"
)

const defaultListEventsPageSize = 25

var activeStatuses = []string{
	string(enum.StatusHigh),
	string(enum.StatusLow),
	string(enum.StatusActive),
	string(enum.StatusFailure),
}

var (
	errInvalidPageNumber       = errors.New("list events page must be greater than zero")
	errInvalidPageSize         = errors.New("list events page size must be greater than zero")
	errInvalidFilterKey        = errors.New("unsupported filter key")
	errInvalidFilterCondition  = errors.New("unsupported filter condition")
	errEmptyFilterValues       = errors.New("filter values are required")
	errInvalidBooleanFilter    = errors.New("boolean filter requires true/false/1/0 value")
	errInvalidBetweenCondition = errors.New("between condition requires exactly two values")
)

// BuildListEventsQuery builds the aggregate read query for current events with
// optional filters and pagination.
//
// Result contract:
//   - Returns one row per current event.
//   - Joins lifecycle and state 1:1 tables.
//   - Includes the most recent occurrence payload via correlated subquery.
//   - Applies validated, allowlisted filters only.
func BuildListEventsQuery(input dto.ListEventsInput) (Query, error) {
	pageSize, err := resolvePageSize(input.PageSize)
	if err != nil {
		return Query{}, err
	}
	if input.Page <= 0 {
		return Query{}, errInvalidPageNumber
	}

	whereSQL, whereArgs, err := buildWhereClause(input.Filters)
	if err != nil {
		return Query{}, err
	}

	return buildPaginatedEventsQuery(input.Page, pageSize, whereSQL, whereArgs), nil
}

// BuildListActiveEventsQuery builds the list query with an enforced
// "state != normal" filter.
func BuildListActiveEventsQuery(input dto.ListEventsInput) (Query, error) {
	return BuildListEventsQuery(withPrependedFilter(input, buildActiveFilter()))
}

// BuildListActiveUnacknowledgedEventsQuery builds the list query with enforced
// active-state and unacknowledged filters.
func BuildListActiveUnacknowledgedEventsQuery(input dto.ListEventsInput) (Query, error) {
	withDefaults := input
	withDefaults.Filters = append([]dto.Filter{
		buildActiveFilter(),
		{
			Key:       dto.FilterKeyAcknowledged,
			Condition: dto.FilterConditionEQ,
			Values:    []string{"false"},
		},
	}, input.Filters...)
	return BuildListEventsQuery(withDefaults)
}

func buildActiveFilter() dto.Filter {
	return dto.Filter{
		Key:       dto.FilterKeyState,
		Condition: dto.FilterConditionIN,
		Values:    activeStatuses,
	}
}

// BuildGetEventHistoryQuery builds the history query for one business event ID.
func BuildGetEventHistoryQuery(input dto.GetEventHistoryInput) Query {
	return Query{
		SQL: "SELECT o.rec_id, e.event_id, o.ts, o.value_type, o.value_num, o.value_text, o.value_bool, o.value_unit " +
			"FROM occurrences o " +
			"JOIN events e ON e.event_id = o.event_id " +
			"WHERE o.event_id = ? " +
			"ORDER BY o.rec_id ASC",
		Args: []any{input.EventID},
	}
}

func withPrependedFilter(input dto.ListEventsInput, filter dto.Filter) dto.ListEventsInput {
	out := input
	out.Filters = append([]dto.Filter{filter}, input.Filters...)
	return out
}

func resolvePageSize(requestedPageSize int) (int, error) {
	if requestedPageSize == 0 {
		return defaultListEventsPageSize, nil
	}
	if requestedPageSize < 0 {
		return 0, errInvalidPageSize
	}
	return requestedPageSize, nil
}

func buildPaginatedEventsQuery(page int, pageSize int, whereSQL string, whereArgs []any) Query {
	offset := (page - 1) * pageSize
	args := make([]any, 0, len(whereArgs)+2)
	args = append(args, whereArgs...)
	args = append(args, pageSize, offset)

	sql := buildBaseListEventsSQL()
	if whereSQL != "" {
		sql += " WHERE " + whereSQL
	}
	sql += " ORDER BY e.created_at DESC, e.event_id DESC LIMIT ? OFFSET ?"

	return Query{SQL: sql, Args: args}
}

func buildBaseListEventsSQL() string {
	return "SELECT " +
		"e.rec_id, e.event_id, e.category, e.severity, e.priority, e.created_at, e.source, " +
		"l.acknowledged, l.acknowledged_by, l.acknowledged_at, l.muted, l.muted_by, l.muted_at, " +
		"s.status, s.message, " +
		"o.ts AS last_value_ts, o.value_type AS last_value_type, o.value_num AS last_value_num, o.value_text AS last_value_text, o.value_bool AS last_value_bool, o.value_unit AS last_value_unit " +
		"FROM events e " +
		"JOIN lifecycle l ON l.event_rec_id = e.rec_id " +
		"JOIN event_state s ON s.event_rec_id = e.rec_id " +
		"LEFT JOIN occurrences o ON o.rec_id = (" +
		"SELECT rec_id FROM occurrences oo WHERE oo.event_id = e.event_id ORDER BY oo.rec_id DESC LIMIT 1" +
		")"
}

func buildWhereClause(filters []dto.Filter) (string, []any, error) {
	if len(filters) == 0 {
		return "", nil, nil
	}

	parts := make([]string, 0, len(filters))
	args := make([]any, 0, len(filters)*2)

	for _, filter := range filters {
		column, err := mapFilterColumn(filter.Key)
		if err != nil {
			return "", nil, err
		}

		conditionSQL, conditionArgs, err := buildConditionClause(column, filter)
		if err != nil {
			return "", nil, err
		}

		parts = append(parts, conditionSQL)
		args = append(args, conditionArgs...)
	}

	return strings.Join(parts, " AND "), args, nil
}

func mapFilterColumn(key dto.FilterKey) (string, error) {
	switch key {
	case dto.FilterKeyCategory:
		return "e.category", nil
	case dto.FilterKeyState:
		return "s.status", nil
	case dto.FilterKeyAcknowledged:
		return "l.acknowledged", nil
	case dto.FilterKeyMuted:
		return "l.muted", nil
	case dto.FilterKeyCreatedAt:
		return "e.created_at", nil
	default:
		return "", fmt.Errorf("%w: %s", errInvalidFilterKey, key)
	}
}

func buildConditionClause(column string, filter dto.Filter) (string, []any, error) {
	if len(filter.Values) == 0 {
		return "", nil, errEmptyFilterValues
	}

	switch filter.Condition {
	case dto.FilterConditionEQ:
		value, err := normalizeFilterValue(filter.Key, filter.Values[0])
		if err != nil {
			return "", nil, err
		}
		return column + " = ?", []any{value}, nil
	case dto.FilterConditionIN:
		return buildINConditionClause(column, filter)
	case dto.FilterConditionBetween:
		return buildBetweenConditionClause(column, filter)
	default:
		return "", nil, fmt.Errorf("%w: %s", errInvalidFilterCondition, filter.Condition)
	}
}

func buildINConditionClause(column string, filter dto.Filter) (string, []any, error) {
	values := make([]any, 0, len(filter.Values))
	placeholders := make([]string, 0, len(filter.Values))

	for _, rawValue := range filter.Values {
		value, err := normalizeFilterValue(filter.Key, rawValue)
		if err != nil {
			return "", nil, err
		}

		values = append(values, value)
		placeholders = append(placeholders, "?")
	}

	return fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", ")), values, nil
}

func buildBetweenConditionClause(column string, filter dto.Filter) (string, []any, error) {
	if len(filter.Values) != 2 {
		return "", nil, errInvalidBetweenCondition
	}

	start, err := normalizeFilterValue(filter.Key, filter.Values[0])
	if err != nil {
		return "", nil, err
	}
	end, err := normalizeFilterValue(filter.Key, filter.Values[1])
	if err != nil {
		return "", nil, err
	}

	return column + " BETWEEN ? AND ?", []any{start, end}, nil
}

func normalizeFilterValue(key dto.FilterKey, value string) (any, error) {
	switch key {
	case dto.FilterKeyAcknowledged, dto.FilterKeyMuted:
		return normalizeBooleanFilterValue(value)
	default:
		return value, nil
	}
}

func normalizeBooleanFilterValue(value string) (int, error) {
	switch strings.ToLower(value) {
	case "1", "true":
		return 1, nil
	case "0", "false":
		return 0, nil
	default:
		return 0, fmt.Errorf("%w: %s", errInvalidBooleanFilter, strconv.Quote(value))
	}
}
