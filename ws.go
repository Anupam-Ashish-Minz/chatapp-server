package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

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

		row := db.QueryRow(`select messages.id, time, body, user_id, users.id,
		users.name, users.email from messages inner join users on user_id =
		users.id where messages.id = ?`, id)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var messageDB Message
		row.Scan(&messageDB.ID, &messageDB.Time, &messageDB.Body, &messageDB.UserID,
			&messageDB.User.ID, &messageDB.User.Name, &messageDB.User.Email)

		wsjson.Write(ctx, c, messageDB)
	}
}
