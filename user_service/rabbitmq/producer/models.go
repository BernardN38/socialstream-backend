package rabbitmq_producer

type UserDeletedMessage struct {
	UserId int32 `json:"userId"`
}
