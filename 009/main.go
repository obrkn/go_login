package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// usersテーブル
type User struct {
	Id        int    `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func main() {
	http.HandleFunc("/login", login)
	http.HandleFunc("/signup", signup)

	http.ListenAndServe(":8080", nil)
}

// ログイン
func login(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		// DB接続
		db, _ := sql.Open("mysql", "root@tcp(localhost:3306)/dbname")
		defer db.Close()

		var user User
		email := r.FormValue("email")
		password := r.FormValue("password")
		err := db.QueryRow("SELECT id, email FROM users WHERE email=? AND password=?", email, password).Scan(&user.Id, &user.Email)
		if err != nil { // 一致するデータがない場合はerrが返る
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		fmt.Fprintln(w, "You loged in")
		return
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

// サインアップ
func signup(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		db, _ := sql.Open("mysql", "root@tcp(localhost:3306)/dbname")
		defer db.Close()

		// 既に使われているメールアドレスか判別
		var exists bool
		email := r.FormValue("email")
		db.
			QueryRow("SELECT EXISTS ( SELECT 1 FROM users WHERE email = ? LIMIT 1)", email).
			Scan(&exists)
		if exists {
			http.Error(w, "Your email is already used", http.StatusBadRequest)
			return
		}

		// DBへ登録
		password := r.FormValue("password")
		db.Exec("INSERT INTO users(email, password) VALUES(?, ?);", email, password)

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
