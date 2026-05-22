package loggingmanager

import (
	"encoding/json"

	loggingmanagerdto "lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/loggingmanager/model"
)

type (
	// ListAlertsInput defines paginated read input including auth context.
	ListAlertsInput struct {
		AccessToken string                     `json:"access_token" validate:"required,min=1,max=8192"`
		Page        int                        `json:"page" validate:"required,gte=1"`
		PageSize    int                        `json:"page_size" validate:"omitempty,gte=1,lte=500"`
		Filters     []loggingmanagerdto.Filter `json:"filters,omitempty" validate:"omitempty,dive"`
	}
	// GetAlertInput defines single-alert read input including auth context.
	GetAlertInput struct {
		AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
		EventID     string `json:"event_id" validate:"required,max=20,loggingmanager_event_id"`
	}
)

// ToDTO converts read list contract input into core DTO input.
func (input ListAlertsInput) ToDTO() loggingmanagerdto.ListEventsInput {
	return loggingmanagerdto.ListEventsInput{
		Page:     input.Page,
		PageSize: input.PageSize,
		Filters:  input.Filters,
	}
}

// ToDTO converts single-alert contract input into core DTO input.
func (input GetAlertInput) ToDTO() loggingmanagerdto.GetEventHistoryInput {
	return loggingmanagerdto.GetEventHistoryInput{EventID: input.EventID}
}

// ListAlertItem defines the flattened read model for alert list responses.
//
// Contract:
//   - Event, lifecycle, and state fields are exposed at one level.
//   - Last-occurrence fields stay nullable when no occurrence exists.
//   - Meta remains an optional key-value dictionary.
type ListAlertItem struct {
	RecID          int64             `json:"RecID"`
	EventID        string            `json:"EventID"`
	Category       string            `json:"Category"`
	Severity       enum.Severity     `json:"Severity"`
	Priority       int               `json:"Priority"`
	CreatedAt      string            `json:"CreatedAt"`
	Source         string            `json:"Source"`
	EventRecID     int64             `json:"EventRecID"`
	Acknowledged   bool              `json:"Acknowledged"`
	AcknowledgedBy string            `json:"AcknowledgedBy"`
	AcknowledgedAt string            `json:"AcknowledgedAt"`
	Muted          bool              `json:"Muted"`
	MutedBy        string            `json:"MutedBy"`
	MutedAt        string            `json:"MutedAt"`
	Status         enum.Status       `json:"Status"`
	Message        string            `json:"Message"`
	LastRecID      *int64            `json:"LastRecID"`
	LastEventID    *string           `json:"LastEventID"`
	LastEventRecID *int64            `json:"LastEventRecID"`
	LastTimestamp  *string           `json:"LastTimestamp"`
	LastValueType  *enum.ValueType   `json:"LastValueType"`
	LastValueNum   *float64          `json:"LastValueNum"`
	LastValueText  *string           `json:"LastValueText"`
	LastValueBool  *bool             `json:"LastValueBool"`
	LastValueUnit  *string           `json:"LastValueUnit"`
	Meta           map[string]string `json:"Meta,omitempty"`
}

type ListAlertsResponse struct {
	Items []ListAlertItem `json:"items"`
}

type GetAlertResponse struct {
	Item *ListAlertItem `json:"item,omitempty"`
}

type itemsEnvelope struct {
	Items []json.RawMessage `json:"items"`
}

// UnmarshalJSON supports both flattened and legacy nested items payloads.
func (response *ListAlertsResponse) UnmarshalJSON(data []byte) error {
	envelope, err := decodeItemsEnvelope(data)
	if err != nil {
		return err
	}

	if len(envelope.Items) == 0 {
		response.Items = []ListAlertItem{}
		return nil
	}

	legacyPayload, err := isLegacyItemsPayload(envelope.Items[0])
	if err != nil {
		return err
	}

	if legacyPayload {
		return response.unmarshalLegacyItems(data)
	}

	return response.unmarshalFlatItems(data)
}

// BuildListAlertItems maps storage read rows into contract list items.
func BuildListAlertItems(events []model.Event) []ListAlertItem {
	items := make([]ListAlertItem, 0, len(events))
	for _, event := range events {
		items = append(items, BuildListAlertItem(event))
	}
	return items
}

// BuildListAlertItem maps one storage read model event into one contract list item.
func BuildListAlertItem(event model.Event) ListAlertItem {
	item := ListAlertItem{
		RecID:          event.Event.RecID,
		EventID:        event.Event.EventID,
		Category:       event.Event.Category,
		Severity:       event.Event.Severity,
		Priority:       event.Event.Priority,
		CreatedAt:      event.Event.CreatedAt,
		Source:         event.Event.Source,
		EventRecID:     event.Lifecycle.EventRecID,
		Acknowledged:   event.Lifecycle.Acknowledged,
		AcknowledgedBy: event.Lifecycle.AcknowledgedBy,
		AcknowledgedAt: event.Lifecycle.AcknowledgedAt,
		Muted:          event.Lifecycle.Muted,
		MutedBy:        event.Lifecycle.MutedBy,
		MutedAt:        event.Lifecycle.MutedAt,
		Status:         event.State.Status,
		Message:        event.State.Message,
	}
	populateLastOccurrenceFields(&item, event)
	if len(event.Meta) > 0 {
		item.Meta = make(map[string]string, len(event.Meta))
		for _, metaRow := range event.Meta {
			item.Meta[metaRow.MetaKey] = metaRow.MetaValue
		}
	}
	return item
}

func populateLastOccurrenceFields(item *ListAlertItem, event model.Event) {
	if event.LastValue == nil {
		return
	}

	lastRecID := event.LastValue.RecID
	lastEventID := event.LastValue.EventID
	lastEventRecID := event.LastValue.EventRecID
	lastTimestamp := event.LastValue.Timestamp
	lastValueType := event.LastValue.ValueType

	item.LastRecID = &lastRecID
	item.LastEventID = &lastEventID
	item.LastEventRecID = &lastEventRecID
	item.LastTimestamp = &lastTimestamp
	item.LastValueType = &lastValueType
	item.LastValueNum = event.LastValue.ValueNum
	item.LastValueText = event.LastValue.ValueText
	item.LastValueBool = event.LastValue.ValueBool
	item.LastValueUnit = event.LastValue.ValueUnit
}

func decodeItemsEnvelope(data []byte) (itemsEnvelope, error) {
	var envelope itemsEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return itemsEnvelope{}, err
	}
	return envelope, nil
}

func isLegacyItemsPayload(item json.RawMessage) (bool, error) {
	var firstItem map[string]json.RawMessage
	if err := json.Unmarshal(item, &firstItem); err != nil {
		return false, err
	}
	_, isLegacy := firstItem["Event"]
	return isLegacy, nil
}

func (response *ListAlertsResponse) unmarshalLegacyItems(data []byte) error {
	var legacyEnvelope struct {
		Items []model.Event `json:"items"`
	}
	if err := json.Unmarshal(data, &legacyEnvelope); err != nil {
		return err
	}
	response.Items = BuildListAlertItems(legacyEnvelope.Items)
	return nil
}

func (response *ListAlertsResponse) unmarshalFlatItems(data []byte) error {
	var flatEnvelope struct {
		Items []ListAlertItem `json:"items"`
	}
	if err := json.Unmarshal(data, &flatEnvelope); err != nil {
		return err
	}
	response.Items = flatEnvelope.Items
	return nil
}
