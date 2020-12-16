package models

import (
	"fmt"
	"image"
	"os"

	"gocv.io/x/gocv"
)

func MozCamera() {
	if len(os.Args) < 3 {
		fmt.Println("How to run:\n\tfaceblur [camera ID] [classifier XML file]")
		return
	}

	// parse args
	deviceID := os.Args[1]
	xmlFile := os.Args[2]

	// open webcam
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Printf("error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("Face Blur")
	defer window.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load(xmlFile) {
		fmt.Printf("Error reading cascade file: %v\n", xmlFile)
		return
	}

	fmt.Printf("Start reading device: %v\n", deviceID)
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed: %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		// detect faces
		rects := classifier.DetectMultiScale(img)
		fmt.Printf("found %d faces\n", len(rects))

		// blur each face on the original image
		for _, r := range rects {
			imgFace := img.Region(r)

			// blur face
			gocv.GaussianBlur(imgFace, &imgFace, image.Pt(75, 75), 0, 0, gocv.BorderDefault)
			imgFace.Close()
		}

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}