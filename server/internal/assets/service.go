package assets

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

var allowedTypes = map[string]string{
	"image/jpeg":       ".jpg",
	"image/png":        ".png",
	"image/gif":        ".gif",
	"image/webp":       ".webp",
	"video/mp4":        ".mp4",
	"video/webm":       ".webm",
	"video/quicktime":  ".mov",
}

type Service struct {
	client     *s3.Client
	bucket     string
	cdnBaseURL string
}

func NewService(ctx context.Context) (*Service, error) {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("S3_REGION")
	endpoint := os.Getenv("S3_ENDPOINT")
	cdnBaseURL := os.Getenv("CDN_BASE_URL")

	if accessKeyID == "" || secretKey == "" || bucket == "" {
		return nil, fmt.Errorf("AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and S3_BUCKET are required")
	}
	if region == "" {
		region = "auto"
	}
	if cdnBaseURL == "" {
		return nil, fmt.Errorf("CDN_BASE_URL is required")
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	clientOpts := []func(*s3.Options){}
	if endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	}

	return &Service{
		client:     s3.NewFromConfig(cfg, clientOpts...),
		bucket:     bucket,
		cdnBaseURL: strings.TrimRight(cdnBaseURL, "/"),
	}, nil
}

func (s *Service) Upload(ctx context.Context, contentType string, body io.Reader, size int64) (string, error) {
	ext, ok := allowedTypes[contentType]
	if !ok {
		return "", fmt.Errorf("unsupported content type: %s", contentType)
	}

	key := "creatives/" + uuid.New().String() + ext

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return "", fmt.Errorf("upload to S3: %w", err)
	}

	return s.cdnBaseURL + "/" + key, nil
}
