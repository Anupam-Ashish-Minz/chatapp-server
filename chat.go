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
	roomID_s := strings.TrimPrefix(r.URL.Path, "/api/messages/")
	if roomID_s == "" {
		log.Println("room id is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	roomID, err := strconv.Atoi(roomID_s)
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

	rows, err := db.Query(`select id, time, body, user_id from messages where
		chatroom_id = ?`, roomID)
	defer rows.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var messages []Message
	for rows.Next() {
		var message Message
		rows.Scan(&message.ID, &message.Body, &message.UserID)
		message.ChatroomID = roomID

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
