package rabbitmq_producer

import "github.com/google/uuid"

type UserProfileImageUploadMsg struct {
	UserId  int32 `json:"userId"`
	MediaId int32 `json:"mediaId"`
}
type MediaUploadedMsg struct {
	MediaId              int32     `json:"mediaId"`
	ExternalIdFull       uuid.UUID `json:"externalIdFull"`
	ExternalIdCompressed uuid.UUID `json:"externalIdCompressed"`
	ContentType          string    `json:"contentType"`
}
type ExternalIdDeletedMsg struct {
	ExternalId uuid.UUID `json:"externalId"`
}
type MediaIdDeletedMsg struct {
	MediaId int32 `json:"mediaId"`
}
