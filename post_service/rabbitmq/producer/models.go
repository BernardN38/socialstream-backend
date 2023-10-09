package rabbitmq_producer

type UserDeletedMsg struct {
	UserId int32 `json:"userId"`
}

type MediaDeletedMsg struct {
	MediaId string `json:"mediaId"`
}
