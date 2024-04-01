package api

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
	"os"
)

var db *gorm.DB
var Path_to_photos string

func InitBase() error {

	e := godotenv.Load()
	if e != nil {
		return e
	}

	username := os.Getenv("db_user")
	password := os.Getenv("db_pass")
	dbName := os.Getenv("db_name")
	dbHost := os.Getenv("db_host")

	dbUri := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, username, dbName, password)
	fmt.Println(dbUri)

	conn, err := gorm.Open("postgres", dbUri)
	if err != nil {
		return err
	}

	db = conn
	db.Debug().AutoMigrate(&Photo{})

	Path_to_photos = os.Getenv("path_to_photos")

	return nil
}

func GetDB() *gorm.DB {
	return db
}
