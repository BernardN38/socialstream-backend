package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"time"

	rabbitmq_producer "github.com/BernardN38/socialstream-backend/post_service/rabbitmq/producer"
	rpc_client "github.com/BernardN38/socialstream-backend/post_service/rpc/client"
	"github.com/BernardN38/socialstream-backend/post_service/sql/posts"
	"github.com/redis/go-redis/v9"
)

type PostService struct {
	postDb          *sql.DB
	postQuries      *posts.Queries
	redisClient     *redis.Client
	rpcClient       *rpc_client.RpcClient
	rabbitmProducer *rabbitmq_producer.RabbitMQProducer
	config          *PostServiceConfig
}
type PostServiceConfig struct {
	MinioBucketName string
}

func New(db *sql.DB, rdb *redis.Client, rpcClient *rpc_client.RpcClient, rabbitmqProducer *rabbitmq_producer.RabbitMQProducer, config PostServiceConfig) (*PostService, error) {
	dbQuries := posts.New(db)
	return &PostService{
		postDb:          db,
		postQuries:      dbQuries,
		redisClient:     rdb,
		rpcClient:       rpcClient,
		rabbitmProducer: rabbitmqProducer,
		config:          &config,
	}, nil
}

func (p *PostService) GetAllPosts(ctx context.Context) ([]posts.Post, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	postsCh := make(chan []posts.Post)
	errCh := make(chan error)
	go func() {
		posts, err := p.postQuries.GetAll(timeoutCtx)
		if err != nil {
			errCh <- err
			return
		}
		postsCh <- posts
	}()
	select {
	case posts := <-postsCh:
		return posts, nil
	case err := <-errCh:
		return nil, err
	case <-timeoutCtx.Done():
		return nil, timeoutCtx.Err()
	}
}
func (p *PostService) DeletePost(ctx context.Context, postId int32, userId int32) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	successCh := make(chan bool)
	errCh := make(chan error)
	go func() {
		postOwnerId, err := p.postQuries.GetPost(timeoutCtx, postId)
		if err != nil {
			errCh <- err
			return
		}
		if userId != postOwnerId {
			errCh <- errors.New("unathorized")
			return
		}
		err = p.postQuries.DeletePost(timeoutCtx, posts.DeletePostParams{
			ID:     postId,
			UserID: userId,
		})
		if err != nil {
			errCh <- err
			return
		}
		successCh <- true
	}()
	select {
	case <-successCh:
		return nil
	case err := <-errCh:
		return err
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	}
}

type PostPageReq struct {
	PageNo   int32 `json:"pageNo"`
	PageSize int32 `json:"pageSize"`
}
type PostPageResp struct {
	Posts      []posts.Post `json:"posts"`
	IsLastPage bool         `json:"isLastPage"`
}

func (p *PostService) GetUserPostsPaginated(ctx context.Context, userId int32, page PostPageReq) (*PostPageResp, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	postsCh := make(chan []posts.Post)
	errCh := make(chan error)
	offset, limit := calculateOffsetAndLimit(page.PageNo, page.PageSize)
	go func() {

		posts, err := p.postQuries.GetPostPage(timeoutCtx, posts.GetPostPageParams{
			UserID: userId,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			errCh <- err
			return
		}
		postsCh <- posts
	}()
	select {
	case posts := <-postsCh:
		resp, err := createPostPageResp(posts, int(page.PageNo), int(page.PageSize), int(limit), int(offset))
		if err != nil {
			return nil, err
		}
		return resp, nil
	case err := <-errCh:
		return nil, err
	case <-timeoutCtx.Done():
		return nil, timeoutCtx.Err()
	}

}
func (p *PostService) CreatePost(ctx context.Context, input CreatePostInput) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	successCh := make(chan struct{})
	errCh := make(chan error)
	go func() {
		var mediaId int32
		if input.MediaSize > 0 {
			startTime := time.Now().UnixMilli()
			mediaBytes, err := io.ReadAll(input.Media)
			if err != nil {
				errCh <- err
				return
			}
			rpcUpload := &rpc_client.RpcImageUpload{
				ImageData:   mediaBytes,
				UserId:      input.UserId,
				ContentType: input.MediaType,
				Size:        input.MediaSize,
			}

			respId, err := p.rpcClient.UploadMedia(rpcUpload)
			if err != nil {
				errCh <- err
				return
			}
			endTime := time.Now().UnixMilli()
			fmt.Println("upload to media service runtime: ", endTime-startTime)
			mediaId = respId
		}
		err := p.postQuries.CreatePost(timeoutCtx, posts.CreatePostParams{
			UserID:   input.UserId,
			Username: input.Username,
			Body:     input.Body,
			MediaID: sql.NullInt32{
				Int32: mediaId,
				Valid: mediaId > 0,
			},
		})
		if err != nil {
			errCh <- err
			return
		}
		successCh <- struct{}{}
	}()
	select {
	case <-successCh:
		return nil
	case err := <-errCh:
		return err
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	}
}
