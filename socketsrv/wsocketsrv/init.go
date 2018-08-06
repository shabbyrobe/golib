package wsocketsrv

import "github.com/gorilla/websocket"

var ping *websocket.PreparedMessage

func init() {
	var err error
	ping, err = websocket.NewPreparedMessage(websocket.PingMessage, nil)
	if err != nil {
		panic(err)
	}
}
