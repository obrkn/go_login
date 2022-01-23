// sessions.go
package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

var (
	store = sessions.NewCookieStore([]byte("32-byte-long-auth-key"))
)

func main() {
	http.HandleFunc("/login", login)
	http.HandleFunc("/form", form)
	http.HandleFunc("/form/post", submitForm)

	http.ListenAndServe(":8080", nil)
}

// ログイン処理
func login(w http.ResponseWriter, r *http.Request) {
	// Cookieをセット
	session, _ := store.Get(r, "cookie-name")
	session.Values["authenticated"] = true
	session.Save(r, w)

	fmt.Println("You logged in")
}

// 更新フォーム
func form(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `
<html>
<body>
	<form method="POST" action="/form/post" accept-charset="UTF-8">
		<input type="text" name="content">
		<input type="submit" value="Submit">
	</form>
</body>
</html>`)
}

// 更新処理
func submitForm(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		session, _ := store.Get(r, "cookie-name")

		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		_ = r.ParseForm()
		fmt.Fprintf(w, "%v\n", r.PostForm)
	}
}
