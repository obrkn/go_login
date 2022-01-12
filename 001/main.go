package main

import (
	"fmt"
	"log"
	"net/http"
)

func secret(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Login")
	if err != nil || cookie.Value != "true" {
		http.Error(w, "Forbidden", http.StatusForbidden)

		// loginしていない場合の処理

		return
	}

	cookie, err = r.Cookie("UserId")
	if err != nil {
		log.Fatal(err)
	}

	// loginしている場合の処理

	fmt.Fprintln(w, "Your user id is", cookie.Value)
}

func login(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:  "Login",
		Value: "true",
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "UserId",
		Value: "69",
	})

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
