package models

import (
	"fmt"
	"github.com/hybridgroup/mjpeg"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"gocv.io/x/gocv"
	"golang.org/x/sync/semaphore"
	"io"
	"os/exec"
	"strconv"
	"time"
)

const (
	DefaultSpeed      = 10
	WaitDroneStartSec = 5
	frameX            = 960
	frameY            = 720
	frameCenterX      = frameX / 2
	frameCenterY      = frameY / 2
	frameArea         = frameX * frameY
	frameSize         = frameArea * 3
	faceDetectXMLFile = "./app/models/default.xml"
	snapshotsFolder = "./static/image/snapshots/"
)

type DroneManager struct {
	*tello.Driver
	Speed                int
	patrolSem            *semaphore.Weighted
	patrolQuit           chan bool
	isPatrolling         bool
	ffmpegIn             io.WriteCloser
	ffmpegOut            io.ReadCloser
	Stream               *mjpeg.Stream
	faceDetectTrackingOn bool
	isSnapShot bool
}

func NewDroneManager(){

	drone := tello.NewDriver("8889")
	fmt.Printf("%T",drone)

	window := gocv.NewWindow("Tello")

	ffmpeg := exec.Command("ffmpeg", "-hwaccel", "auto", "-hwaccel_device", "opencl", "-i", "pipe:0",
		"-pix_fmt", "bgr24", "-s", strconv.Itoa(frameX)+"x"+strconv.Itoa(frameY), "-f", "rawvideo", "pipe:1")
	ffmpegIn, _ := ffmpeg.StdinPipe()
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	work := func() {
		if err := ffmpeg.Start(); err != nil {
			fmt.Println(err)
			return
		}
		// telloに接続を確認して、Macのビデオを起動
		drone.On(tello.ConnectedEvent, func(data interface{}) {
			fmt.Println("Connected Tello")
			drone.StartVideo()
			drone.SetVideoEncoderRate(tello.VideoBitRateAuto)
			drone.SetExposure(0)

			gobot.Every(100*time.Millisecond, func() {
				drone.StartVideo()
			})

			//droneManager.StreamVideo()
		})

		//FFMPEG functionとvideoデータをつなぐ
		drone.On(tello.VideoFrameEvent, func(data interface{}) {
			pkt := data.([]byte)
			if _, err := ffmpegIn.Write(pkt); err != nil {
				fmt.Println(err)
			}
		})
	}

	robot := gobot.NewRobot("tello",
		[]gobot.Connection{},
		[]gobot.Device{drone},
		work,
	)

	robot.Start(false)

	go func() {
		for {
			Manual(drone)
		}
	}()

	// ffmpegの出力をMac向けに変換
	for {
		buf := make([]byte, frameSize)
		if _, err := io.ReadFull(ffmpegOut, buf); err != nil {
			fmt.Println(err)
			continue
		}
		img, _ := gocv.NewMatFromBytes(frameY, frameX, gocv.MatTypeCV8UC3, buf)
		if img.Empty() {
			continue
		}

		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}

func Manual(drone *tello.Driver){

	var operation string

	fmt.Println("Manual operation Can be entered")
	fmt.Scan(&operation)

	switch operation {
	case "takeoff":
		drone.TakeOff()
	case "land":
		drone.Land()
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
	default:
		fmt.Println("Command ERROR")
	}

}
