package s3

import (
	"errors"
	"time"

	"github.com/redhatinsights/insights-ingress-go/internal/stage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	sess     *session.Session
	uploader *s3manager.Uploader
	client   *s3.S3
)

// Stager provides the mechanism to stage a payload via aws S3
type Stager struct {
	Bucket   string
	Sess     *session.Session
	Uploader *s3manager.Uploader
	Client   *s3.S3
}

func init() {
	sess = session.Must(session.NewSession())
	uploader = s3manager.NewUploader(sess)
	client = s3.New(sess)
}

// Stage stores the file in s3 and returns a presigned url
func (s *Stager) Stage(in *stage.Input) (string, error) {
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(in.Key),
		Body:   in.Payload,
		Metadata: map[string]*string{
			"requestID": aws.String(in.Key),
			"account":   aws.String(in.Account),
			"org":       aws.String(in.OrgId),
		},
	})
	if err != nil {
		return "", errors.New("Failed to upload to s3: " + err.Error())
	}

	return s.GetURL(in.Key)
}

// GetURL gets a Presigned URL from S3
func (s *Stager) GetURL(requestID string) (string, error) {
	req, _ := client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(requestID),
	})
	url, err := req.Presign(24 * time.Hour)
	if err != nil {
		return "", errors.New("Failed to generate persigned url: " + err.Error())
	}

	return url, nil
}
