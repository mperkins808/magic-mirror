# magic-mirror

HTTP requests routed through websockets. Establish remote connections from a server to localhosts

## install

With docker

```bash
docker pull mattyp123/magicmirror:0.0.1
```

As a binary. Tested on Mac, Results may vary

```
curl -sSL https://raw.githubusercontent.com/mperkins808/magic-mirror/main/install.sh\?foo\=bar | sudo bash
```

## usage

magic mirror is a client that connects to a server through a websocket. Then when HTTP requests are received it forwards them onto the specified host

| **Arg** | **Description**                                                                                                     |
| ------- | ------------------------------------------------------------------------------------------------------------------- |
| help    | display list of arguments                                                                                           |
| name    | optionally name your connection. Parsed as a request argument                                                       |
| remote  | the server to connect to                                                                                            |
| local   | optionally force all requests to go to a localhost. If environment variable `SAFE` = `true` then local is mandatory |
| apikey  | optionally attach a bearer token to the connection                                                                  |

**Establishing a connection with a server**

As a Binary

```bash
magicmirror --remote wss://myremotehost/mirrors --local http://localhost:9090 --apikey <INSERT> --name prometheus
```

With Docker

```bash
docker run mattyp123/magicmirror --remote wss://myremotehost/mirrors --local http://localhost:9090 --apikey <INSERT> --name prometheus
```

## Schema

magic mirror expects a specific message structure for requests and returns a specific response struct. All messages send to magicmirror must be base64 encoded and all messages magicmirror sends to the server will be base64 encoded. Messages should be base64 encodings of the following json objects

**Request Object**

```
{
    "method" : string,
    "uri" : string,
    "body" : []byte,
    "headers" : map[string][]string
}
```

**Response Object**

```
{
    "body" : []byte,
    "headers": map[string][]string
    "status_code" : int
}
```

**Example Request Object**

```json
{
  "method": "GET",
  "uri": "http://localhost:9090/metrics",
  "body": "",
  "headers": {
    "Authorization": ["Bearer FOOBAR"]
  }
}
```

**Example Response Object**

```json
{
  "body": "Foo bar Foo Bbar (byte array but as a string would look like this)",
  "headers": {
    "SomeHeader": {
      "Foo": "bar"
    }
  },
  "status_code": 200
}
```

## How a Go Server could handle the connection

Recommended to use my other package [socketmanager](https://github.com/mperkins808/socketmanager) to handle the requests

```go
func WSMirrorHandler(w http.ResponseWriter, r *http.Request, sm *socketmanager.SimpleSocketManager, upgrader websocket.Upgrader) {

    // handling authentication
	if IsValidAPIToken(r) {
		responses.Response(w, http.StatusForbidden, "token is invalid")
		return
	}

	// what the user calls the connection
    // in this server there must be a name
	name := r.URL.Query().Get("name")
	if name == "" {
		responses.Response(w, http.StatusBadRequest, "you need to supply a name")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}

	userid := RetrieveID(r)
	if err != nil {
		responses.Response(w, http.StatusForbidden, "token is invalid")
	}

	uid := fmt.Sprintf("%s-%s", userid, name)

    // adding the connection to socket manager, this allows us to use the connection in other parts of the server
	sm.Add(uid, name)
	sm.SetArb(uid, constants.MIRROR_CONNECTION, conn)
}

```

Then in other functions we can retrieve the connection and forward requests to it

Add the encoder to easily encode and decode messages

```bash
go get -u github.com/mperkins808/magic-mirror/go/pkg/encoder
```

```go
func ExampleHandler(w http.ResponseWriter, r *http.Request, sm *socketmanager.SimpleSocketManager) {

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:9090/metrics", nil)
    id := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")

    // extracting the connection from socketmanager

    uid := fmt.Sprintf("%s-%s", id, name)

    // remember we set this in the above function
	arb := sm.GetArb(uid, constants.MIRROR_CONNECTION)

	if arb.Err != nil {
        // connection not found
		return nil, arb.Err
        responses.Response(w, http.StatusInternalServerError, err.Error())

	}

	conn := arb.Value.(*websocket.Conn)

	encoded, err := encoder.EncodeRequest(req)
	if err != nil {
		log.Error("failed to encode request ", err)
		responses.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	conn.WriteMessage(1, []byte(encoded))

	_, msg, err := conn.ReadMessage()

	if err != nil {
		log.Error("failed to read connection message ", err)
		responses.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp, err := encoder.DecodeResponse(string(msg))
	if err != nil {
		log.Error("failed to read connection message ", err)
		responses.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	b, _ := io.ReadAll(resp.Body)
	responses.BuildableResponse(w, resp.StatusCode, b)

}
```
