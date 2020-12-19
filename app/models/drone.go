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
	"log"
	"os/exec"
	"strconv"
	"time"
)

const (
	DefaultSpeed      = 10
	WaitDroneStartSec = 5
	frameX            = 480
	frameY            = 360
	frameArea         = frameX * frameY
	frameSize         = frameArea * 3
	faceDetectXMLFile = "./app/models/default.xml"
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
	isSnapShot           bool
}

func NewDroneManager() *DroneManager {

	drone := tello.NewDriver("8889")
	fmt.Printf("%T", drone)

	//window := gocv.NewWindow("Tello")

	ffmpeg := exec.Command("ffmpeg", "-hwaccel", "auto", "-hwaccel_device", "opencl", "-i", "pipe:0",		"-pix_fmt", "bgr24", "-s", strconv.Itoa(frameX)+"x"+strconv.Itoa(frameY), "-f", "rawvideo", "pipe:1")
	ffmpegIn, _ := ffmpeg.StdinPipe()
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	droneManager := &DroneManager{
		Driver:       drone,
		Speed:        DefaultSpeed,
		isPatrolling: false,
		ffmpegIn:     ffmpegIn,
		ffmpegOut:    ffmpegOut,
		Stream:       mjpeg.NewStream(),
	}

	work := func() {
		if err := ffmpeg.Start(); err != nil {
			fmt.Println(err)
			return
		}
		// ドローン接続
		drone.On(tello.ConnectedEvent, func(data interface{}) {
			fmt.Println("Connected Tello")
			drone.StartVideo()
			drone.SetVideoEncoderRate(tello.VideoBitRateAuto)
			drone.SetExposure(0)

			gobot.Every(100*time.Millisecond, func() {
				drone.StartVideo()
			})

			droneManager.StreamVideo()
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

	go robot.Start(false)
	time.Sleep(WaitDroneStartSec * time.Second)

	//go func() {
	//	for {
	//		Manual(drone)
	//	}
	//}()

	// ffmpegの出力をMac向けに変換
	//for {
	//	buf := make([]byte, frameSize)
	//	if _, err := io.ReadFull(ffmpegOut, buf); err != nil {
	//		fmt.Println(err)
	//		continue
	//	}
	//	img, _ := gocv.NewMatFromBytes(frameY, frameX, gocv.MatTypeCV8UC3, buf)
	//	if img.Empty() {
	//		continue
	//	}
	//
	//	window.IMShow(img)
	//	if window.WaitKey(1) >= 0 {
	//		break
	//	}
	//}
	return droneManager
}

func (d *DroneManager) StreamVideo() {
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

					break
				}
			}

			jpegBuf, _ := gocv.IMEncode(".jpg", img)
			d.Stream.UpdateJPEG(jpegBuf)
		}
	}(d)
}
