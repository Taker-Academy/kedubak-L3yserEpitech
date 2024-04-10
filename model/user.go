package model

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	CreatedAt  time.Time          `bson:"createdAt"`
	Email      string             `bson:"email"`
	FirstName  string             `bson:"firstName"`
	LastName   string             `bson:"lastName"`
	Password   string             `bson:"password"`
	LastUpVote time.Time          `bson:"lastUpVote"`
	ID         primitive.ObjectID `bson:"_id,omitempty"`
}

type Comment struct {
	CreatedAt time.Time          `bson:"createdAt"`
	FirstName string             `bson:"firstName"`
	ID        primitive.ObjectID `bson:"_id"`
	Content   string             `bson:"content"`   
}

type Post struct {
	CreatedAt time.Time          `bson:"createdAt"` 
	UserID   primitive.ObjectID `bson:"userId"`    
	FirstName string            `bson:"firstName"` 
	Title    string             `bson:"title"`     
	Content  string             `bson:"content"`   
	Comments []Comment          `bson:"comments"`  
	UpVotes  []string           `bson:"upVotes"`
}
