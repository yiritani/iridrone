package controller

import (
	"html/template"
	"iridoron/app/models"
	"net/http"
)


func getTemplate(temp string) (*template.Template, error) {
	return template.ParseFiles("app/views/layout.html", temp)
}

func viewIndexHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := getTemplate("app/views/index.html")
	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}



func StartWebServer() {
	drone := models.DroneManager{}
	http.HandleFunc("/", viewIndexHandler)
	//http.ListenAndServe(":8080", nil)
	http.Handle("/video/streaming", drone.Stream)

}
