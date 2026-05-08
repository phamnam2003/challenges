package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/zap"
)

var (
	endpoint     = flag.String("endpoint", "http://localhost:9000", "endpoint for S3-compatible storage")
	accessKey    = flag.String("access-key", "minioadmin", "access key for S3-compatible storage")
	secretKey    = flag.String("secret-key", "minioadmin", "secret key for S3-compatible storage")
	bucketName   = flag.String("bucket", "my-bucket", "bucket name to use")
	usePathStyle = flag.Bool("path-style", true, "whether to use path-style addressing for S3")
	region       = flag.String("region", "us-east-1", "AWS region for S3-compatible storage")
)

var logger *zap.Logger

func main() {
	flag.Parse()

	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	if *endpoint == "" {
		logger.Fatal("endpoint is required")
	}
	if *accessKey == "" {
		logger.Fatal("access-key is required")
	}
	if *secretKey == "" {
		logger.Fatal("secret-key is required")
	}
	if *bucketName == "" {
		logger.Fatal("bucket is required")
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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(*region),
		config.WithHTTPClient(httpcli),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				*accessKey, *secretKey, "",
			),
		),
	)
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = endpoint
		o.UsePathStyle = *usePathStyle
	})

	// === ListBuckets ===
	buckets, err := ListBuckets(ctx, client)
	if err != nil {
		logger.Fatal("failed to list buckets", zap.Error(err))
	}
	logger.Info("list buckets", zap.Strings("buckets", buckets))

	// === PutObject: upload 5 files vào các "thư mục" khác nhau ===
	dummyData := []byte("hello s3")
	uploads := []string{
		"images/avatar/user1.png",
		"images/avatar/user2.png",
		"images/banner/hero.jpg",
		"documents/report.pdf",
		"readme.txt",
	}
	for _, key := range uploads {
		if err := PutObject(ctx, client, *bucketName, key, dummyData); err != nil {
			logger.Fatal("put object failed", zap.String("key", key), zap.Error(err))
		}
	}
	logger.Info("put objects", zap.Int("count", len(uploads)))

	// === ListObjects với prefix ===
	keys, prefixes, _ := ListObjects(ctx, client, *bucketName, "")
	logger.Info("list objects", zap.String("prefix", ""), zap.Strings("keys", keys), zap.Strings("prefixes", prefixes))

	keys, prefixes, _ = ListObjects(ctx, client, *bucketName, "images/")
	logger.Info("list objects", zap.String("prefix", "images/"), zap.Strings("keys", keys), zap.Strings("prefixes", prefixes))

	keys, prefixes, _ = ListObjects(ctx, client, *bucketName, "images/avatar/")
	logger.Info("list objects", zap.String("prefix", "images/avatar/"), zap.Strings("keys", keys), zap.Strings("prefixes", prefixes))

	// === HeadObject ===
	head, err := HeadObject(ctx, client, *bucketName, "images/avatar/user1.png")
	if err != nil {
		logger.Error("head object failed", zap.Error(err))
	} else {
		logger.Info("head object",
			zap.String("key", "images/avatar/user1.png"),
			zap.Int64("content-length", *head.ContentLength),
			zap.String("content-type", *head.ContentType),
			zap.Time("last-modified", *head.LastModified),
		)
	}

	// === CopyObject ===
	copySource := *bucketName + "/images/avatar/user1.png"
	err = CopyObject(ctx, client, copySource, *bucketName, "backup/user1_copy.png")
	if err != nil {
		logger.Error("copy object failed", zap.Error(err))
	} else {
		logger.Info("copy object", zap.String("from", "images/avatar/user1.png"), zap.String("to", "backup/user1_copy.png"))
	}

	// === MultipartUpload ===
	largeData := bytes.Repeat([]byte("x"), 10*1024*1024)
	err = MultipartUpload(ctx, client, *bucketName, "large/bigfile.bin", largeData, 5*1024*1024)
	if err != nil {
		logger.Error("multipart upload failed", zap.Error(err))
	} else {
		logger.Info("multipart upload", zap.String("key", "large/bigfile.bin"), zap.Int("size_mb", 10), zap.Int("parts", 2))
	}

	// === PresignURL GET ===
	getURL, err := PresignURL(ctx, client, http.MethodGet, *bucketName, "readme.txt", 15*time.Minute)
	if err != nil {
		logger.Error("presign GET failed", zap.Error(err))
	} else {
		logger.Info("presign GET", zap.String("key", "readme.txt"), zap.String("url", getURL))
	}

	// === PresignURL PUT + UploadWithPresignedURL ===
	putURL, err := PresignURL(ctx, client, http.MethodPut, *bucketName, "presigned/hello.txt", 15*time.Minute)
	if err != nil {
		logger.Error("presign PUT failed", zap.Error(err))
	} else {
		err = UploadWithPresignedURL(putURL, []byte("uploaded via presigned URL"))
		if err != nil {
			logger.Error("upload with presigned URL failed", zap.Error(err))
		} else {
			logger.Info("upload with presigned URL", zap.String("key", "presigned/hello.txt"))
		}
	}

	// === DeleteObject ===
	err = DeleteObject(ctx, client, *bucketName, "backup/user1_copy.png")
	if err != nil {
		logger.Error("delete object failed", zap.Error(err))
	} else {
		logger.Info("delete object", zap.String("key", "backup/user1_copy.png"))
	}

	// === GetObject ===
	content, err := GetObject(ctx, client, *bucketName, "presigned/hello.txt")
	if err != nil {
		logger.Error("get object failed", zap.Error(err))
	} else {
		logger.Info("get object", zap.String("key", "presigned/hello.txt"), zap.String("content", string(content)))
	}
}

