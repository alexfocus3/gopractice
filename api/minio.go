package api

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"os"
	"strconv"
	"time"
)

var client *minio.Client
var Minio_BucketName, Minio_BucketCachName string

func bucketExists(client *minio.Client, bucketName string) error {

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer cancel()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: "ru", ObjectLocking: false})
		if err != nil {
			return err
		}
	}
	return nil
}

func InitMinio() error {

	b, err := strconv.ParseBool(os.Getenv("minio_useSSL"))
	if err != nil {
		return err
	}
	minioclient, err := minio.New(os.Getenv("minio_endpoint"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("minio_accessKeyID"), os.Getenv("minio_secretAccessKey"), ""),
		Secure: b})
	if err != nil {
		return err
	}
	client = minioclient
	Minio_BucketName = os.Getenv("minio_BucketName")
	if err := bucketExists(client, Minio_BucketName); err != nil {

		//err = minioClient.MakeBucket(context.Background(), "mybucket", minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: true})
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}

		return err
	}
	Minio_BucketCachName = os.Getenv("minio_BucketCachName")
	if err := bucketExists(client, Minio_BucketCachName); err != nil {
		return err
	}
	return err
}

func GetMinio() *minio.Client {
	return client
}
