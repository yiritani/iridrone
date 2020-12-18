package main

import (
	models "iridrone/app/models"
	"iridrone/app/controller"

)

func main() {
	go controller.StartWebServer()
	models.NewDroneManager()
	//models.Manual()
}