package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSockServer struct {
	dataBuf chan []byte
	conn    *websocket.Conn
}

func NewWebSockServer() *WebSockServer {

	ws := new(WebSockServer)
	ws.dataBuf = make(chan []byte, 5)

	// sender thread to send msg to a client
	go func() {
		for {
			msg := <-ws.dataBuf
			if ws.conn == nil {
				fmt.Println("Nothing to reload\n")
				continue
			}
			//if err := conn.WriteMessage(1, []byte("reload")); err != nil {
			if err := ws.conn.WriteMessage(1, msg); err != nil {
				fmt.Println(err)
			}
		}
	}()

	return ws

}

func (ws *WebSockServer) WSHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	fmt.Println("new ws conn\n")
	ws.conn, err = websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	if err = ws.conn.WriteMessage(1, []byte("welcome")); err != nil {
		fmt.Println(err)
	}
}

func (ws *WebSockServer) Send(msg []byte) {
	ws.dataBuf <- msg
}
