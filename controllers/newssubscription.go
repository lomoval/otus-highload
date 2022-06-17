package controllers

import (
	"app/service"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{}

// NewsSubscription handles WebSocket requests.
type NewsSubscription struct {
	Base
}

type client struct {
	userID  int64
	sockets map[*websocket.Conn]struct{}
}

var (
	clients = sync.Map{}
)

func init() {
	go handleMessages()
}

func handleMessages() {
	for {
		clients.Range(func(key, value interface{}) bool {
			client := value.(client)

			news, err := service.GetRabbitNews(client.userID)
			if err != nil {
				log.Err(err).Msgf("failed to get rabbit messages")
				return true
			}
			for _, n := range news {
				for socket, _ := range client.sockets {
					err := socket.WriteJSON(n)
					log.Debug().Msgf("  message %v", n)
					if err != nil {
						log.Err(err).Msgf("failed to send websocket message")
						socket.Close()
						delete(client.sockets, socket)
					}
				}
				if len(client.sockets) == 0 {
					clients.Delete(client.userID)
					service.RemoveRabbitUser(client.userID)
				}
			}
			return true
		})
		time.Sleep(200 * time.Millisecond)
	}
}

func (c *NewsSubscription) Get() {
	ws, err := upgrader.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request, nil)
	if err != nil {
		log.Err(err).Msgf("cannot setup WebSocket connection - upgrade failed")
		c.AbortInternalError()
		return
	}
	defer ws.Close()

	id := c.user().Id
	if _, ok := clients.Load(c.user().Id); !ok {
		clients.Store(c.user().Id, client{
			userID:  c.user().Id,
			sockets: map[*websocket.Conn]struct{}{ws: {}},
		})
	} else {
		value, _ := clients.Load(c.user().Id)
		value.(client).sockets[ws] = struct{}{}
	}

	for {
		if _, ok := clients.Load(id); ok {
			time.Sleep(time.Second * 5)
		} else {
			return
		}
	}
}
