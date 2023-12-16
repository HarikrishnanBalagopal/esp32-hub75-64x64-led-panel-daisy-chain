package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	_ "image/png"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	TOTAL_ROWS      = 128
	TOTAL_COLS      = TOTAL_ROWS
	ROWS            = 64
	COLS            = ROWS
	HOST            = "" // "0.0.0.0"
	PORT            = "52275"
	TYPE            = "tcp4"
	FRAME_LEN       = ROWS * COLS * 3
	expected_length = TOTAL_ROWS * TOTAL_COLS * 4
	WEB_STATIC_DIR  = "public"
)

var (
	rgb_frame_panel_0 = make([]byte, FRAME_LEN)
	rgb_frame_panel_1 = make([]byte, FRAME_LEN)
	rgb_frame_panel_2 = make([]byte, FRAME_LEN)
	rgb_frame_panel_3 = make([]byte, FRAME_LEN)
	full_frame        = make([]byte, 4*FRAME_LEN)
	upgrader          = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func fill_panels(frame []byte) {
	// panel 0
	for row := 0; row < ROWS; row++ {
		for col := 0; col < COLS; col++ {
			src := 4 * (row*TOTAL_COLS + col)
			dst := 3 * (row*COLS + col)
			rgb_frame_panel_0[dst+0] = frame[src+0]
			rgb_frame_panel_0[dst+1] = frame[src+1]
			rgb_frame_panel_0[dst+2] = frame[src+2]
		}
	}
	// panel 1
	for row := 0; row < ROWS; row++ {
		for col := 0; col < COLS; col++ {
			src := 4 * (row*TOTAL_COLS + COLS + col)
			dst := 3 * (row*COLS + col)
			rgb_frame_panel_1[dst+0] = frame[src+0]
			rgb_frame_panel_1[dst+1] = frame[src+1]
			rgb_frame_panel_1[dst+2] = frame[src+2]
		}
	}
	// panel 2
	for row := 0; row < ROWS; row++ {
		for col := 0; col < COLS; col++ {
			src := 4 * ((ROWS+row)*TOTAL_COLS + col)
			dst := 3 * ((ROWS-1-row)*COLS + (COLS - 1 - col))
			rgb_frame_panel_2[dst+0] = frame[src+0]
			rgb_frame_panel_2[dst+1] = frame[src+1]
			rgb_frame_panel_2[dst+2] = frame[src+2]
		}
	}
	// panel 3
	for row := 0; row < ROWS; row++ {
		for col := 0; col < COLS; col++ {
			src := 4 * ((ROWS+row)*TOTAL_COLS + COLS + col)
			dst := 3 * ((ROWS-1-row)*COLS + (COLS - 1 - col))
			rgb_frame_panel_3[dst+0] = frame[src+0]
			rgb_frame_panel_3[dst+1] = frame[src+1]
			rgb_frame_panel_3[dst+2] = frame[src+2]
		}
	}
	// copy all panels into the full frame
	if copy(full_frame[0*FRAME_LEN:], rgb_frame_panel_3) != FRAME_LEN {
		panic("copied bytes failed for 1st panel")
	}
	if copy(full_frame[1*FRAME_LEN:], rgb_frame_panel_2) != FRAME_LEN {
		panic("copied bytes failed for 2nd panel")
	}
	if copy(full_frame[2*FRAME_LEN:], rgb_frame_panel_0) != FRAME_LEN {
		panic("copied bytes failed for 3rd panel")
	}
	if copy(full_frame[3*FRAME_LEN:], rgb_frame_panel_1) != FRAME_LEN {
		panic("copied bytes failed for 4th panel")
	}
}

func handleIncomingRequest(conn net.Conn) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	for err == nil {
		current_panel := buffer[0] - '0'
		// logrus.Infof("got data from client: '%s' %x current_panel: %d", buffer, buffer[0], current_panel)
		// respond
		// gif
		if CURR_GIF_FRAME >= 0 && (current_panel == 0 || current_panel == 4) {
			fill_panels(FRAMES[CURR_GIF_FRAME][:])
			CURR_GIF_FRAME = (CURR_GIF_FRAME + 1) % len(FRAMES)
		}
		// gif
		xxxx := rgb_frame_panel_3
		if current_panel == 1 {
			xxxx = rgb_frame_panel_2
		} else if current_panel == 2 {
			xxxx = rgb_frame_panel_0
		} else if current_panel == 3 {
			xxxx = rgb_frame_panel_1
		} else if current_panel == 4 {
			// send all 4 frames
			xxxx = full_frame
		}
		if _, err := conn.Write(xxxx); err != nil {
			logrus.Errorf("failed to write. Error: %q", err)
			break
		}
		// logrus.Infof("sent data, waiting for client request")
		_, err = conn.Read(buffer)
	}
	logrus.Errorf("connection read failed. Error: %q", err)
	conn.Close()
}

// --------------------------------------

