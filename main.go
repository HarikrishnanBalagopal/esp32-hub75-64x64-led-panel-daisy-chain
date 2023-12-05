package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
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
}

func handleIncomingRequest(conn net.Conn) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	for err == nil {
		current_panel := buffer[0] - '0'
		// logrus.Infof("got data from client: '%s' %x current_panel: %d", buffer, buffer[0], current_panel)
		// respond
		xxxx := rgb_frame_panel_3
		if current_panel == 1 {
			xxxx = rgb_frame_panel_2
		} else if current_panel == 2 {
			xxxx = rgb_frame_panel_0
		} else if current_panel == 3 {
			xxxx = rgb_frame_panel_1
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

func setup() {
	if img_bytes, err := os.ReadFile("image128.png"); err == nil {
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
