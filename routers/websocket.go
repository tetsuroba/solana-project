package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func TransactionBroadcasterWebsocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		err = conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}
}
