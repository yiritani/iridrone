package main

import (
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	_ "net/http/pprof"
	"time"
)


func main() {
	doron := tello.NewDriver("8889")
	work := func(){
		doron.TakeOff()

		gobot.After(5 * time.Second, func() {
			doron.Forward(10)
		})
		gobot.After(5 * time.Second, func() {
			doron.Backward(10)
		})
		gobot.After(5 * time.Second, func() {
			doron.Right(10)
		})
		gobot.After(5 * time.Second, func() {
			doron.Left(10)
		})
		gobot.After(5 * time.Second, func() {
			doron.BackFlip()
		})
		gobot.After(10 * time.Second, func() {
			doron.FrontFlip()
		})
		gobot.After(8 * time.Second, func() {
			doron.Land()
		})

	}

	robot := gobot.NewRobot("tello",
		[]gobot.Connection{},
		[]gobot.Device{doron},
		work)
	robot.Start()

}