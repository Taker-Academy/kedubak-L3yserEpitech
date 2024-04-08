package repository

import(
	"context"
	"fmt"
	"kedubak/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type UserRepo struct {
	MongoColletion *mongo.Collection
}

func (r *UserRepo) InsertUser(user *model.User) (interface{}, error) {
    result, err := r.MongoColletion.InsertOne(context.Background(), user)
    if err != nil {
        return nil, err
    }
    return result.InsertedID, nil
}

func (r *UserRepo) FindUserByID(userID primitive.ObjectID) (*model.User, error) {
    var user model.User

    err := r.MongoColletion.FindOne(context.Background(), 
            bson.D{{Key: "_id", Value: userID}}).Decode(&user)

    if err != nil {
        return nil, err
    }

    return &user, nil
}

func (r *UserRepo) FindAllUser() ([]model.User, error) {
	result, err := r.MongoColletion.Find(context.Background(), bson.D{})

	if err != nil {
		return nil, err
	}

	var users []model.User
	err = result.All(context.Background(), &users)
	if err != nil {
		return nil, fmt.Errorf("results decode error %s", err.Error())
	}
	
	return users, nil
}

func (r *UserRepo) UpdateUserByID(userID primitive.ObjectID, updateUser *model.User) (int64, error) {
    result, err := r.MongoColletion.UpdateOne(context.Background(),
        bson.D{{Key: "_id", Value: userID}},
        bson.D{{Key: "$set", Value: updateUser}})

    if err != nil {
        return 0, err
    }

    return result.ModifiedCount, nil
}

func (r *UserRepo) DeleteUserByID(userID primitive.ObjectID) (int64, error) {
    result, err := r.MongoColletion.DeleteOne(context.Background(),
        bson.D{{Key: "_id", Value: userID}})
    
    if err != nil {
        return 0, err
    }
    
    return result.DeletedCount, nil
}

func (r *UserRepo) DeleteAllUser() (int64, error) {
	result, err := r.MongoColletion.DeleteMany(context.Background(), bson.D{})

	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}