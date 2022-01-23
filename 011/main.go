package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// 設定値
var (
	layout = "2006-01-02 15:04:05"
	jst, _ = time.LoadLocation("Asia/Tokyo")
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

func main() {
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)

	http.ListenAndServe(":8080", nil)
}

// サインアップ
func signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		db, _ := sql.Open("mysql", "root@tcp(localhost:3306)/dbname")
		defer db.Close()

		email := r.FormValue("email")
		password := r.FormValue("password")

		// パスワードが8文字未満か？
		if len(password) < 8 {
			http.Error(w, "Too short password", http.StatusBadRequest)
			return
		}

		// 既に存在するemailか？
		var exists bool
		db.
			QueryRow("SELECT EXISTS ( SELECT 1 FROM users WHERE email = ? LIMIT 1)", email).
			Scan(&exists)
		if exists {
			http.Error(w, "Already used email", http.StatusBadRequest)
			return
		}

		// パスワードを暗号化
		hashed_password, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		// データベースへ登録
		db.Exec("INSERT INTO users(email, password) VALUES(?, ?);", email, hashed_password)

		fmt.Fprintln(w, "Registration completed")
		return
	}

	// サインアップフォーム
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

// ログイン
func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		db, _ := sql.Open("mysql", "root@tcp(localhost:3306)/dbname")
		defer db.Close()

		// emailが一致するユーザーを検索
		var user User
		err := db.
			QueryRow("SELECT id, email, password, failed_attempts, locked_at FROM users WHERE email=?", r.FormValue("email")).
			Scan(&user.Id, &user.Email, &user.Password, &user.FailedAttempts, &user.LockedAt)
		if err != nil {
			http.Error(w, "Wrong email", http.StatusForbidden)
			return
		}

		// ロックから30分以内のログインは許可しない
		now := time.Now()
		if t, err := time.ParseInLocation(layout, user.LockedAt, jst); err == sql.ErrNoRows || t.Add(30*time.Minute).After(now) {
			http.Error(w, "Locked account", http.StatusForbidden)
			return
		}

		// パスワードが正しいか？
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(r.FormValue("password"))); err != nil {
			if user.FailedAttempts >= 10 {
				// ログイン失敗回数が10回以上の場合はロックをかける
				db.Exec("UPDATE users SET failed_attempts=?, locked_at=NOW() WHERE id=?", user.FailedAttempts+1, user.Id)
			} else {
				// ログイン失敗回数が10回未満の場合は失敗回数を増やす
				db.Exec("UPDATE users SET failed_attempts=? WHERE id=?", user.FailedAttempts+1, user.Id)
			}
			http.Error(w, "Wrong password", http.StatusForbidden)
			return
		}

		// ログインに成功した場合はログイン失敗回数をリセット
		db.Exec("UPDATE users SET failed_attempts=0 WHERE id=?", user.Id)

		fmt.Fprintln(w, "You are logged in as", user.Id)
		return
	}

	// ログインフォーム
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
