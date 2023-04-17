package main

import (
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
	Password string `json:"password"`
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

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	router := http.NewServeMux()

	// get request
	router.HandleFunc("/api/messages/", messageListHandler)
	router.HandleFunc("/api/chatrooms", chatroomsHandler)
	// post request
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/signup", signupHandler)
	router.HandleFunc("/api/message", messageHandler)
	router.HandleFunc("/ws/message", wsMessageHandler)

	// just for testing purposes
	router.HandleFunc("/api", apiHandler)

	// 404
	router.HandleFunc("/", rootHandler)

	handler := cors.New(cors.Options{
		AllowedOrigins:     []string{"http://localhost:5173"},
		AllowedMethods:     []string{"GET", "POST"},
		ExposedHeaders:     []string{"Content-Type", "Authorization"},
		AllowedHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials:   true,
		OptionsPassthrough: false,
		Debug:              false,
	}).Handler(router)

	http.ListenAndServe(":4000", handler)
}
