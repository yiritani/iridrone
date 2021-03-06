package controller

import (
	"fmt"
	"html/template"
	"iridrone/app/models"
	"net/http"
)

var appContext struct {
	DroneManager *models.DroneManager
}

func init() {
	appContext.DroneManager = models.NewDroneManager()
}

func getTemplate(temp string) (*template.Template, error) {
	return template.ParseFiles("app/views/layout.html", temp)
}

func viewIndexHandler(w http.ResponseWriter, r *http.Request) {
	drone := appContext.DroneManager

	t, _ := getTemplate("app/views/index.html")
	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	operation := r.FormValue("operation")
	fmt.Println(operation)
	Manual(operation, drone)
}

func Manual(operation string, drone *models.DroneManager) {

	switch operation {
	case "takeoff":
		drone.TakeOff()
	case "rflip":
		drone.RightFlip()
	case "lflip":
		drone.LeftFlip()
	case "fflip":
		drone.FrontFlip()
	case "bflip":
		drone.BackFlip()
	case "throw":
		drone.ThrowTakeOff()
	case "bounce":
		drone.Bounce()
	case "land":
		drone.Land()
	}
}

func StartWebServer() {
	http.HandleFunc("/", viewIndexHandler)
	http.Handle("/video/streaming", appContext.DroneManager.Stream)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8080", nil)
}