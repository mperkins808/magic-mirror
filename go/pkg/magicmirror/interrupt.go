package magicmirror

import (
	"os"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

func handleInterrupt(interrupt chan os.Signal, conn *websocket.Conn) {
	for {
		select {
		case <-interrupt:
			log.Println("Interrupt signal received, closing connection")

			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Errorf("Error sending close message: %v", err)
				return
			}
			select {
			case <-time.After(time.Second):
			case <-interrupt:
				conn.Close()
				return
			}
			conn.Close()
			return
		}
	}
}
