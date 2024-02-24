package magicmirror

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// estalishes a websocket connection to remote host and forwards message requests
func Mirror(remote, local string, name string, apikey string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {

		for {
			listen(interrupt, remote, local, name, apikey)

			select {
			case <-interrupt:
				log.Println("Interrupt signal received, stopping...")
				return
			case <-time.After(3 * time.Second): // Wait before retrying to connect
			}
		}
	}()

	<-interrupt
	log.Println("Mirror exiting.")
}

func listen(interrupt chan os.Signal, remote string, local string, name string, apikey string) bool {

	uRemote, err := url.Parse(remote)
	if err != nil {
		log.Fatalf("Failed to parse remote URL: %v", err)
	}

	dialer := websocket.Dialer{HandshakeTimeout: 45 * time.Second}
	header := make(http.Header)
	if apikey != "" {
		header.Set("Authorization", fmt.Sprintf("Bearer %s", apikey))
	}

	connectionURL := fmt.Sprintf("%s?name=%s", uRemote.String(), name)

	conn, res, err := dialer.Dial(connectionURL, header)
	if res != nil && res.StatusCode != http.StatusSwitchingProtocols {
		msg, err := io.ReadAll(res.Body)
		if err != nil {
			msg = []byte("")
		}
		log.Fatalf("failed to connect to remote host: %v. response code %v with response \n%v", remote, res.Status, string(msg))
	}
	if err != nil {
		log.Errorf("Failed to connect to remote host: %v %v. Retrying...", remote, err)
		return false
	}

	defer conn.Close()
	go handleInterrupt(interrupt, conn)

	log.Infof("Connected to %v", uRemote.String())
	failedRead := 0
	for {
		if failedRead >= MAX_CONSECUTIVE_FAILED_MESSAGE_READS {
			log.Fatalf("failed to read message %v times in a row. exiting", MAX_CONSECUTIVE_FAILED_MESSAGE_READS)
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Errorf("Error reading message: %v. Attempting to reconnect...", err)
			return false
		}

		resp, err := HandleMessage(message, local)
		if err != nil {
			log.Errorf("Error handling message: %v", err)
			failedRead++
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte(resp))
		if err != nil {
			log.Errorf("Error writing message: %v", err)
		}
	}
}
