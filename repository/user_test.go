package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"kedubak/model"
	"log"
	"testing"
	"time"
)

func newMongoClient() *mongo.Client {
	mongoTestClient, err := mongo.Connect(context.Background(),
		options.Client().ApplyURI("mongodb+srv://juleslordet69:root@taker.lwsikv1.mongodb.net/?retryWrites=true&w=majority&appName=taker"))

	if err != nil {
		log.Fatal("error while connecting mongodb", err)
	}

	log.Println("mongodb succesfully connected.")

	err = mongoTestClient.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal("ping failed | ", err)
	}

	log.Println("ping success | ")

	return mongoTestClient
}

func TestMongoOperations(t *testing.T) {
	mongoTestClient := newMongoClient()
	defer mongoTestClient.Disconnect(context.Background())

	user1 := primitive.NewObjectID()
	// user2 := primitive.NewObjectID()

	coll := mongoTestClient.Database("Takerdb").Collection("user")

	userRepo := UserRepo{MongoColletion: coll}

	// Insert User
	t.Run("Insert User 1", func(t *testing.T) {
		user := model.User{
			CreatedAt: time.Now(),
			Email:      "email",
			FirstName:  "prénom",
			LastName:   "nom",
			Password:   "root",
			LastUpVote: time.Now().Add(-1 * time.Minute),
			ID:         user1,
		}

		result, err := userRepo.InsertUser(&user)

		if err != nil {
			t.Fatal("insert 1 operation failed", err)
		}

		t.Log("Insert 1 successful", result)
	})

	// Find User by ID
	t.Run("Get user 1", func(t *testing.T) {
		result, err := userRepo.FindUserByID(user1)

		if err != nil {
			t.Fatal("operation by id operation failed", err)
		}

		t.Log("user 1", result.FirstName)
	})

	// delete User by ID
	// t.Run("Delete user ID", func(t *testing.T) {
		// result, err := userRepo.DeleteUserByID(user1)
// 
		// if err != nil {
			// t.Fatal("operation by id operation failed", err)
		// }
// 
		// t.Log("delete : ", result)
	// })
}
