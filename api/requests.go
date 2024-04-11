package api

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/minio/minio-go/v7"
	u "gopractice/tools"
	"os"
	"strconv"
	"strings"
)

// JWT claims struct
type Token struct {
	UserId uint
	jwt.StandardClaims
}

// a struct to upload file
type Photo struct {
	gorm.Model
	Sort        int    `json:"sort"`
	Description string `json:"description"`
	ProductCode string `json:"productcode"`
	File        string `json:"file"`
	Uuid        string `json:"uuid"`
	Size        string `json:"size"`
}

type qphoto struct {
	Uuid        string `json:"uuid"`
	File        string `json:"file"`
	Description string `json:"description"`
	Size        string `json:"size"`
}

func (photo *Photo) Upload() (map[string]interface{}, error) {

	if resp, ok := photo.Validate(1); !ok {
		return resp, nil
	}

	db := GetDB()

	// Update
	result := db.Model(&photo).Where("product_code = ?", photo.ProductCode).Where("sort = ?", photo.Sort).Updates(&photo)
	if result.Error != nil {
		return u.Message(false, "fail to update data"), result.Error
	}
	if result.RowsAffected <= 0 {
		err := db.Create(&photo).Error
		if err != nil {
			return u.Message(false, "fail to create new data"), err
		}
	}

	resp := u.Message(true, "success")
	resp["photo"] = photo
	return resp, nil
}

// Validate incoming photo details...
func (photo *Photo) Validate(val int) (map[string]interface{}, bool) {

	if len(photo.Description) == 0 {
		return u.Message(false, "Description is empty"), false
	}

	if photo.Sort == 0 {
		return u.Message(false, "Sort is empty"), false
	}

	if len(photo.ProductCode) == 0 {
		return u.Message(false, "Product code is empty"), false
	}

	if val == 1 {
		if len(photo.File) == 0 {
			return u.Message(false, "File name is empty"), false
		}
	}

	if len(photo.ProductCode) > 20 {
		return u.Message(false, "Product code must before 20 symbols"), false
	}

	return u.Message(false, "Requirement passed"), true
}

func GetList(product_code string) []*qphoto {

	db := GetDB()

	// Select product_code from photos where product_code = @
	photos := make([]*qphoto, 0)
	db.Table("photos").Where("product_code = ?", product_code).Order("sort").Find(&photos)
	return photos
}

func (photo *Photo) Update() (map[string]interface{}, error) {

	if resp, ok := photo.Validate(0); !ok {
		return resp, nil
	}

	db := GetDB()

	newSort := strconv.Itoa(photo.Sort)
	newDescription := photo.Description
	// Find first record according to condition "uuid = ?"
	result1 := db.Model(&photo).Where("uuid = ?", photo.Uuid).First(&photo)
	if result1.RowsAffected <= 0 {
		return u.Message(false, "Record not found"), result1.Error
	}

	preSort := strconv.Itoa(photo.Sort)
	extension := photo.File[strings.LastIndex(photo.File, "."):]
	fileName := photo.ProductCode + "_" + preSort + extension
	newFileName := photo.ProductCode + "_" + newSort + extension

	boolValue, err := strconv.ParseBool(os.Getenv("useS3"))
	if err != nil {
		return u.Message(false, "Can't convert str to bool type"), err
	}
	if boolValue {
		minioClient := GetMinio()
		// Source object
		srcOpts := minio.CopySrcOptions{
			Bucket: Minio_BucketName,
			Object: fileName,
		}
		// Destination object
		dstOpts := minio.CopyDestOptions{
			Bucket: Minio_BucketName,
			Object: newFileName,
		}
		// Copy object call

		_, err := minioClient.CopyObject(context.Background(), dstOpts, srcOpts)
		//fmt.Println(res)
		if err != nil {
			return u.Message(false, "MinIO copy error"), err
		}
		opts := minio.RemoveObjectOptions{
			GovernanceBypass: true,
		}
		err = minioClient.RemoveObject(context.Background(), Minio_BucketName, fileName, opts)
		if err != nil {
			return u.Message(false, "MinIO remove error"), err
		}

	} else {
		err := os.Rename(Path_to_photos+fileName, Path_to_photos+newFileName)
		if err != nil {
			return u.Message(false, "fail to update file name"), err
		}
	}
	//Update
	photo.File = newFileName
	if photo.Sort, err = strconv.Atoi(newSort); err != nil {
		return u.Message(false, "fail to convert Atoi"), err
	}
	photo.Description = newDescription
	result := result1.Updates(photo)
	if result.Error != nil {
		return u.Message(false, "fail to update data"), err
	}

	resp := u.Message(true, "success")
	resp["photo"] = photo
	return resp, nil

}
