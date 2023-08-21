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

func (rc *RpcClient) UploadMedia(ImageUpload *ImageUpload) error {
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

func (rc *RpcClient) DeleteMedia(mediaId uuid.UUID) error {
	var replyErr error
	err := rc.mediaServiceRpcClient.Call("RpcServer.DeleteImage", mediaId, &replyErr)
	if err != nil {
		return err
	}
	if replyErr != nil {
		return replyErr
	}
	return nil
}