func ListBuckets(ctx context.Context, client *s3.Client) ([]string, error) {
	output, err := client.ListBuckets(ctx, &s3.ListBucketsInput{BucketRegion: region})
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
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
		Body:   bytes.NewReader(data),
	})
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

func ListObjects(ctx context.Context, client *s3.Client, bucketName, prefix string) (keys []string, prefixes []string, err error) {
	delimiter := "/"
	input := &s3.ListObjectsV2Input{
		Bucket:    &bucketName,
		Delimiter: &delimiter,
	}
	if prefix != "" {
		input.Prefix = &prefix
	}
	output, err := client.ListObjectsV2(ctx, input)
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

func HeadObject(ctx context.Context, client *s3.Client, bucketName, objectKey string) (*s3.HeadObjectOutput, error) {
	return client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	})
}

func CopyObject(ctx context.Context, client *s3.Client, copySource, destBucket, destKey string) error {
	_, err := client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     &destBucket,
		Key:        &destKey,
		CopySource: &copySource,
	})
	return err
}

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

		pn := partNumber
		uploadResp, err := client.UploadPart(ctx, &s3.UploadPartInput{
			Bucket:     &bucketName,
			Key:        &objectKey,
			UploadId:   uploadID,
			PartNumber: &pn,
			Body:       bytes.NewReader(data[start:end]),
		})
		if err != nil {
			_, _ = client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   &bucketName,
				Key:      &objectKey,
				UploadId: uploadID,
			})
			return err
		}

		completedParts = append(completedParts, types.CompletedPart{
			ETag:       uploadResp.ETag,
			PartNumber: &pn,
		})
		logger.Debug("uploaded part", zap.Int32("part", pn), zap.Int("bytes", end-start))
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

func PresignURL(ctx context.Context, client *s3.Client, method, bucketName, objectKey string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(client)
	switch method {
	case http.MethodGet:
		req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: &bucketName,
			Key:    &objectKey,
		}, s3.WithPresignExpires(expiry))
		if err != nil {
			return "", err
		}
		return req.URL, nil
	case http.MethodPut:
		req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
			Bucket: &bucketName,
			Key:    &objectKey,
		}, s3.WithPresignExpires(expiry))
		if err != nil {
			return "", err
		}
		return req.URL, nil
	default:
		return "", fmt.Errorf("unsupported presign method: %s", method)
	}
}

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
		return fmt.Errorf("presigned upload failed: %s %s", resp.Status, string(body))
	}
	return nil
}
