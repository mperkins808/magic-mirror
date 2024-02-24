package encoder

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type EncodedRequest struct {
	Method  string              `json:"method"`
	Uri     string              `json:"uri"`
	Body    []byte              `json:"body"`
	Headers map[string][]string `json:"headers"`
}

type EncodedResponse struct {
	Body       []byte              `json:"body"`
	Headers    map[string][]string `json:"headers"`
	StatusCode int                 `json:"status_code"`
}

// base64 encodes a request
func EncodeRequest(req *http.Request) (string, error) {
	if req == nil {
		return "", fmt.Errorf("request is nil")
	}

	uri := fmt.Sprintf("%v://%v%v", req.URL.Scheme, req.URL.Host, req.URL.Path)
	method := req.Method

	var body []byte = nil

	if req.Body != nil {
		bod, err := io.ReadAll(req.Body)
		if err != nil {
			return "", fmt.Errorf("failed to decode request body %v", err)
		}
		body = bod
	}

	headers := make(http.Header)

	// Populate the headers
	for k, v := range req.Header {
		for _, hv := range v {
			headers.Add(k, hv)
		}
	}

	encoded := EncodedRequest{
		Method:  method,
		Uri:     uri,
		Body:    body,
		Headers: headers,
	}

	b, err := json.Marshal(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to encode request %v", err)
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

// base64 decodes a request
func DecodeRequest(req []byte) (*http.Request, error) {
	var decoded EncodedRequest
	decodedBytes, err := base64.StdEncoding.DecodeString(string(req))
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode request: %v", err)
	}

	err = json.Unmarshal(decodedBytes, &decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %v", err)
	}

	request, err := http.NewRequest(decoded.Method, decoded.Uri, strings.NewReader(string(decoded.Body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set the headers for the request
	for k, v := range decoded.Headers {
		for _, hv := range v {
			request.Header.Add(k, hv)
		}
	}

	return request, nil
}

// base64 encodes a response
func EncodeResponse(res *http.Response) (string, error) {
	if res == nil {
		return "", fmt.Errorf("response is nil")
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		b = nil
	}

	encoded := EncodedResponse{
		Body:       b,
		Headers:    res.Header,
		StatusCode: res.StatusCode,
	}

	bytes, err := json.Marshal(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to json encode response %v", err)
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// base64 decodes a response
func DecodeResponse(encodedResp string) (*http.Response, error) {
	var decoded EncodedResponse
	b, err := base64.StdEncoding.DecodeString(encodedResp)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode response: %v", err)
	}

	// Unmarshal the JSON into the EncodedResponse struct
	err = json.Unmarshal(b, &decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// Construct the response body from the decoded body
	responseBody := io.NopCloser(bytes.NewReader(decoded.Body))

	// Construct the http.Response
	response := &http.Response{
		StatusCode: decoded.StatusCode,
		Header:     make(http.Header),
		Body:       responseBody,
	}

	// Populate the headers
	for k, v := range decoded.Headers {
		for _, hv := range v {
			response.Header.Add(k, hv)
		}
	}

	return response, nil
}
