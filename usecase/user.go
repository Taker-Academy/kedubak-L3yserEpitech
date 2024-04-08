package usecase

import (
	"encoding/json"
	"kedubak/model"
	"kedubak/repository"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserService struct {
	MongoColletion *mongo.Collection
}

type Response struct {
	Data  	interface{}		`json:"data,omitempty"`
	Error 	string      	`json:"error,omitempty"`
}

func (svc *UserService) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}
	defer json.NewEncoder(w).Encode(res)

	var user model.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("invalid body ", err)
		res.Error = err.Error()
		return
	}

	// assign new user id
	user.UserID = uuid.NewString()
	repo := repository.UserRepo{MongoColletion: svc.MongoColletion}

	// insert user
	insertID, err := repo.InsertUser(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("insert error ", err)
		res.Error = err.Error()
		return
	}

	res.Data = user.UserID
	w.WriteHeader(http.StatusOK)

	log.Println("user inserted with ID ", insertID, user)	
}	

func (svc *UserService) GetUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}
	defer json.NewEncoder(w).Encode(res)

	// get user id
	userID := mux.Vars(r)["id"]
	log.Println("user id ", userID)

	repo := repository.UserRepo{MongoColletion: svc.MongoColletion}

	user, err := repo.FindUserByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("error: ", err)
		res.Error = err.Error()
		return
	}

	res.Data = user
	w.WriteHeader(http.StatusOK)
}

func (svc *UserService) GetAllUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}
	defer json.NewEncoder(w).Encode(res)

	// get user id
	userID := mux.Vars(r)["id"]
	log.Println("user id ", userID)

	repo := repository.UserRepo{MongoColletion: svc.MongoColletion}

	user, err := repo.FindAllUser()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("error: ", err)
		res.Error = err.Error()
		return
	}

	res.Data = user
	w.WriteHeader(http.StatusOK)
}

func (svc *UserService) UpdateUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}
	defer json.NewEncoder(w).Encode(res)

	// get user id
	userID := mux.Vars(r)["id"]
	log.Println("user id ", userID)

	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("invalid user id")
		res.Error = "invalid user id"
		return
	}

	var user model.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("invalid body ", err)
		res.Error = err.Error()
		return
	}
	
	user.UserID = userID
	repo := repository.UserRepo{MongoColletion: svc.MongoColletion}
	count, err := repo.UpdateUserByID(userID, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("error: ", err)
		res.Error = err.Error()
		return
	}

	res.Data = count
	w.WriteHeader(http.StatusOK)
}

func (svc *UserService) DeleteUserByID(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("Content-Type", "application/json")

    res := &Response{}
    defer json.NewEncoder(w).Encode(res)

    // Récupère l'userID depuis les paramètres de la requête
    userIDStr := mux.Vars(r)["id"]
    log.Println("user id ", userIDStr)

    // Convertis l'userID de string à primitive.ObjectID
    userID, err := primitive.ObjectIDFromHex(userIDStr)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        log.Println("Invalid user ID format: ", err)
        res.Error = "Invalid user ID format"
        return
    }

    repo := repository.UserRepo{MongoColletion: svc.MongoColletion}

    // Appelle DeleteUserByID avec un ObjectID
    count, err := repo.DeleteUserByID(userID)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        log.Println("error: ", err)
        res.Error = err.Error()
        return
    }

    res.Data = count
    w.WriteHeader(http.StatusOK)
}

func (svc *UserService) DeleteAllUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}
	defer json.NewEncoder(w).Encode(res)

	// get user id
	userID := mux.Vars(r)["id"]
	log.Println("user id ", userID)

	repo := repository.UserRepo{MongoColletion: svc.MongoColletion}

	count, err := repo.DeleteAllUser()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("error: ", err)
		res.Error = err.Error()
		return
	}

	res.Data = count
	w.WriteHeader(http.StatusOK)
}