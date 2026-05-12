package loggingmanager

import (
	loggingmanagerdto "lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/model"
)

type ListAlertsInput = loggingmanagerdto.ListEventsInput

type ListAlertsResponse struct {
	Items []model.Event `json:"items"`
}
