package rpc_server

import (
	"net/rpc"

	"github.com/BernardN38/flutter-backend/user_service/service"
	"github.com/google/uuid"
)

// Handler is the struct which exposes the User Server methods
type RpcServer struct {
	mediaService *service.UserService
}
type ImageUpload struct {
	ImageData   []byte
	MediaId     uuid.UUID
	ContentType string
}

// New returns the object for the RPC handler
func NewRpcServer(mediaService service.UserService) *RpcServer {
	s := &RpcServer{
		mediaService: &mediaService,
	}
	err := rpc.Register(s)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *RpcServer) Test(payload ImageUpload, reply *error) error {
	return nil
}
