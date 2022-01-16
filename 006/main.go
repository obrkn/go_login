package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

// テスト用設定値
var (
	verifyKey = []byte("Drs3amXNE8PnhWxip779Li49auQLx5v5") // 秘密鍵
	store     = sessions.NewCookieStore(verifyKey)
	layout    = "2006-01-02 15:04:05 +0900 JST"
	jst, _    = time.LoadLocation("Asia/Tokyo")
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
	if r.Method == "POST" {
		db, err := sql.Open("mysql", "root@tcp(localhost:3306)/dbname")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		var user User
		email := r.FormValue("email")
		password := r.FormValue("password")
		err = db.QueryRow("SELECT id, email FROM users WHERE email=? AND password=?", email, password).Scan(&user.Id, &user.Email)
		if err == sql.ErrNoRows {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		} else if err != nil {
			log.Fatal(err)
		}

		session, _ := store.Get(r, "Session")
		session.Options.HttpOnly = true

		session.Values["Login"] = true
		session.Values["UserId"] = user.Id

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

		http.Redirect(w, r, "http://localhost:8080/secret", http.StatusMovedPermanently)
	}

	tmpl := `
<!DOCTYPE html>
<body>
	<h1>Login</h1>
	<form method="post" action="/login">
		<label for="email">email:</label><br>
		<input type="email" id="email" name="email" pattern=".+@.+\..+" size="30" required><br><br>
		<label for="password">password:</label><br>
		<input type="password" id="password" name="password" minlength="8" required><br><br>
		<input type="submit" value="Submit">
	</form>
</body>`

	fmt.Fprintln(w, tmpl)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "Session")

	session.Options.MaxAge = -1
	session.Save(r, w)

	fmt.Fprintln(w, "You logged out")
}

type User struct {
	Id        int    `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func signup(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		db, err := sql.Open("mysql", "root@tcp(localhost:3306)/dbname")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		var user User
		email := r.FormValue("email")
		err = db.QueryRow("SELECT id, email FROM users WHERE email=?", email).Scan(&user.Id, &user.Email)
		if err != nil && err != sql.ErrNoRows {
			log.Fatal(err)
		} else if err != sql.ErrNoRows {
			http.Error(w, "Your email is already used", http.StatusBadRequest)
			return
		}
		password := r.FormValue("password")
		hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec("INSERT INTO users(email, password, last_login_at) VALUES(?, ?, NOW());", email, hashed_password)
		if err != nil {
			log.Fatal(err)
		}

		http.Redirect(w, r, "http://localhost:8080/login", http.StatusMovedPermanently)
	}

	tmpl := `
<!DOCTYPE html>
<body>
	<h1>Sign Up</h1>
	<form method="post" action="/signup">
		<label for="email">email:</label><br>
		<input type="email" id="email" name="email" pattern=".+@.+\..+" size="30" required><br><br>
		<label for="password">password:</label><br>
		<input type="password" id="password" name="password" minlength="8" required><br><br>
		<input type="submit" value="Submit">
	</form>
</body>`

	fmt.Fprintln(w, tmpl)
}

func main() {
	http.HandleFunc("/secret", secret)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/signup", signup)

	http.ListenAndServe(":8080", nil)
}
