package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

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

func wsMessageHandler(w http.ResponseWriter, r *http.Request) {

	coo, err := r.Cookie(AUTH_COOKIE)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
	}

	tokenString := coo.Value
	userID, err := parseToken(tokenString)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	defer c.Close(websocket.StatusGoingAway, "websocket closed")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	// using timeout might be a good idea but in the client add a method
	// such that it will reconnect the client is the user gain activity back
	// though not having notification because of a non active websocket
	// connection seems annoying so indefinited websocket connection does not
	// seem that bad after all
	//
	// ctx, cancel := context.WithTimeout(r.Context(), time.Hour*10)
	// defer cancel()

	ctx := r.Context()

	db, err := sql.Open("sqlite3", DB_NAME)
	defer db.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for {
		var message MessageWS

		if err = wsjson.Read(ctx, c, &message); err != nil {
			log.Println(err)
			break
		}

		res, err := db.Exec(`insert into messages (time, body, user_id,
			chatroom_id) values (?, ?, ?, ?)`, time.Now().Format(time.RFC3339),
			message.Body, userID, message.ChatroomID)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		id, err := res.LastInsertId()
		if err != nil {
			log.Println(err)
			break
		}

		err = wsjson.Write(ctx, c, map[string]interface{}{"id": id})
		if err != nil {
			log.Println(err)
			break
		}
	}
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
