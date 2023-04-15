package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func messageListHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := validateAuth(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	chatroomID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/messages/"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", DB_NAME)
	defer db.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	row := db.QueryRow(`select user_id from messages where user_id = ? and
		chatroom_id = ?`, userID, chatroomID)
	// what is this ? why !?
	row.Scan(&userID)

	if userID == 0 {
		return
	}

	rows, err := db.Query(`select messages.id, time, body, user_id, users.id,
		users.name, users.email from messages inner join users on user_id =
		users.id where chatroom_id = ?`, chatroomID)
	defer rows.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var messages []Message
	for rows.Next() {
		var message Message
		rows.Scan(&message.ID, &message.Time, &message.Body, &message.UserID,
			&message.User.ID, &message.User.Name, &message.User.Email)
		message.ChatroomID = chatroomID

		messages = append(messages, message)
	}

	json.NewEncoder(w).Encode(messages)
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(AUTH_COOKIE)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	tokenString := cookie.Value
	userID, err := parseToken(tokenString)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var msg Message
	json.NewDecoder(r.Body).Decode(&msg)

	msg.UserID = userID

	if msg.Body == "" || msg.ChatroomID == 0 {
		log.Println("empty message body or missing chatroom id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO check if user belongs to this chatroom he says he belongs to

	db, err := sql.Open("sqlite3", DB_NAME)
	defer db.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(`insert into messages (time, body, user_id, chatroom_id)`,
		time.Now().Format(time.RFC3339),
		msg.Body,
		userID,
		msg.ChatroomID,
	)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func chatroomsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", DB_NAME)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// TODO query chatrooms which concers the currently logged in user
	rows, err := db.Query(`select id, name from chatrooms`)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var chatrooms []Chatroom

	for rows.Next() {
		var chatroom Chatroom
		rows.Scan(&chatroom.ID, &chatroom.Name)
		chatrooms = append(chatrooms, chatroom)
	}

	json.NewEncoder(w).Encode(chatrooms)
}