const (
	// EXPECTED_W = 498
	EXPECTED_W = 128
	EXPECTED_H = EXPECTED_W
	// EXPECTED_LEN is the length of the palleted image.
	// Paletted is an in-memory image of uint8 indices into a given palette.
	EXPECTED_LEN = EXPECTED_W * EXPECTED_H
	FRAME_W      = 128
	FRAME_H      = FRAME_W
	MY_FRAME_LEN = FRAME_W * FRAME_H * 4
)

type Frame = [MY_FRAME_LEN]byte

var (
	FRAMES         []Frame
	CURR_GIF_FRAME = -1
)

// func load_gif(gif_path string) error {
// 	img_bytes, err := os.ReadFile(gif_path)
// 	if err != nil {
// 		return err
// 	}
// 	img, err := gif.DecodeAll(bytes.NewReader(img_bytes))
// 	if err != nil {
// 		return err
// 	}
// 	// logrus.Infof("img_format %s", img_format)
// 	// if img_format != "gif" {
// 	// 	panic(img_format)
// 	// }
// 	logrus.Infof(
// 		"img %T len(img.Disposal) %+v img.LoopCount %+v None %+v Back %+v Prev %+v",
// 		img,
// 		len(img.Disposal),
// 		img.LoopCount,
// 		gif.DisposalNone,
// 		gif.DisposalBackground,
// 		gif.DisposalPrevious,
// 	)
// 	logrus.Infof("img.Disposal %+v", img.Disposal)
// 	num_frames := len(img.Image)
// 	logrus.Infof("len: %d", num_frames)
// 	frame0 := img.Image[0]
// 	logrus.Infof("frame0 %T", frame0)
// 	logrus.Infof("stride: %d", frame0.Stride)
// 	logrus.Infof("len: %d EXPECTED_LEN %d", len(frame0.Pix), EXPECTED_LEN)
// 	// if len(frame0.Pix) != EXPECTED_LEN {
// 	// 	return fmt.Errorf("expected: %d actual: %d", EXPECTED_LEN, len(frame0.Pix))
// 	// }
// 	FRAMES = make([]Frame, num_frames)
// 	pal := img.Image[0].Palette
// 	var bg_color color.Color = color.RGBA{R: 0, G: 0, B: 0, A: 255}
// 	if int(img.BackgroundIndex) >= len(pal) {
// 		bg_color = pal[img.BackgroundIndex]
// 	}
// 	for n := 0; n < num_frames; n++ {
// 		logrus.Infof("frame[%d]", n)
// 		// prev_frame := img.Image[(n-1+num_frames)%num_frames]
// 		frame := img.Image[n]
// 		prev_idx := (n - 1 + num_frames) % num_frames
// 		for y := 0; y < FRAME_H; y++ {
// 			for x := 0; x < FRAME_W; x++ {
// 				idx := (y*FRAME_W + x) * 4
// 				color := frame.At(x, y)
// 				r, g, b, a := color.RGBA()
// 				// logrus.Infof("a %+v", a)
// 				if a == 0 {
// 					if n == 0 {
// 						r, g, b, _ := bg_color.RGBA()
// 						FRAMES[n][idx+0] = byte(r)
// 						FRAMES[n][idx+1] = byte(g)
// 						FRAMES[n][idx+2] = byte(b)
// 						continue
// 					}
// 					// color := prev_frame.At(x, y)
// 					// r, g, b, _ := color.RGBA()
// 					FRAMES[n][idx+0] = FRAMES[prev_idx][idx+0]
// 					FRAMES[n][idx+1] = FRAMES[prev_idx][idx+1]
// 					FRAMES[n][idx+2] = FRAMES[prev_idx][idx+2]
// 					continue
// 				}
// 				FRAMES[n][idx+0] = byte(r)
// 				FRAMES[n][idx+1] = byte(g)
// 				FRAMES[n][idx+2] = byte(b)
// 			}
// 		}
// 	}
// 	CURR_GIF_FRAME = 0
// 	return nil
// }

func getGifDimensions(gif *gif.GIF) (x, y int) {
	var lowestX int
	var lowestY int
	var highestX int
	var highestY int

	for _, img := range gif.Image {
		if img.Rect.Min.X < lowestX {
			lowestX = img.Rect.Min.X
		}
		if img.Rect.Min.Y < lowestY {
			lowestY = img.Rect.Min.Y
		}
		if img.Rect.Max.X > highestX {
			highestX = img.Rect.Max.X
		}
		if img.Rect.Max.Y > highestY {
			highestY = img.Rect.Max.Y
		}
	}

	return highestX - lowestX, highestY - lowestY
}

