package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

func createToken(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
	})

	tokenString, err := token.SignedString(JWT_SECRET)

	return tokenString, err
}

func parseToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return JWT_SECRET, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["user_id"].(int); ok {
			return userID, nil
		}
		return 0, fmt.Errorf("claim user_id is incorrect")
	}
	return 0, fmt.Errorf("unknown error")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	} else if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("use post request"))
		return
	}
	var reqUser User
	json.NewDecoder(r.Body).Decode(&reqUser)
	if reqUser.Email == "" || reqUser.Password == "" {
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

	row := db.QueryRow(`select id, email, password from users where email = ?`, reqUser.Email)
	var dbUser User
	row.Scan(&dbUser.ID, &dbUser.Email, &dbUser.Password)

	if reqUser.Email != dbUser.Email || reqUser.Password != dbUser.Password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid email or password"))
		log.Println("auth request failed invalid email or password from user ", reqUser.Email)
		return
	}

	tokenString, err := createToken(dbUser.ID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     AUTH_COOKIE,
		Value:    tokenString,
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   3600 * 24 * 30,
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	var reqUser User
	json.NewDecoder(r.Body).Decode(&reqUser)
	if reqUser.Name == "" || reqUser.Email == "" || reqUser.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("empty user")
		return
	}

	db, err := sql.Open("sqlite3", DB_NAME)
	defer db.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := db.Exec(`insert into users (name, email, password) values
		(?, ?, ?)`, reqUser.Name, reqUser.Email, reqUser.Password)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userID, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenString, err := createToken(int(userID))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{
		Name:     AUTH_COOKIE,
		Value:    tokenString,
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   3600 * 24 * 30,
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
	})
}
