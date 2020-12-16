package models

import (
	"fmt"
	"github.com/hybridgroup/mjpeg"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"gocv.io/x/gocv"
	"golang.org/x/sync/semaphore"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strconv"
	"time"
)

const (
	DefaultSpeed      = 10
	WaitDroneStartSec = 5
	frameX            = 960 / 3
	frameY            = 720 / 3
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

func NewDroneManager() *DroneManager{
	drone := tello.NewDriver("8889")
	ffmpeg := exec.Command("ffmpeg", "-hwaccel", "auto", "-hwaccel_device", "opencl", "-i", "pipe:0", "-pix_fmt", "bgr24",
		"-s", strconv.Itoa(960)+"x"+strconv.Itoa(720), "-f", "rawvideo", "pipe:1")
	ffmpegIn, _ := ffmpeg.StdinPipe()
	ffmpegOut, _ := ffmpeg.StdoutPipe()



	droneManager := &DroneManager{
		Driver:               drone,
		Speed:                DefaultSpeed,
		ffmpegIn:             ffmpegIn,
		ffmpegOut:            ffmpegOut,
		Stream:               mjpeg.NewStream(),
		faceDetectTrackingOn: false,
	}
	work := func(){
		drone.On(tello.ConnectedEvent, func(data interface{}) {
			drone.StartVideo()
			drone.SetVideoEncoderRate(tello.VideoBitRateAuto)
			drone.SetExposure(0)

			gobot.Every(100*time.Millisecond, func() {
				drone.StartVideo()
			})

			//droneManager.StreamVideo()
		})
		drone.On(tello.VideoFrameEvent, func(data interface{}){
			pkt := data.([]byte)
			if _, err := ffmpegIn.Write(pkt); err != nil {
				log.Println(err)
			}
		})
	}
	robot := gobot.NewRobot("tello", []gobot.Connection{}, []gobot.Device{drone}, work)
	go robot.Start()
	time.Sleep(5 * time.Second)
	return droneManager
}

func (d *DroneManager)StreamVideo() {
	go func(d *DroneManager) {
		classifier := gocv.NewCascadeClassifier()
		defer classifier.Close()
		if !classifier.Load(faceDetectXMLFile) {
			log.Printf("Error reading cascade file: %v\n", faceDetectXMLFile)
			return
		}
		blue := color.RGBA{0, 0, 255, 0}

		for {
			buf := make([]byte, frameSize)
			if _, err := io.ReadFull(d.ffmpegOut, buf); err != nil {
				log.Println(err)
			}
			img, _ := gocv.NewMatFromBytes(frameY, frameX, gocv.MatTypeCV8UC3, buf)

			if img.Empty() {
				continue
			}

			if d.faceDetectTrackingOn {
				rects := classifier.DetectMultiScale(img)
				log.Printf("found %d faces\n", len(rects))

				if len(rects) == 0 {
					d.Hover()
				}
				for _, r := range rects {
					gocv.Rectangle(&img, r, blue, 3)
					pt := image.Pt(r.Max.X, r.Min.Y-5)
					gocv.PutText(&img, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)

				}
			}

			jpegBuf, _ := gocv.IMEncode(".jpg", img)
			if d.isSnapShot {
				backupFileName := snapshotsFolder + time.Now().Format(time.RFC3339) + ".jpg"
				ioutil.WriteFile(backupFileName, jpegBuf, 0644)
				snapshotFileName := snapshotsFolder + "snapshot.jpg"
				ioutil.WriteFile(snapshotFileName, jpegBuf, 0644)
				d.isSnapShot = false
			}
			d.Stream.UpdateJPEG(jpegBuf)
		}
	}(d)
}

func Manual(){
	var operation string

	conn, _ := net.Dial("udp","192.168.10.1:8889")
	conn.Write([]byte("command"))
		for {
			fmt.Scan(&operation)
			conn.Write([]byte(operation))
		}
}