package service

import (
	"context"
	"log"

	"github.com/minio/minio-go/v7"
)

func setup(minioClient minio.Client, bucketName string) error {
	exists, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		return err
	}
	if !exists {
		// Create the bucket
		err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
		log.Printf("Bucket '%s' created successfully\n", bucketName)
	} else {
		log.Printf("Bucket '%s' already exists\n", bucketName)
	}
	return nil
}
