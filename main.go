package main

import (
	"context"
	"kedubak/usecase"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var mongoClient *mongo.Client

func init() {
	if os.Getenv("TEST_ENV") == "true" {
		return
	}

	// Chargement du fichier .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Erreur lors du chargement du fichier env:", err)
	}
	log.Println("Fichier env chargé")

	var err error
	mongoClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal("Erreur de connexion à MongoDB:", err)
	}

	if err = mongoClient.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatal("Échec de la connexion à MongoDB:", err)
	}

	log.Println("Connexion à MongoDB réussie")
}

func main() {
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatal("Erreur lors de la déconnexion de MongoDB:", err)
		}
	}()

	coll := mongoClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("COLLECTION_NAME"))

	// Création du service utilisateur
	userService := usecase.UserService{MongoColletion: coll}

	r := mux.NewRouter()

	// Configuration des routes
	r.Use(loggingMiddleware)
	r.HandleFunc("/health", healthHandler).Methods(http.MethodGet)
	r.HandleFunc("/auth/register", userService.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/auth/login", userService.LoginUser).Methods(http.MethodPost)
	r.HandleFunc("/user/me", userService.GetInfoUser).Methods(http.MethodGet) // Correction pour utiliser GET et la route /users
	r.HandleFunc("/user/edit", userService.UpdateUser).Methods(http.MethodPut)
	r.HandleFunc("/user/remove", userService.DeleteUser).Methods(http.MethodDelete)
	r.HandleFunc("/post", userService.CreatePost).Methods(http.MethodPost)

	// Configuration du middleware CORS
	corsHandler := handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
		handlers.AllowedOrigins([]string{"*"}), // À ajuster selon les besoins
		handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions}),
	)

	// Démarrage du serveur
	log.Println("Le serveur est en cours d'exécution sur le port 8080")
	if err := http.ListenAndServe(":8080", corsHandler(r)); err != nil {
		log.Fatal("Erreur lors du démarrage du serveur:", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Requête reçue sur /health")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("running..."))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf(" loggingMiddleWare : Requête reçue: %s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
