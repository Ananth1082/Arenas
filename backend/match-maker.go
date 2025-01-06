package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/Ananth1082/arenas/prisma/db"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type Message struct {
	// Type 0: log, 1: info, 2: error,3: user message, 4: ping (no data)
	Type int         `json:"type"`
	Data interface{} `json:"data"`
}

func matchMaking(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		// Greet the client
		if err := websocket.JSON.Send(ws, Message{
			Type: 0,
			Data: "Hello, welcome to the match making service",
		}); err != nil {
			c.Logger().Error("Error sending greeting:", err)
			return
		}

		player1 := new(UserInfo)
		err := websocket.JSON.Receive(ws, player1)
		if err != nil {
			log.Println("Error receiving user info:", err)
			websocket.JSON.Send(ws, Message{
				Type: 2,
				Data: "Error receiving user info",
			})
		}
		log.Println("player1", player1)

		select {
		case player2 := <-MMQueue:
			games, err := client.Games.FindMany().Exec(context.Background())
			if err != nil || len(games) == 0 {
				log.Println("Error fetching games:", err)
				websocket.JSON.Send(ws, Message{
					Type: 2,
					Data: "Error fetching games",
				})
				return
			}
			randGame := games[rand.Intn(len(games))]
			match, err := client.Matches.CreateOne(
				db.Matches.Player1.Link(db.User.ID.Equals(player1.ID)),
				db.Matches.Player2.Link(db.User.ID.Equals(player2.ID)),
				db.Matches.Game.Link(db.Games.ID.Equals(randGame.ID)),
				db.Matches.Time.Set(db.DateTime(time.Now().Add(5*time.Minute))),
			).Exec(context.Background())
			if err != nil {
				log.Println("Error creating match:", err)
				websocket.JSON.Send(ws, Message{
					Type: 2,
					Data: "Error creating match",
				})
				return
			}
			matchMap.Store(match.ID, Match{
				Players: [2]Player{
					{
						UserInfo: *player1,
						Conn:     nil,
						msgQueue: make(chan string, 5),
					},
					{
						UserInfo: player2,
						Conn:     nil,
						msgQueue: make(chan string, 5),
					},
				},
				gameState: GameState{
					issueTime: time.Now(),
					duration:  5,
					endTime:   time.Now().Add(BUFF_TIME).Add(time.Duration(5) * time.Second),
				},
			})

			websocket.JSON.Send(ws, Message{
				Type: 1,
				Data: match,
			})
			isDone <- match
		case MMQueue <- *player1:
			websocket.JSON.Send(ws, Message{
				Type: 0,
				Data: "Waiting for another player",
			})
			match := <-isDone
			websocket.JSON.Send(ws, Message{
				Type: 1,
				Data: match,
			})
		}

	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
