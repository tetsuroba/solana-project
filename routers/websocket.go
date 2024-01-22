package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"solana/models"
	"sync"
)

var (
	clients   sync.Map // Connected clients
	broadcast = make(chan models.SolanaPayload)
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func TransactionSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to set websocket upgrade", err)
		return
	}

	clients.Store(conn, true)

	err = conn.WriteJSON(gin.H{"message": "Connected to transaction socket"})
	if err != nil {
		logger.Error("Error writing to websocket", "error", err)
		return
	}
}

// StartWebSocketManager starts a goroutine that sends messages to WebSocket clients
func StartWebSocketManager() {
	go func() {
		for {
			msg := <-broadcast

			clients.Range(func(key, value interface{}) bool {
				client, ok := key.(*websocket.Conn)
				if !ok {
					return true
				}

				err := client.WriteJSON(msg)
				if err != nil {
					logger.Error("Error writing to websocket", "error", err)
					err = client.Close()
					if err != nil {
						logger.Error("Error closing websocket", "error", err)
						return false
					}
					disconnectClient(client)
				}
				return true
			})
		}
	}()
}

func disconnectClient(conn *websocket.Conn) {
	clients.Delete(conn)
}
