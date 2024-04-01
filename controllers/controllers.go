package controllers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

	r.ParseMultipartForm(1024 * 1024 * 1)       //1 mb
	file, fileHeader, err := r.FormFile("file") //retrieve the file from form data
	//replace file with the product code and sort your sent your image with
	if err != nil {
		resp := u.Message(false, "Can't retrieve the file from form data")
		u.Respond(w, resp)
		return
	}
	fileName := fileHeader.Filename
	extension := fileName[strings.LastIndex(fileName, "."):]
	productId := mux.Vars(r)["product_code"]
	sortNumber := r.FormValue("sort")
	description := r.FormValue("description")

	defer file.Close() //close the file when we finish

	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	//image.RegisterFormat("jpg", "jpg", jpeg.Decode, jpeg.DecodeConfig)

	//this is path which we want to store the file
	path := api.Path_to_photos + productId + "_" + sortNumber + extension

	photo := &api.Photo{}
	//Path := r.URL.Path
	photo.Description = description
	sort1, err := strconv.Atoi(sortNumber)
	if err != nil {
		resp := u.Message(false, "Can't convert Sort to int type")
		u.Respond(w, resp)
		return
	}
	photo.Sort = sort1
	photo.ProductCode = productId
	photo.File = path
	photo.Uuid = uuid.New().String()

	f, err := os.Open(path)
	if err != nil {
		resp := u.Message(false, "Can't open file")
		u.Respond(w, resp)
		return
	}
	defer f.Close()
	im, _, err := image.DecodeConfig(f)
	photo.Size = strconv.Itoa(im.Width) + "x" + strconv.Itoa(im.Height)
	resp, err := photo.Upload() //Uploading photo
	if err == nil {
		io.Copy(f, file)
	} else {
		resp := u.Message(false, "Fail copy file")
		u.Respond(w, resp)
		return
	}

	u.Respond(w, resp)
	return
}

var List = func(w http.ResponseWriter, r *http.Request) {

	productId := mux.Vars(r)["product_code"]
	data := api.GetList(productId)

	//image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	//
	//var doNotOpenedFiles string
	//var doNotDecodedFiles string
	//for _, value := range data {
	//
	//	if reader, err := os.Open(value.File); err == nil {
	//		defer reader.Close()
	//		_, _, err := image.DecodeConfig(reader)
	//		if err != nil {
	//			//fmt.Fprintf(os.Stderr, "%s: %v\n", imgFile.Name(), err)
	//			if len(doNotDecodedFiles) != 0 {
	//				doNotDecodedFiles += ", "
	//			}
	//			doNotDecodedFiles += value.File
	//			continue
	//		}
	//
	//		//fmt.Printf("%s %d %d\n", imgFile.Name(), im.Width, im.Height)
	//	} else {
	//		if len(doNotOpenedFiles) != 0 {
	//			doNotOpenedFiles += ", "
	//		}
	//		doNotOpenedFiles += value.File
	//		//Println("Impossible to open the file:", err)
	//	}
	//
	//}
	//if len(doNotDecodedFiles) != 0 {
	//	resp := u.Message(false, "Can't decode files:"+doNotDecodedFiles)
	//	u.Respond(w, resp)
	//	return
	//}
	//if len(doNotOpenedFiles) != 0 {
	//	resp := u.Message(false, "Can't open files:"+doNotDecodedFiles)
	//	u.Respond(w, resp)
	//	return
	//}
	resp := u.Message(true, "success")
	jres, err := json.Marshal(data)
	if err != nil {
		resp := u.Message(false, "Can't transform slice to json")
		u.Respond(w, resp)
		return
	}
	resp["data"] = string(jres)
	u.Respond(w, resp)
	return
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
		resp := u.Message(false, "Can't update data")
		u.Respond(w, resp)
		return
	}

	u.Respond(w, resp)
	return
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
	}

	size1, err := strconv.Atoi(sizes[0])
	if err != nil {
		resp := u.Message(false, "Can't get sizes 1 from request")
		u.Respond(w, resp)
	}
	size11 := uint(size1)
	if size11 == 0 {
		resp := u.Message(false, "Can't get sizes 1 from request")
		u.Respond(w, resp)
	}

	size2, err := strconv.Atoi(sizes[1])
	if err != nil {
		resp := u.Message(false, "Can't get sizes 2 from request")
		u.Respond(w, resp)
	}
	size22 := uint(size2)
	if size22 == 0 {
		resp := u.Message(false, "Can't get sizes 2 from request")
		u.Respond(w, resp)
	}
	fileName3 := fileName2 + "_" + size + extension

	isCachFile := false
	_, err = os.Stat(fileName3)
	if err != nil {
		if !os.IsNotExist(err) {
			isCachFile = true
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
		imgOut, err := os.Create(api.Path_to_photos + "cach/" + fileName3)
		if err != nil {
			resp := u.Message(false, "Unable to create cach file")
			u.Respond(w, resp)
			return
		}
		jpeg.Encode(imgOut, imgJpg, nil)
		defer imgOut.Close()
	}

	w.Header().Set("Content-Type", "application/json")

	// force a download with the content- disposition field
	w.Header().Set("Content-Disposition", "attachment; filename="+api.Path_to_photos+"cach/"+fileName3)

	// serve file out.
	http.ServeFile(w, r, api.Path_to_photos+"cach/"+fileName3)

	resp := u.Message(true, "Send file ok")
	u.Respond(w, resp)
	return
}
