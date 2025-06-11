package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s3Client)

	object := s3.GetObjectInput{
		Bucket: &bucket,
		Key: &key,
	}

	presignReq, err := presignClient.PresignGetObject(context.TODO(), &object, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", fmt.Errorf("unable to create presign request")
	}

	presignedURL := presignReq.URL

	return presignedURL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}
	
	fields := strings.Split(*video.VideoURL, ",")
	if len(fields) == 0 {
		return database.Video{}, fmt.Errorf("incorrect video url")
	}

	presignedURL, err := generatePresignedURL(cfg.s3Client, fields[0], fields[1], 60 * time.Minute)
	if err != nil {
		return database.Video{}, fmt.Errorf("unable to create presigned url")
	}

	video.VideoURL = &presignedURL

	return video, nil
}