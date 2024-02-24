package magicmirror

import (
	"fmt"
	"net/http"

	"github.com/mperkins808/magic-mirror/go/pkg/messages"
)

// handleMessage takes a base64 encoded message, decodes it, and makes a HTTP request.
func HandleMessage(encoded []byte, local string) (string, error) {

	req, err := messages.DecodeRequest(encoded, local)
	if err != nil {
		return "", err
	}

	// Make the HTTP request using http.Client
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %v", err)
	}

	resp, err := messages.EncodeResponse(response)
	return resp, err

}
