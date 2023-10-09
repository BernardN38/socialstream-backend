package rpc_server

import (
	"context"
	"log"
	"net/rpc"

	"github.com/BernardN38/socialstream-backend/user_service/service"
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
type ProfileImageUpdateReq struct {
	UserId  int32
	MediaId int32
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

func (s *RpcServer) GetUserProfileImageId(userId int32, reply *int32) error {
	profileImageMediaId, err := s.userService.GetUserProfileImage(context.Background(), userId)
	if err != nil {
		log.Println(err)
		return err
	}
	*reply = profileImageMediaId
	return nil
}
func (s *RpcServer) UpdateUserProfileImageId(updateReq ProfileImageUpdateReq, reply *error) error {
	ctx := context.Background()
	resp := s.userService.UpdateUserProfileImageId(ctx, updateReq.UserId, updateReq.MediaId)
	if resp != nil {
		log.Println(resp)
		reply = &resp
		return resp
	}
	return nil
}
