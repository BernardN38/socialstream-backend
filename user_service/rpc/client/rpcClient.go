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
	mediaServiceRpcClient *rpc.Client
}

func New(mediaServiceClient *rpc.Client) (*RpcClient, error) {
	return &RpcClient{
		mediaServiceRpcClient: mediaServiceClient,
	}, nil
}

func (rc *RpcClient) UploadImage(ImageUpload *ImageUpload) error {
	var replyErr error
	err := rc.mediaServiceRpcClient.Call("RpcServer.UploadImage", ImageUpload, &replyErr)
	if err != nil {
		return err
	}
	if replyErr != nil {
		return replyErr
	}
	return nil
}
