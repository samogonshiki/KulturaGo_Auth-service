package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3 struct {
	bucket string
	public string
	up     *manager.Uploader
	ps     *s3.PresignClient
}

func New(
	ctx context.Context,
	bucket, region, endpoint, publicURL string,
	accessKey, secretKey string,
) (*S3, error) {

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, r string, _ ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL:               endpoint,
						HostnameImmutable: true,
						SigningRegion:     region,
					}, nil
				})),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	cl := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3{
		bucket: bucket,
		public: strings.TrimRight(publicURL, "/"),
		up:     manager.NewUploader(cl),
		ps:     s3.NewPresignClient(cl),
	}, nil
}

func (s *S3) PresignAvatarPut(ctx context.Context, uid int64, email string) (putURL, key string, err error) {
	key = FileName(uid, email)

	out, err := s.ps.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		ACL:    s3types.ObjectCannedACLPublicRead,
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", "", err
	}

	return out.URL, key, nil
}

func (s *S3) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", s.public, s.bucket, key)
}

func (s *S3) Upload(ctx context.Context, key string, body io.Reader, ct string) (string, error) {
	if ct == "" {
		ct = "application/octet-stream"
	}
	_, err := s.up.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(ct),
		ACL:         s3types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", err
	}
	return s.PublicURL(key), nil
}
