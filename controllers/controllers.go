package controllers

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/nfnt/resize"
	"gopractice/api"
	u "gopractice/tools"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var Upload = func(w http.ResponseWriter, r *http.Request) {

	err := r.ParseMultipartForm(1024 * 1024 * 1) //1 mb
	if err != nil {
		resp := u.Message(false, err.Error())
		u.Respond(w, resp)
		return
	}
	file, fileHeader, err := r.FormFile("file") //retrieve the file from form data
	//replace file with the product code and sort your sent your image with
	if err != nil {
		resp := u.Message(false, "Can't retrieve the file from form data: "+err.Error())
		u.Respond(w, resp)
		return
	}
	defer file.Close() //close the file when we finish

	fileName := fileHeader.Filename
	extension := fileName[strings.LastIndex(fileName, "."):]
	productId := mux.Vars(r)["product_code"]
	sortNumber := r.FormValue("sort")
	description := r.FormValue("description")

	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)

	//this is path which we want to store the file
	newFileName := productId + "_" + sortNumber + extension
	path := api.Path_to_photos + newFileName

	photo := &api.Photo{}
	photo.Description = description
	sort1, err := strconv.Atoi(sortNumber)
	if err != nil {
		resp := u.Message(false, "Can't convert Sort to int type: "+err.Error())
		u.Respond(w, resp)
		return
	}
	photo.Sort = sort1
	photo.ProductCode = productId
	photo.File = path
	photo.Uuid = uuid.New().String()

	boolValue, err := strconv.ParseBool(os.Getenv("useS3"))
	if err != nil {
		resp := u.Message(false, "Can't convert str to bool type: "+err.Error())
		u.Respond(w, resp)
		return
	}
	if boolValue {
		photo.File = "minio:/" + api.Minio_BucketName + "/" + newFileName
	}

	im, _, err := image.DecodeConfig(file)
	if err != nil {
		resp := u.Message(false, err.Error())
		u.Respond(w, resp)
		return
	}
	photo.Size = strconv.Itoa(im.Width) + "x" + strconv.Itoa(im.Height)
	resp, err := photo.Upload() //Uploading photo
	if err == nil {
		if boolValue {
			minioClient := api.GetMinio()
			_, err = minioClient.PutObject(context.Background(), api.Minio_BucketName, newFileName, file, fileHeader.Size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
			if err != nil {
				resp := u.Message(false, "MinIO download error: "+err.Error())
				u.Respond(w, resp)
				return
			}
		} else {
			f, err := os.Create(path)
			if err != nil {
				resp := u.Message(false, "Can't create epmty file: "+err.Error())
				u.Respond(w, resp)
				return
			}
			defer f.Close()
			_, err = file.Seek(0, io.SeekStart)
			if err != nil {
				resp := u.Message(false, "Can't 'Seek' file: "+err.Error())
				u.Respond(w, resp)
				return
			}
			_, err = io.Copy(f, file)
			if err != nil {
				resp := u.Message(false, "Can't copy file: "+err.Error())
				u.Respond(w, resp)
				return
			}
		}
	} else {
		resp := u.Message(false, "Fail upload file: "+err.Error())
		u.Respond(w, resp)
		return
	}
	u.Respond(w, resp)
}

var List = func(w http.ResponseWriter, r *http.Request) {

	productId := mux.Vars(r)["product_code"]
	data := api.GetList(productId)
	resp := u.Message(true, "success")
	jres, err := json.Marshal(data)
	if err != nil {
		resp := u.Message(false, "Can't transform slice to json")
		u.Respond(w, resp)
		return
	}
	resp["data"] = string(jres)
	u.Respond(w, resp)
}

var Update = func(w http.ResponseWriter, r *http.Request) {

	type message struct {
		Uuid        string `json:"uuid"`
		Description string `json:"description"`
		Sort        string `json:"sort"`
	}
	var m message

	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		resp := u.Message(false, err.Error())
		u.Respond(w, resp)
		return
	}

	photo := &api.Photo{}
	if photo.Sort, err = strconv.Atoi(m.Sort); err != nil {
		resp := u.Message(false, "Can't transform Atoi: "+m.Sort)
		u.Respond(w, resp)
		return
	}
	photo.Uuid = m.Uuid
	photo.Description = m.Description

	productId := mux.Vars(r)["product_code"]
	photo.ProductCode = productId

	resp, err := photo.Update() //Uploading photo
	if err != nil {
		resp := u.Message(false, "Can't update data: "+err.Error())
		u.Respond(w, resp)
		return
	}

	u.Respond(w, resp)
}

