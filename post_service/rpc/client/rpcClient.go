package rpc_client

import (
	"net/rpc"

	"github.com/google/uuid"
)

type RpcClient struct {
	mediaServiceRpcClient *rpc.Client
}

type RpcImageUpload struct {
	ImageData   []byte
	UserId      int32
	ContentType string
	Size        int64
}

func New(mediaServiceClient *rpc.Client) (*RpcClient, error) {
	return &RpcClient{
		mediaServiceRpcClient: mediaServiceClient,
	}, nil
}

func (rc *RpcClient) UploadMedia(ImageUpload *RpcImageUpload) (int32, error) {
	var mediaId int32
	err := rc.mediaServiceRpcClient.Call("RpcServer.UploadImage", ImageUpload, &mediaId)
	if err != nil {
		return 0, err
	}
	return mediaId, nil
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
