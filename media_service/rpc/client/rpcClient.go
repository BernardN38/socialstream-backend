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

func (rc *RpcClient) GetUserProfileImageIdRpc(userId int32) (int32, error) {
	var reply int32
	err := rc.userServiceRpcClient.Call("RpcServer.GetUserProfileImageId", userId, &reply)
	if err != nil {
		return 0, err
	}
	return reply, nil
}

type ProfileImageUpdateInput struct {
	UserId  int32
	MediaId int32
}

func (rc *RpcClient) UpdateUserProfileImage(imageUpdate ProfileImageUpdateInput) error {
	var reply error
	err := rc.userServiceRpcClient.Call("RpcServer.UpdateUserProfileImageId", imageUpdate, &reply)
	if err != nil {
		return err
	}
	if reply != nil {
		return reply
	}
	return nil
}
