package api

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
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
	Description string `json:"description"`
	Sort        int    `json:"sort"`
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
	result := db.Model(photo).Where("product_code = ?", photo.ProductCode).Where("sort = ?", photo.Sort).Updates(photo)
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

	if len(string(photo.Sort)) == 0 {
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

	// Find first record according to condition "uuid = ?"
	result1 := db.Model(photo).Where("uuid = ?", photo.Uuid).First(&photo)

	fileName := photo.File
	extension := fileName[strings.LastIndex(fileName, "."):]
	newFileName := Path_to_photos + photo.ProductCode + "_" + strconv.Itoa(photo.Sort) + extension
	err := os.Rename(photo.File, newFileName)
	if err != nil {
		return u.Message(false, "fail to update file name"), err
	}
	//Update
	photo.File = newFileName
	result := result1.Updates(photo)
	if result.Error != nil {
		return u.Message(false, "fail to update data"), err
	}

	resp := u.Message(true, "success")
	resp["photo"] = photo
	return resp, nil

}
