package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"github.com/rs/cors"
)

const (
	DB_NAME     = "test.db"
	AUTH_COOKIE = "auth-cookie"
)

var JWT_SECRET = []byte("aseotuasoetu")

type User struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

type Message struct {
	ID         int    `json:"id,omitempty"`
	Body       string `json:"body"`
	Time       string `json:"time"`
	UserID     int    `json:"user_id"`
	ChatroomID int    `json:"chatroom_id"`
	User       User   `json:"user,omitempty"`
}

type Chatroom struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MessageWS struct {
	Body       string `json:"body"`
	ChatroomID int    `json:"chatroom_id"`
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("hi"))
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("the api is working"))
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := validateAuth(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	db, err := sql.Open("sqlite3", DB_NAME)
	defer db.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var user User
	row := db.QueryRow(`select name, email from user where id = ?`, userID)
	user.ID = userID
	row.Scan(&user.Name, &user.Email)

	json.NewEncoder(w).Encode(user)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	router := http.NewServeMux()

	// get request
	router.HandleFunc("/api/messages/", messageListHandler)
	router.HandleFunc("/api/chatrooms", chatroomsHandler)
	router.HandleFunc("/api/profile", profileHandler)
	// post request
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/signup", signupHandler)
	router.HandleFunc("/logout", logoutHandler)
	router.HandleFunc("/api/message", messageHandler)
	router.HandleFunc("/ws/message", wsMessageHandler)

	// just for testing purposes
	router.HandleFunc("/api", apiHandler)

	// 404
	router.HandleFunc("/", rootHandler)

	handler := cors.New(cors.Options{
		AllowedOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:     []string{"GET", "POST"},
		ExposedHeaders:     []string{"Content-Type", "Authorization"},
		AllowedHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials:   true,
		OptionsPassthrough: false,
		Debug:              false,
	}).Handler(router)

	http.ListenAndServe(":4000", handler)
}
