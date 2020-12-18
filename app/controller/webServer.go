package controller

import (
	"fmt"
	"html/template"
	"net/http"
)


func getTemplate(temp string) (*template.Template, error) {
	return template.ParseFiles("app/views/layout.html", temp)
}

func viewIndexHandler(w http.ResponseWriter, r *http.Request) {
	operation := r.FormValue("operation")
	fmt.Println(operation)
	t, _ := getTemplate("app/views/index.html")
	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}



func StartWebServer() {
	http.HandleFunc("/", viewIndexHandler)

	http.ListenAndServe(":8080", nil)
}
