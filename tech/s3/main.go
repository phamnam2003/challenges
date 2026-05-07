package main

import (
	"bytes"
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	mode         = flag.String("mode", "minio", "mode to run object storage: minio | aws | seaweedfs | rustfs")
	endpoint     = flag.String("endpoint", "http://localhost:9000", "endpoint for S3-compatible storage")
	accessKey    = flag.String("access-key", "minioadmin", "access key for S3-compatible storage")
	secretKey    = flag.String("secret-key", "minioadmin", "secret key for S3-compatible storage")
	bucketName   = flag.String("bucket", "my-bucket", "bucket name to use")
	usePathStyle = flag.Bool("path-style", true, "whether to use path-style addressing for S3")
	region       = flag.String("region", "us-east-1", "AWS region for S3-compatible storage")
)

func main() {
	flag.Parse()
	if *endpoint == "" {
		panic("endpoint is required")
	}
	if *accessKey == "" {
		panic("access-key is required")
	}
	if *secretKey == "" {
		panic("secret-key is required")
	}
	if *bucketName == "" {
		panic("bucket is required")
	}

	httpcli := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(*region),
		config.WithHTTPClient(httpcli),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				*accessKey, *secretKey, "",
			),
		),
	)
	if err != nil {
		panic(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = endpoint
		o.UsePathStyle = *usePathStyle
	})
	buckets, err := ListBuckets(context.Background(), client)
	if err != nil {
		panic("failed to list buckets: " + err.Error())
	}
	log.Printf("buckets: %+v", buckets)

	file, err := os.ReadFile("./tech/s3/docker-compose.yml")
	if err != nil {
		panic("failed to read docker-compose file: " + err.Error())
	}
	err = PutObject(context.Background(), client, *bucketName, "docker-compose.yml", file)
	if err != nil {
		panic("failed to put object: " + err.Error())
	}
	obj, err := ListObjects(context.Background(), client, *bucketName)
	if err != nil {
		panic("failed to list objects: " + err.Error())
	}
	log.Printf("objects in bucket: %+v", obj)
	for _, key := range obj {
		data, err := GetObject(context.Background(), client, *bucketName, key)
		if err != nil {
			log.Printf("failed to get object %s: %v", key, err)
			continue
		}
		log.Printf("object %s content:\n%s", key, string(data))
	}
}

func ListBuckets(ctx context.Context, client *s3.Client) ([]string, error) {
	output, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	buckets := make([]string, 0, len(output.Buckets))
	for _, b := range output.Buckets {
		buckets = append(buckets, *b.Name)
	}
	return buckets, nil
}

func CreateBucket(ctx context.Context, client *s3.Client, bucketName string) error {
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: &bucketName,
	})
	return err
}

func DeleteBucket(ctx context.Context, client *s3.Client, bucketName string) error {
	_, err := client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: &bucketName,
	})
	return err
}

func PutObject(ctx context.Context, client *s3.Client, bucketName, objectKey string, data []byte) error {
	o, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
		Body:   bytes.NewReader(data),
	})
	log.Printf("put object: %+v", o)
	return err
}

func GetObject(ctx context.Context, client *s3.Client, bucketName, objectKey string) ([]byte, error) {
	output, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	})
	if err != nil {
		return nil, err
	}
	defer output.Body.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(output.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DeleteObject(ctx context.Context, client *s3.Client, bucketName, objectKey string) error {
	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	})
	return err
}

func ListObjects(ctx context.Context, client *s3.Client, bucketName string) ([]string, error) {
	output, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	})
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(output.Contents))
	for _, obj := range output.Contents {
		keys = append(keys, *obj.Key)
	}
	return keys, nil
}
