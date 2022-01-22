package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/signup", ShowSignupForm)
	r.HandleFunc("/signup/post", SubmitSignupForm).Methods("POST")

	// Protect ミドルウェアは、非冪等なHTTPメソッドの場合にサブミットされたフォーム
	// をチェックし、CSRFトークンが一致しているか検証する。
	//
	// auth-key はコードに書き込まず、/dev/urandom したものを外部から与えること。
	// キーが変わるとそれまでに発行したトークンの検証に失敗する。
	//
	// HTTP な開発環境では opt に csrf.Secure(false) 指定が必要。本番では外すこと。
	h := csrf.Protect([]byte("32-byte-long-auth-key"), csrf.Secure(false))(r)
	http.ListenAndServe(":8080", h)
}

func ShowSignupForm(w http.ResponseWriter, r *http.Request) {

	// {{.CSRFField}} にCSRFトークンの隠しinputを埋め込む。
	t, _ := template.New("form").Parse(`
<html>
  <body>
      <form method="POST" action="/signup/post" accept-charset="UTF-8">
          <input type="text" name="name">
          <input type="text" name="email">
          {{.CSRFField}}
          <input type="submit" value="Sign up!">
      </form>
  </body>
</html> `)
	t.Execute(w, map[string]interface{}{
		"CSRFField": csrf.TemplateField(r),
	})
}

func SubmitSignupForm(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	fmt.Fprintf(w, "%v\n", r.PostForm)
	// TODO: ユーザー登録
}
