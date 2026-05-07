package main

import (
	"bytes"
	"context"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
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

// ListObjectsWithPrefix lists objects and "subdirectories" under a given prefix.
// Returns object keys and common prefixes (virtual folders).
func ListObjectsWithPrefix(ctx context.Context, client *s3.Client, bucketName, prefix string) (keys []string, prefixes []string, err error) {
	delimiter := "/"
	output, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:    &bucketName,
		Prefix:    &prefix,
		Delimiter: &delimiter,
	})
	if err != nil {
		return nil, nil, err
	}
	for _, obj := range output.Contents {
		keys = append(keys, *obj.Key)
	}
	for _, cp := range output.CommonPrefixes {
		prefixes = append(prefixes, *cp.Prefix)
	}
	return keys, prefixes, nil
}

// HeadObject retrieves object metadata without downloading the body.
func HeadObject(ctx context.Context, client *s3.Client, bucketName, objectKey string) (*s3.HeadObjectOutput, error) {
	return client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	})
}

// CopyObject copies an object within the same bucket or across buckets.
// copySource format: "source-bucket/source-key"
func CopyObject(ctx context.Context, client *s3.Client, copySource, destBucket, destKey string) error {
	_, err := client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     &destBucket,
		Key:        &destKey,
		CopySource: &copySource,
	})
	return err
}

// MultipartUpload uploads data in parts. partSize is the size of each part in bytes (minimum 5MB).
func MultipartUpload(ctx context.Context, client *s3.Client, bucketName, objectKey string, data []byte, partSize int) error {
	createResp, err := client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	})
	if err != nil {
		return err
	}
	uploadID := createResp.UploadId

	var completedParts []types.CompletedPart
	partNumber := int32(1)

	for start := 0; start < len(data); start += partSize {
		end := start + partSize
		if end > len(data) {
			end = len(data)
		}

		uploadResp, err := client.UploadPart(ctx, &s3.UploadPartInput{
			Bucket:     &bucketName,
			Key:        &objectKey,
			UploadId:   uploadID,
			PartNumber: &partNumber,
			Body:       bytes.NewReader(data[start:end]),
		})
		if err != nil {
			// Abort on failure
			_, _ = client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   &bucketName,
				Key:      &objectKey,
				UploadId: uploadID,
			})
			return err
		}

		completedParts = append(completedParts, types.CompletedPart{
			ETag:       uploadResp.ETag,
			PartNumber: &partNumber,
		})
		log.Printf("uploaded part %d (%d bytes)", partNumber, end-start)
		partNumber++
	}

	_, err = client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   &bucketName,
		Key:      &objectKey,
		UploadId: uploadID,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	return err
}

// PresignGetObject generates a presigned URL for downloading an object.
func PresignGetObject(ctx context.Context, client *s3.Client, bucketName, objectKey string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(client)
	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// PresignPutObject generates a presigned URL for uploading an object.
func PresignPutObject(ctx context.Context, client *s3.Client, bucketName, objectKey string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(client)
	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// UploadWithPresignedURL uploads data to a presigned PUT URL using a plain HTTP client.
func UploadWithPresignedURL(presignedURL string, data []byte) error {
	req, err := http.NewRequest(http.MethodPut, presignedURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.ContentLength = int64(len(data))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("presigned upload failed: %s %s", resp.Status, string(body))
	}
	return nil
}
