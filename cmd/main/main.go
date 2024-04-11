package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gopractice/api"
	"gopractice/controllers"
	"gopractice/deployments/app"
	"log"
	"net/http"
	"os"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	err := godotenv.Load()
	if err != nil {
		return err
	}

	err = api.InitBase()
	if err != nil {
		return err
	}

	err = api.InitMinio()
	if err != nil {
		return err
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/products/{product_code}/upload", controllers.Upload).Methods("POST")
	router.HandleFunc("/api/products/{product_code}/list", controllers.List).Methods("GET")
	router.HandleFunc("/api/products/{product_code}/update", controllers.Update).Methods("PUT")
	router.HandleFunc("/api/products/{product_code}/getFile", controllers.GetPhoto).Methods("GET")

	router.Use(app.JwtAuthentication) //attach JWT auth middleware

	port := os.Getenv("port")
	if port == "" {
		port = "8001" //localhost
	}

	fmt.Println("Listening port: " + port)

	err = http.ListenAndServe(":"+port, router) //Launch the app
	if err != nil {
		return err
	}

	return nil
}
