package main

import (
	"fmt"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

// テスト用設定値
var (
	user_id    = "69"
	secret_key = "Drs3amXNE8PnhWxip779Li49auQLx5v5"
)

type User struct {
	UserId string
	Login  bool
	jwt.StandardClaims
}

func secret(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Session")
	if err != nil {
		log.Fatal(err)
	}
	// Parse the token
	// token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
	// 	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
	// 		return nil, fmt.Errorf(("Invalid Signing Method"))
	// 	}
	// 	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
	// 		return nil, fmt.Errorf(("Expired token"))
	// 	}

	// 	return secret_key, nil
	// })
	// if err != nil {
	// 	fmt.Fprintf(w, err.Error())
	// }
	var user User
	token, err := jwt.ParseWithClaims(cookie.Value, &user, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return []byte(secret_key), nil
	})
	if err != nil {
		fmt.Println("dfasdfadfdfa")
		log.Fatal(err)
	}

	fmt.Println(token)
	// if err != nil || cookie.Value != "true" {
	// 	http.Error(w, "Forbidden", http.StatusForbidden)

	// 	// loginしていない場合の処理

	// 	return
	// }

	// cookie, err = r.Cookie("UserId")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// loginしている場合の処理

	fmt.Fprintln(w, "Your user id is", user.UserId)
}

func login(w http.ResponseWriter, r *http.Request) {
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), &User{
		UserId: user_id,
		Login:  true,
	})
	tokenString, _ := token.SignedString([]byte(secret_key))
	cookie := &http.Cookie{
		Name:  "Session",
		Value: tokenString,
	}
	http.SetCookie(w, cookie)

	fmt.Fprintln(w, "You logged in")
}

func logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Login")
	if err != nil || cookie.Value != "true" {
		http.Error(w, "Forbidden", http.StatusForbidden)
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
