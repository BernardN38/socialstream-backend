package rpc_server

import (
	"context"
	"net/rpc"

	"github.com/BernardN38/flutter-backend/user_service/service"
	"github.com/google/uuid"
)

// Handler is the struct which exposes the User Server methods
type RpcServer struct {
	userService *service.UserService
}
type ImageUpload struct {
	ImageData   []byte
	MediaId     uuid.UUID
	ContentType string
}

// New returns the object for the RPC handler
func NewRpcServer(userService *service.UserService) (*RpcServer, error) {
	s := &RpcServer{
		userService: userService,
	}
	err := rpc.Register(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *RpcServer) GetUserProfileImageId(userId int32, reply *uuid.UUID) error {
	profileImageMediaId, err := s.userService.GetUserProfileImage(context.Background(), userId)
	if err != nil {
		return err
	}
	*reply = profileImageMediaId
	return nil
}