// Decode reads and analyzes the given reader as a GIF image
func SplitAnimatedGIF(reader io.Reader, out_dir string) (err error) {
	defer func() {
		if _err := recover(); _err != nil {
			err = fmt.Errorf("error while decoding: %w", _err.(error))
		}
	}()
	gif, err := gif.DecodeAll(reader)
	if err != nil {
		return err
	}
	imgWidth, imgHeight := getGifDimensions(gif)
	overpaintImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(overpaintImage, overpaintImage.Bounds(), gif.Image[0], image.Point{}, draw.Src)
	num_frames := len(gif.Image)
	FRAMES = make([]Frame, num_frames)
	for i, srcImg := range gif.Image {
		draw.Draw(overpaintImage, overpaintImage.Bounds(), srcImg, image.Point{}, draw.Over)
		// file, err := os.Create(fmt.Sprintf("%s/img%d.png", out_dir, i))
		// if err != nil {
		// 	return err
		// }
		// if err := png.Encode(file, overpaintImage); err != nil {
		// 	return err
		// }
		// file.Close()
		for y := 0; y < 128; y++ {
			for x := 0; x < 128; x++ {
				idx := (y*128 + x) * 4
				c := overpaintImage.At(x, y)
				r, g, b, _ := c.RGBA()
				FRAMES[i][idx+0] = byte(r)
				FRAMES[i][idx+1] = byte(g)
				FRAMES[i][idx+2] = byte(b)
			}
		}
	}
	return nil
}

func load_gif(gif_path string) error {
	gif_bytes, err := os.ReadFile(gif_path)
	if err != nil {
		return err
	}
	out_dir := "output"
	if err := os.MkdirAll(out_dir, 0777); err != nil {
		return err
	}
	if err := SplitAnimatedGIF(bytes.NewReader(gif_bytes), out_dir); err != nil {
		return err
	}
	CURR_GIF_FRAME = 0
	return nil
}

// --------------------------------------

func setup() {
	gif_path := "image128.gif"
	logrus.Infof("trying to load the gif '%s'", gif_path)
	if err := load_gif(gif_path); err == nil {
		return
	} else {
		logrus.Errorf("failed to load the gif '%s'. Error: %q", gif_path, err)
	}
	img_path := "image128.png"
	logrus.Infof("trying to load the image '%s'", img_path)
	if img_bytes, err := os.ReadFile(img_path); err == nil {
		img, img_format, err := image.Decode(bytes.NewReader(img_bytes))
		if err != nil {
			panic(err)
		}
		if img_format != "png" {
			panic(img_format)
		}
		img_rgba, ok := img.(*image.RGBA)
		if !ok {
			logrus.Warnf("not imgrgba")
			rect := img.Bounds()
			img_rgba = image.NewRGBA(rect)
			draw.Draw(img_rgba, rect, img, rect.Min, draw.Src)
		}
		frame := img_rgba.Pix
		if len(frame) != expected_length {
			panic(len(frame))
		}
		fill_panels(frame)
		logrus.Infof("setup done with an image")
		return
	} else {
		logrus.Errorf("failed to load the image '%s'. Error: %q", img_path, err)
	}
	// if img_bytes, err := os.ReadFile("image128.bytes"); err == nil {
	// 	const mylen = (128 * 128 * 3)
	// 	if len(img_bytes) != mylen {
	// 		panic("failed size check")
	// 	}
	// 	frame := make([]byte, expected_length)
	// 	for i, j := 0, 0; i < mylen; i += 3 {
	// 		frame[j+0] = img_bytes[i+0]
	// 		frame[j+1] = img_bytes[i+1]
	// 		frame[j+2] = img_bytes[i+2]
	// 		j += 4
	// 	}
	// 	fill_panels(frame)
	// 	return
	// }
	for i := 0; i < FRAME_LEN; i++ {
		rgb_frame_panel_0[i] = 0
		if i%3 == 0 {
			rgb_frame_panel_0[i] = byte(i + 24)
		}
		if i%3 == 1 {
			rgb_frame_panel_0[i] = byte(i)
		}
		if i%3 == 2 {
			rgb_frame_panel_0[i] = 0
		}
	}
}

func main() {
	fmt.Println("start")
	setup()
	go func() {
		listen, err := net.Listen(TYPE, HOST+":"+PORT)
		if err != nil {
			logrus.Fatal(err)
		}
		addr := listen.Addr().String()
		logrus.Infof("tcp frame server listening on address: '%s'", addr)
		// close listener
		defer listen.Close()
		for {
			conn, err := listen.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go handleIncomingRequest(conn)
		}
	}()
	router := mux.NewRouter()
	router.PathPrefix("/ws").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logrus.Errorf("failed to upgrade to a websocket")
			return
		}
		CURR_GIF_FRAME = -1
		conn.WriteMessage(websocket.BinaryMessage, []byte("this is a binary message from the server"))
		for {
			msgtype, frame, err := conn.ReadMessage()
			if err != nil {
				logrus.Errorf("failed to read a message from the websocket. error: %q", err)
				return
			}
			if msgtype != websocket.BinaryMessage {
				logrus.Errorf("expected a binary message on the websocket. actual: '%s'", string(frame))
				continue
			}
			if len(frame) != expected_length {
				logrus.Errorf("expected length '%d'. actual length: '%d'", expected_length, len(frame))
				continue
			}
			// logrus.Infof("got a binary message on the websocket of length: %d", len(frame))
			fill_panels(frame)
		}
	})

	router.PathPrefix("/").Handler(http.FileServer(http.Dir(WEB_STATIC_DIR)))
	srv := &http.Server{
		Handler:      router,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("listening")
	log.Fatal(srv.ListenAndServe())
	fmt.Println("end")
}
