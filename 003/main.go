package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

// テスト用設定値
var (
	testUserId = "f63644c8-1e80-7975-6637-681c173971c2"     // ユーザーID
	verifyKey  = []byte("Drs3amXNE8PnhWxip779Li49auQLx5v5") // 秘密鍵
	store      = sessions.NewCookieStore(verifyKey)
)

func secret(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "Session")

	if auth, ok := session.Values["Login"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	fmt.Fprintln(w, "Your user id is", session.Values["UserId"])
}

func login(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "Session")

	session.Options = &sessions.Options{
		HttpOnly: true,
	}

	session.Values["Login"] = true
	session.Values["UserId"] = testUserId
	session.Save(r, w)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "Session")

	session.Options.MaxAge = -1
	session.Save(r, w)
}

func main() {
	http.HandleFunc("/secret", secret)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)

	http.ListenAndServe(":8080", nil)
}
