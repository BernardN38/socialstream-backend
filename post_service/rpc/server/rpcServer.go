package rpc_server

import (
	"net/rpc"

	"github.com/BernardN38/socialstream-backend/post_service/service"
	"github.com/google/uuid"
)

// Handler is the struct which exposes the User Server methods
type RpcServer struct {
	postService *service.PostService
}
type ImageUpload struct {
	ImageData   []byte
	MediaId     uuid.UUID
	ContentType string
}
type ProfileImageUpdateReq struct {
	UserId  int32
	MediaId int32
}

// New returns the object for the RPC handler
func NewRpcServer(postService *service.PostService) (*RpcServer, error) {
	s := &RpcServer{
		postService: postService,
	}
	err := rpc.Register(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *RpcServer) health(userId int32, reply *bool) error {
	*reply = true
	return nil
}