var GetPhoto = func(w http.ResponseWriter, r *http.Request) {

	productId := mux.Vars(r)["product_code"]
	data := api.GetList(productId)
	if len(data) == 0 {
		resp := u.Message(false, "Can't find that file")
		u.Respond(w, resp)
	}
	file := data[0].File
	fileName := filepath.Base(file)
	fileName2 := fileName[:strings.LastIndex(fileName, ".")]
	extension := fileName[strings.LastIndex(fileName, "."):]
	size := r.URL.Query().Get("size")
	sizes := strings.Split(size, "x")
	if len(sizes) < 2 {
		resp := u.Message(false, "Can't get sizes from request")
		u.Respond(w, resp)
		return
	}

	size1, err := strconv.Atoi(sizes[0])
	if err != nil {
		resp := u.Message(false, "Can't get sizes 1 from request")
		u.Respond(w, resp)
		return
	}
	size11 := uint(size1)
	if size11 == 0 {
		resp := u.Message(false, "Can't get sizes 1 from request")
		u.Respond(w, resp)
		return
	}

	size2, err := strconv.Atoi(sizes[1])
	if err != nil {
		resp := u.Message(false, "Can't get sizes 2 from request")
		u.Respond(w, resp)
		return
	}
	size22 := uint(size2)
	if size22 == 0 {
		resp := u.Message(false, "Can't get sizes 2 from request")
		u.Respond(w, resp)
		return
	}
	fileName3 := fileName2 + "_" + size + extension

	boolValue, err := strconv.ParseBool(os.Getenv("useS3"))
	if err != nil {
		resp := u.Message(false, "Can't convert str to bool type")
		u.Respond(w, resp)
		return
	}

	pathToSendFile := api.Path_to_cach_photos + fileName3
	if boolValue {
		pathToSendFile = os.TempDir() + string(filepath.Separator) + fileName3
	}

	isCachFile := true
	if boolValue {
		minioClient := api.GetMinio()
		err := minioClient.FGetObject(context.Background(), api.Minio_BucketCachName, fileName3, pathToSendFile, minio.GetObjectOptions{})
		if err != nil {
			errResponse := minio.ToErrorResponse(err)
			if errResponse.Code == s3.ErrCodeNoSuchKey {
				isCachFile = false
				err := minioClient.FGetObject(context.Background(), api.Minio_BucketName, fileName, pathToSendFile, minio.GetObjectOptions{})
				if err != nil {
					resp := u.Message(false, err.Error())
					u.Respond(w, resp)
					return
				}
			} else {
				resp := u.Message(false, err.Error())
				u.Respond(w, resp)
				return
			}
		}
		file = pathToSendFile
		cleanup := func() {
			os.Remove(file)
		}
		defer cleanup()
	} else {
		_, err = os.Stat(fileName3)
		if err != nil {
			if os.IsNotExist(err) {
				isCachFile = false
			} else {
				resp := u.Message(false, err.Error())
				u.Respond(w, resp)
				return
			}
		}
	}

	if !isCachFile {
		// open file (check if exists)
		imgIn, err := os.Open(file)
		if err != nil {
			resp := u.Message(false, "Unable to open file: "+file)
			u.Respond(w, resp)
			return
		}
		imgJpg, err := jpeg.Decode(imgIn)
		if err != nil {
			resp := u.Message(false, "Unable to decode file")
			u.Respond(w, resp)
			return
		}
		defer imgIn.Close()
		imgJpg = resize.Resize(size11, size22, imgJpg, resize.Bicubic) // <-- Собственно изменение размера картинки
		imgOut, err := os.Create(pathToSendFile)
		if err != nil {
			resp := u.Message(false, "Unable to create cach file")
			u.Respond(w, resp)
			return
		}
		err = jpeg.Encode(imgOut, imgJpg, nil)
		if err != nil {
			resp := u.Message(false, "Unable to encode file")
			u.Respond(w, resp)
			return
		}
		defer imgOut.Close()
	}

	if boolValue {
		minioClient := api.GetMinio()
		_, err := minioClient.FPutObject(context.Background(), api.Minio_BucketCachName, fileName3, pathToSendFile, minio.PutObjectOptions{
			ContentType: "application/csv",
		})
		if err != nil {
			resp := u.Message(false, err.Error())
			u.Respond(w, resp)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")

	// force a download with the content- disposition field
	w.Header().Set("Content-Disposition", "attachment; filename="+pathToSendFile)
	// serve file out.
	http.ServeFile(w, r, pathToSendFile)

	resp := u.Message(true, "Send file ok")
	u.Respond(w, resp)
}
