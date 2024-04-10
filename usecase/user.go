package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"kedubak/model"
	"kedubak/repository"
	"log"
	"net/http"
	"strings"
	"time"
	"os"
)

func GenerateJWT(userEmail string, userID string) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["userEmail"] = userEmail
	claims["userID"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(os.Getenv("TOKEN_KEY")))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

type UserService struct {
	MongoColletion *mongo.Collection
}

type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

type APIResponse struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

type UserData struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

func (svc *UserService) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	log.Println("Tentative de création d'un nouvel utilisateur")

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: err.Error()})
		return
	}

	user.ID = primitive.NewObjectID()
	token, err := GenerateJWT(user.Email, "x9e3ea")
	if err != nil {
		fmt.Println("Erreur lors de la génération du token:", err)
		return
	}
	repo := repository.UserRepo{MongoColletion: svc.MongoColletion}
	user.CreatedAt = time.Now()
	user.LastUpVote = time.Now().Add(-1 * time.Minute)
	if _, err := repo.InsertUser(&user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Erreur lors de l'insertion de l'utilisateur"})
		return
	}

	userData := UserData{
		Token: token,
		User:  user,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(APIResponse{OK: true, Data: userData})
	log.Println("Utilisateur inséré avec succès", user)
}

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (svc *UserService) LoginUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	log.Println("Tentative de login")

	var loginReq UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Invalid request body"})
		return
	}

	var user model.User
	if err := svc.MongoColletion.FindOne(context.Background(), bson.M{"email": loginReq.Email}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Invalid credentials"})
		} else {
			log.Printf("Error finding user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Internal server error"})
		}
		return
	}

	// Puisque vous n'utilisez pas de hachage de mot de passe, comparez directement
	if user.Password != loginReq.Password {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Invalid credentials"})
		return
	}

	// Génération d'un token JWT
	token, err := GenerateJWT(user.Email, user.ID.Hex())
	if err != nil {
		log.Printf("Error generating token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Could not generate token"})
		return
	}

	// Réponse avec token et informations utilisateur
	userData := UserData{
		Token: token,
		User:  user,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{OK: true, Data: userData})
}

func (svc *UserService) GetInfoUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	userIDStr, err := ValidateJWT(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Invalid or missing token"})
		return
	}

	// Convertir userIDStr en ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Invalid user ID format"})
		return
	}

	repo := repository.UserRepo{MongoColletion: svc.MongoColletion}

	var user model.User
	if err := repo.MongoColletion.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "User not found"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Internal server error"})
		}
		return
	}

	// Réponse avec les informations de l'utilisateur
	responseData := map[string]string{
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{OK: true, Data: responseData})
}

type CustomClaims struct {
	UserID    string `json:"userID"`
	UserEmail string `json:"userEmail"`
	jwt.StandardClaims
}

type UserUpdateRequest struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"password,omitempty"`
}

func ValidateJWT(tokenString string) (string, error) {
	splitToken := strings.Split(tokenString, "Bearer ")
	if len(splitToken) != 2 {
		return "", errors.New("invalid or missing token")
	}
	tokenString = splitToken[1]

	var mySigningKey = []byte("votre_cle_secrete")

	// Parse le token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims.UserID, nil
	} else {
		return "", err
	}
}

func (svc *UserService) UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var updateReq UserUpdateRequest

	// Extraire le token JWT du header Authorization
	token := r.Header.Get("Authorization")
	userID, err := ValidateJWT(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or missing token"})
		return
	}

	// Décoder le corps de la requête
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Mettre à jour l'utilisateur dans la base de données
	update := bson.M{"$set": bson.M{}}
	if updateReq.FirstName != "" {
		update["$set"].(bson.M)["firstName"] = updateReq.FirstName
	}
	if updateReq.LastName != "" {
		update["$set"].(bson.M)["lastName"] = updateReq.LastName
	}
	if updateReq.Email != "" {
		update["$set"].(bson.M)["email"] = updateReq.Email
	}
	if updateReq.Password != "" {
		// Ici, vous devriez hasher le mot de passe avant de le stocker
		update["$set"].(bson.M)["password"] = updateReq.Password
	}

	// Assurez-vous d'avoir une fonction pour trouver l'utilisateur par ID et mettre à jour ses informations
	objID, err := primitive.ObjectIDFromHex(userID) // Supposant que userID est un string
	if err != nil {
		// Gérer l'erreur si la conversion échoue
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to convert userID to ObjectID"})
		return
	}

	// Envoie de `update` à la méthode `UpdateUserByID`
	repo := repository.UserRepo{MongoColletion: svc.MongoColletion}

	if _, err := repo.UpdateUserByID(context.Background(), objID, update); err != nil {
		log.Printf("Error updating user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
		return
	}

	// Répondre avec les informations mises à jour
	// Note: Dans une application réelle, vous voudrez peut-être récupérer l'utilisateur mis à jour de la DB pour renvoyer ses dernières infos
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":   true,
		"data": updateReq, // Attention: renvoyer le mot de passe n'est pas sécurisé
	})
}

func (svc *UserService) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	userIDStr, err := ValidateJWT(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Invalid or missing token"})
		return
	}

	repo := repository.UserRepo{MongoColletion: svc.MongoColletion}

	objID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Invalid user ID format"})
		return
	}

	// Récupérer les informations de l'utilisateur avant de supprimer pour la réponse
	var user model.User
	if err := repo.MongoColletion.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user); err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "User not found"})
		return
	}

	// Suppression de l'utilisateur
	count, err := repo.DeleteUserByID(context.Background(), objID)
	if err != nil || count == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Error deleting user"})
		return
	}

	// Préparer la réponse avec les informations de l'utilisateur
	response := map[string]interface{}{
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"removed":   true,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{OK: true, Data: response})
}

type PostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (svc *UserService) CreatePost(w http.ResponseWriter, r *http.Request) {
	log.Println("Création de post")
	w.Header().Add("Content-Type", "application/json")
	var user model.User

	// Extraction et validation du token JWT
	token := r.Header.Get("Authorization")
	userIDStr, err := ValidateJWT(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Invalid or missing token"})
		return
	}

	log.Println("userIDStr extrait du token:", userIDStr)
	// Conversion de l'ID utilisateur en ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Invalid user ID format"})
		return
	}

	log.Println("salut")
	// Décodage du corps de la requête pour obtenir le titre et le contenu du post
	var postReq PostRequest
	if err := json.NewDecoder(r.Body).Decode(&postReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Invalid request body"})
		return
	}
	log.Println(postReq.Title)

	// Création du post
	post := model.Post{
		CreatedAt: time.Now(),
		UserID:    userID,
		FirstName: user.FirstName,
		Title:     postReq.Title,
		Content:   postReq.Content,
		Comments:  []model.Comment{},
		UpVotes:   []string{},
	}

	// Insertion du post dans la base de données
	insertResult, err := svc.MongoColletion.InsertOne(context.Background(), post)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{OK: false, Error: "Internal server error"})
		return
	}

	// Récupération de l'ID du document inséré
	newPostID := insertResult.InsertedID

	// Répondre avec les informations du post créé, incluant le nouvel ID
	post.UserID = newPostID.(primitive.ObjectID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(APIResponse{OK: true, Data: post})
}