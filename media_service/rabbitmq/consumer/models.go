package rabbitmq_comsumer

import "github.com/google/uuid"

type ExternalIdDeletedMsg struct {
	ExternalId uuid.UUID `json:"externalId"`
}

type MediaCompressedMsg struct {
	MediaId              int32     `json:"mediaId"`
	ExternalIdCompressed uuid.UUID `json:"externalIdCompressed"`
}
type MediaDeletedMsg struct {
	MediaId int32 `json:"mediaId"`
}
