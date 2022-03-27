package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/gographics/imagick.v2/imagick"
)

func editImage(body []byte) []byte {
	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	dw := imagick.NewDrawingWand()
	pw := imagick.NewPixelWand()

	pw.SetColor("#FFFFFF")

	dw.SetFillColor(pw)
	dw.SetFontSize(72)
	dw.Annotation(150, 210, "LGTM")

	mw.ReadImageBlob(body)
	mw.ResizeImage(500, 400, imagick.FILTER_POINT, 0)

	mw.DrawImage(dw)

	return mw.GetImageBlob()
}

func handler(ctx context.Context, s3Event events.S3Event) (string, error) {
	sess := session.Must(session.NewSession())
	svc := s3.New(sess)
	for _, record := range s3Event.Records {
		s3Record := record.S3
		fmt.Printf("Bucket = %s, Key = %s \n", s3Record.Bucket.Name, s3Record.Object.Key)

		originBucketName := aws.String(s3Record.Bucket.Name)
		originObjectKey := aws.String(s3Record.Object.Key)
		obj, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: originBucketName,
			Key:    originObjectKey,
		})
		if err != nil {
			return "", err
		}

		b := obj.Body
		defer b.Close()

		body, err := ioutil.ReadAll(b)
		if err != nil {
			return "", err
		}

		// Put Edit Image
		svc.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("UPLOAD_BUCKET")),
			Key:    aws.String(s3Record.Object.Key),
			Body:   strings.NewReader(string(editImage(body))),
		})

		// Remove Original File
		svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: originBucketName,
			Key:    originObjectKey,
		})
	}

	return "ok", nil
}

func main() {
	lambda.Start(handler)
}
