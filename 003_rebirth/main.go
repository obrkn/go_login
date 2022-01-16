package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

// テスト用設定値
var (
	testUserId = 69                                         // ユーザーID
	verifyKey  = []byte("Drs3amXNE8PnhWxip779Li49auQLx5v5") // 秘密鍵
	store      = sessions.NewCookieStore(verifyKey)
	layout     = "2006-01-02 15:04:05 +0900 JST"
	jst, _     = time.LoadLocation("Asia/Tokyo")
)

func secret(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "Session")

	if auth, ok := session.Values["Login"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if str, ok := session.Values["ExpiredAt"].(string); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	} else {
		// 期限を過ぎた場合はエラーを返す
		now := time.Now()
		if time, err := time.ParseInLocation(layout, str, jst); err != nil || time.Before(now) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	fmt.Fprintln(w, "Your user id is", session.Values["UserId"])
}

func login(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "Session")
	session.Options.HttpOnly = true

	session.Values["Login"] = true
	session.Values["UserId"] = testUserId

	if r.FormValue("is_autologin_on") == "true" {
		// 自動ログインをオンにした場合
		session.Values["ExpiredAt"] = time.Now().AddDate(0, 0, 7).Format(layout) // 1週間後に設定
		session.Options.MaxAge = 60 * 60 * 24 * 7                                // 1週間後に設定
	} else {
		// 自動ログインをオフにした場合
		session.Values["ExpiredAt"] = time.Now().Add(time.Hour).Format(layout) // 1時間後に設定
		session.Options.MaxAge = 0                                             // 0に設定すると、ブラウザを閉じたときに無効になる
	}
	session.Save(r, w)

	fmt.Fprintln(w, "You logged in")
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "Session")

	session.Options.MaxAge = -1
	session.Save(r, w)

	fmt.Fprintln(w, "You logged out")
}

func main() {
	http.HandleFunc("/secret", secret)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)

	http.ListenAndServe(":8080", nil)
}
