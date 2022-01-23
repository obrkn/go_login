package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

// テスト用設定値
var (
	verifyKey = []byte("Drs3amXNE8PnhWxip779Li49auQLx5v5") // 秘密鍵
	store     = sessions.NewCookieStore(verifyKey)
	layout    = "2006-01-02 15:04:05 +0900 JST"
	dbLayout  = "2006-01-02 15:04:05"
	jst, _    = time.LoadLocation("Asia/Tokyo")
)

type User struct {
	Id             int    `json:"id"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	FailedAttempts int    `json:"failed_attempts"`
	LockedAt       string `json:"locked_at"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

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
		if t, err := time.ParseInLocation(layout, str, jst); err != nil || t.Before(now) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	fmt.Fprintln(w, "Your user id is", session.Values["UserId"])
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		db, _ := sql.Open("mysql", "root@tcp(localhost:3306)/dbname")
		defer db.Close()

		var user User
		err := db.
			QueryRow("SELECT id, email, password, failed_attempts, locked_at FROM users WHERE email=?", r.FormValue("email")).
			Scan(&user.Id, &user.Email, &user.Password, &user.FailedAttempts, &user.LockedAt)

		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		now := time.Now()
		if t, err := time.ParseInLocation(dbLayout, user.LockedAt, jst); err == sql.ErrNoRows || t.Add(30*time.Minute).After(now) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if err := bcrypt.CompareHashAndPassword(
			[]byte(user.Password),
			[]byte(r.FormValue("password"))); err != nil {
			if user.FailedAttempts >= 10 {
				db.Exec("UPDATE users SET failed_attempts=?, locked_at=NOW() WHERE id=?", user.FailedAttempts+1, user.Id)
			} else {
				db.Exec("UPDATE users SET failed_attempts=? WHERE id=?", user.FailedAttempts+1, user.Id)
			}
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		db.Exec("UPDATE users SET failed_attempts=0 WHERE id=?", user.Id)

		session, _ := store.Get(r, "Session")
		session.Options.HttpOnly = true

		session.Values["Login"] = true
		session.Values["UserId"] = user.Id

		if r.FormValue("is_autologin_on") == "on" {
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

	t, _ := template.New("Login").Parse(`
<!DOCTYPE html>
<body>
	<h1>Login</h1>
	<form method="post" action="/login">
		{{.CSRFField}}
		<label for="email">email:</label><br>
		<input type="email" id="email" name="email" pattern=".+@.+\..+" size="30" required><br><br>
		<label for="password">password:</label><br>
		<input type="password" id="password" name="password" minlength="8" required><br><br>
		<label for="is_autologin_on">auto login:</label>
		<input type="checkbox" id="is_autologin_on" name="is_autologin_on"><br><br>
		<input type="submit" value="Submit">
	</form>
</body>`)

	t.Execute(w, map[string]interface{}{
		"CSRFField": csrf.TemplateField(r),
	})
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "Session")

	session.Options.MaxAge = -1
	session.Save(r, w)

	fmt.Fprintln(w, "You logged out")
}

func signup(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		db, _ := sql.Open("mysql", "root@tcp(localhost:3306)/dbname")
		defer db.Close()

		var exists bool
		email := r.FormValue("email")
		db.
			QueryRow("SELECT EXISTS ( SELECT 1 FROM users WHERE email = ? LIMIT 1)", email).
			Scan(&exists)
		if exists {
			http.Error(w, "Your email is already used", http.StatusBadRequest)
			return
		}
		password := r.FormValue("password")
		if len(password) < 8 {
			http.Error(w, "Your password is too short", http.StatusBadRequest)
		}
		hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec("INSERT INTO users(email, password) VALUES(?, ?);", email, hashed_password)
		if err != nil {
			log.Fatal(err)
		}

		http.Redirect(w, r, "http://localhost:8080/login", http.StatusMovedPermanently)
	}

	t, _ := template.New("Signup").Parse(`
<!DOCTYPE html>
<body>
	<h1>Sign Up</h1>
	<form method="post" action="/signup">
		{{.CSRFField}}
		<label for="email">email:</label><br>
		<input type="email" id="email" name="email" pattern=".+@.+\..+" size="30" required><br><br>
		<label for="password">password:</label><br>
		<input type="password" id="password" name="password" minlength="8" required><br><br>
		<input type="submit" value="Submit">
	</form>
</body>`)

	t.Execute(w, map[string]interface{}{
		"CSRFField": csrf.TemplateField(r),
	})
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/secret", secret)
	r.HandleFunc("/login", login)
	r.HandleFunc("/logout", logout)
	r.HandleFunc("/signup", signup)

	h := csrf.Protect(verifyKey, csrf.Secure(false))(r)
	http.ListenAndServe(":8080", h)
}
