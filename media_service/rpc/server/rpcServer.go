package rpc_server

import (
	"context"
	"net/rpc"

	"github.com/BernardN38/flutter-backend/media_service/service"
	"github.com/google/uuid"
)

type ImageUpload struct {
	ImageData   []byte
	MediaId     uuid.UUID
	ContentType string
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

func (s *RpcServer) UploadImage(payload ImageUpload, reply *error) error {
	var resp error
	resp = s.mediaService.UploadMedia(context.Background(), service.RpcImageUpload{
		MediaData:   payload.ImageData,
		MediaId:     payload.MediaId,
		ContentType: payload.ContentType,
	})
	reply = &resp
	return nil
}
