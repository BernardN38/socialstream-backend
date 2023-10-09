package service

import (
	"errors"

	"github.com/BernardN38/socialstream-backend/post_service/sql/posts"
)

func calculateOffsetAndLimit(pageNo int32, pageSize int32) (int32, int32) {
	// account for 0 index pageNo
	offset := (pageNo - 1) * pageSize
	// to check if it is the last page
	limit := pageSize + 1
	return offset, limit
}

func createPostPageResp(posts []posts.Post, pageNo int, pageSize int, limit int, offset int) (*PostPageResp, error) {
	//check posts is not empty
	if len(posts) <= 0 {
		return nil, errors.New("no posts found")
	}
	// check if last page
	isLastPage := len(posts) <= pageSize
	// check which posts to return
	if !isLastPage {
		posts = posts[:len(posts)-1]
	}

	return &PostPageResp{
		Posts:      posts,
		IsLastPage: isLastPage,
	}, nil

}
