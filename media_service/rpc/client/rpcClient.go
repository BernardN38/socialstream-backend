package rpc_client

import (
	"net/rpc"

	"github.com/google/uuid"
)

type ImageUpload struct {
	ImageData   []byte
	MediaId     uuid.UUID
	ContentType string
}

type RpcClient struct {
	userServiceRpcClient *rpc.Client
}

func New(userServiceClient *rpc.Client) (*RpcClient, error) {
	return &RpcClient{
		userServiceRpcClient: userServiceClient,
	}, nil
}

func (rc *RpcClient) GetUserProfileImageIdRpc(userId int32) (uuid.UUID, error) {
	var reply uuid.UUID
	err := rc.userServiceRpcClient.Call("RpcServer.GetUserProfileImageId", userId, &reply)
	if err != nil {
		return uuid.UUID{}, err
	}
	return reply, nil
}
