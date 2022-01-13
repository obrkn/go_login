package main

import (
	"fmt"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

// テスト用設定値
var (
	testUserId = "69"
	verifyKey  = []byte("Drs3amXNE8PnhWxip779Li49auQLx5v5")
)

type User struct {
	UserId string
	Login  bool
	jwt.StandardClaims
}

func secret(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Session")
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)

		// loginしていない場合の処理

		return
	}

	var user User
	token, err := jwt.ParseWithClaims(cookie.Value, &user, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})
	if err != nil {
		log.Fatal(err)
	}

	if !token.Valid || !user.Login {
		http.Error(w, "Forbidden", http.StatusForbidden)

		// loginしていない場合の処理

		return
	}

	// loginしている場合の処理

	fmt.Fprintln(w, "Your user id is", user.UserId)
}

func login(w http.ResponseWriter, r *http.Request) {
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), &User{
		UserId: testUserId,
		Login:  true,
	})
	tokenString, _ := token.SignedString(verifyKey)
	cookie := &http.Cookie{
		Name:     "Session",
		Value:    tokenString,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	fmt.Fprintln(w, "You logged in")
}

func logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Session")
	if err != nil {
		log.Fatal(err)
	}
	var user User
	token, err := jwt.ParseWithClaims(cookie.Value, &user, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})
	if err != nil {
		log.Fatal(err)
	}

	if !token.Valid || !user.Login {
		http.Error(w, "Forbidden", http.StatusForbidden)

		// loginしていない場合の処理

		return
	}
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)

	fmt.Fprintln(w, "You logged out")
}

func main() {
	http.HandleFunc("/secret", secret)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)

	http.ListenAndServe(":8080", nil)
}
