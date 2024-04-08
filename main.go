package main

import (
	"context"
	"kedubak/usecase"
	"log"
	"net/http"
	"os"

	// "github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var mongoClient *mongo.Client

func init() {
	// load .envfile
	err := godotenv.Load()
	if err != nil {
		log.Fatal("env load error", err)
	}
	log.Println("env file loaded")
	mongoClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGO_URI")))

	if err != nil {
		log.Fatal("connection error", err)
	}

	mongoClient.Ping(context.Background(), readpref.Primary())

	if err != nil {
		log.Fatal("ping failed", err)
	}

	log.Println("mongo connected")
}

func main() {
	// close mongo connection
	defer mongoClient.Disconnect(context.Background())

	coll := mongoClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("COLLECTION_NAME"))

	// create user service
	userService := usecase.UserService{MongoColletion: coll}

	r := mux.NewRouter()

	r.HandleFunc("/health", healthHandler).Methods(http.MethodGet)

	r.HandleFunc("/user", userService.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/user/{id}", userService.GetUserByID).Methods(http.MethodGet)
	r.HandleFunc("/user", userService.GetAllUser).Methods(http.MethodPost)
	r.HandleFunc("/user/{id}", userService.UpdateUserByID).Methods(http.MethodPut)
	r.HandleFunc("/user/{id}", userService.DeleteUserByID).Methods(http.MethodDelete)
	r.HandleFunc("/user", userService.DeleteAllUser).Methods(http.MethodDelete)
	
	log.Println("server is running on 4444")
	http.ListenAndServe(":4444", r)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("running..."))
}
