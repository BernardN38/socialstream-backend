package rpc_server

import (
	"bytes"
	"context"
	"net/rpc"

	"github.com/BernardN38/socialstream-backend/media_service/service"
	"github.com/google/uuid"
)

type ImageUpload struct {
	ImageData   []byte
	UserId      int32
	ContentType string
	Size        int64
}

// Handler is the struct which exposes the User Server methods
type RpcServer struct {
	mediaService *service.MediaService
}

// New returns the object for the RPC handler
func New(mediaService *service.MediaService) *RpcServer {
	s := &RpcServer{
		mediaService: mediaService,
	}
	err := rpc.Register(s)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *RpcServer) UploadImage(payload ImageUpload, reply *int32) error {
	reader := bytes.NewReader(payload.ImageData)
	mediaId, err := s.mediaService.UploadMedia(context.Background(), service.ImageUpload{
		MediaData:     reader,
		UserId:        payload.UserId,
		ContentType:   payload.ContentType,
		ContentLength: payload.Size,
	})
	if err != nil {
		return err
	}
	*reply = *mediaId
	return nil
}
func (s *RpcServer) DeleteImage(payload uuid.UUID, reply *error) error {
	err := s.mediaService.DeleteExternalId(context.Background(), payload)
	if err != nil {
		return err
	}
	reply = &err
	return nil
}
