package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func main() {
	h1 := func(w http.ResponseWriter, _ *http.Request) {
		const tpl = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="csrf-token" content="b7VZuZLmKFgkNlfXlS6Jw8cPQUreLHgRl66WrY/XiyKTNgL+ewBT8bVqTnPL+1X41td/0YBfeplZkwSK9vtoMQ==" />
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Document</title>
</head>
<body>
<script
  src="https://code.jquery.com/jquery-3.6.0.min.js"
  integrity="sha256-/xUj+3OJU5yExlq6GSYGSHk7tPXikynS7ogEvDej/m4="
  crossorigin="anonymous"></script>
	<form method="post" action="/post">
		<label for="fname">First name:</label><br>
		<input type="text" id="fname" name="fname" value="John"><br>
		<label for="lname">Last name:</label><br>
		<input type="text" id="lname" name="lname" value="Doe"><br><br>
		<input type="submit" value="Submit">
	</form>
</body>
<script>
$.ajaxSetup({
	headers: {
			'X-CSRF-TOKEN': $('meta[name="csrf-token"]').attr('content')
	}
});
</script>
</html>`

		check := func(err error) {
			if err != nil {
				log.Fatal(err)
			}
		}
		t, err := template.New("webpage").Parse(tpl)
		check(err)

		data := struct {
			Title string
			Items []string
		}{
			Title: "My page",
			Items: []string{
				"My photos",
				"My blog",
			},
		}

		err = t.Execute(w, data)
		check(err)

	}

	h2 := func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "success")
	}

	http.HandleFunc("/", h1)
	http.HandleFunc("/post", h2)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
