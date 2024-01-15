package caching

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/git-lfs/git-lfs/v3/config"
)

type S3CachingAdapter struct {
	client        *s3.Client
	configuration *cachingConfiguration
}

func NewS3CachingAdapter(cfg *config.Configuration) (*S3CachingAdapter, error) {
	configuration := GetCachingConfiguration(cfg)
	if !configuration.enabled() {
		fmt.Fprintf(os.Stderr, "Found no caching configuration for this repository. Not caching anything.\n")
		return nil, nil
	}
	jsonConfiguration, err := json.Marshal(configuration)
	if err == nil {
		fmt.Fprintf(os.Stderr, "Using S3 caching adapter with configuration %s\n", jsonConfiguration)
	} else {
		fmt.Fprintf(os.Stderr, "Using S3 caching adapter with configuration %+v\n", configuration)
	}
	client, err := configuration.newClient()
	if err != nil {
		return nil, err
	}
	return &S3CachingAdapter{
		client:        client,
		configuration: configuration,
	}, nil
}

func (a *S3CachingAdapter) exists(ctx context.Context, oid string, size int64) (bool, error) {
	object, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: a.configuration.Bucket,
		Key:    aws.String(fmt.Sprintf("%s/%s", *a.configuration.Prefix, oid)),
	})
	if err != nil {
		var responseError *awshttp.ResponseError
		if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	if *object.ContentLength != size {
		return false, fmt.Errorf("object size mismatch: expected %d, got %d", size, *object.ContentLength)
	}
	return true, nil
}

func (a *S3CachingAdapter) Download(dest string, oid string, size int64, progressCallback func(bytesSoFar int64, bytesSinceLast int64)) (bool, error) {
	if ok, err := a.exists(context.Background(), oid, size); !ok {
		return false, err
	}

	// Download the object from the S3 bucket
	resp, err := a.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: a.configuration.Bucket,
		Key:    aws.String(fmt.Sprintf("%s/%s", *a.configuration.Prefix, oid)),
	})
	if err != nil {
		return false, fmt.Errorf("failed to download object: %v", err)
	}
	defer resp.Body.Close()

	// Create the destination file
	file, err := os.Create(dest)
	if err != nil {
		return false, fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write resp.Body to the file with progress indicator
	_, err = io.Copy(file, &progressReader{reader: resp.Body, progressCallback: progressCallback})
	if err != nil {
		return false, fmt.Errorf("failed to write to file: %v", err)
	}

	return true, nil
}

func (a *S3CachingAdapter) Upload(source string, oid string, size int64) (bool, error) {
	uploaded, err := a.exists(context.Background(), oid, size)
	if uploaded && err == nil {
		return false, nil
	}

	// Open the source file
	file, err := os.Open(source)
	if err != nil {
		return false, fmt.Errorf("failed to open source file: %v", err)
	}
	defer file.Close()

	// Upload the file to the S3 bucket
	_, err = a.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: a.configuration.Bucket,
		Key:    aws.String(fmt.Sprintf("%s/%s", *a.configuration.Prefix, oid)),
		Body:   file,
	})
	if err != nil {
		return false, fmt.Errorf("failed to upload file to S3: %v", err)
	}

	return true, nil
}
