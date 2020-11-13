package admin

import (
	"container/list"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	Author  string `json:"author"`
	Message string `json:"message"`
}

type Session struct {
	SessionID string `json:"sessionid"`
	Extra     string `json:"extra"`
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WShandler(w http.ResponseWriter, r *http.Request, connections *list.List) {
	wsupgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: ", err)
		return
	}

	listentry := connections.PushBack(conn)
	conn.WriteJSON(&Session{SessionID: "90kj23lkj09", Extra: "SESSION EXTRAS "})
	conn.WriteJSON(&Message{Author: "KALLE", Message: "This is a message from Mr ANKA"})

	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		fmt.Println("RECEIVED MESSAGE: ", t, string(msg))

		for e := connections.Front(); e != nil; e = e.Next() {
			conn := e.Value.(*websocket.Conn)
			conn.WriteMessage(t, msg)
		}
	}

	connections.Remove(listentry)
}
