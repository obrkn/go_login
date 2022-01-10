package main

import (
	"fmt"
	"net/http"
)

func secret(w http.ResponseWriter, r *http.Request) {
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Print secret message
	fmt.Fprintln(w, "The cake is a lie!")
}

func login(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:  "user_id", // ここにcookieの名前を記述
		Value: "bar",     // ここにcookieの値を記述
	}
	http.SetCookie(w, cookie)

	fmt.Println(cookie)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Revoke users authentication
	// session.Values["authenticated"] = false
	// store.MaxAge(-1)
	session.Options.MaxAge = -1
	session.Save(r, w)
}

func main() {
	http.HandleFunc("/secret", secret)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)

	http.ListenAndServe(":8080", nil)
}
