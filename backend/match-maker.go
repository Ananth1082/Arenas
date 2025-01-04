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

func matchMaking(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		// Greet the client
		if err := websocket.Message.Send(ws, "Hello, Client!"); err != nil {
			c.Logger().Error("Error sending greeting:", err)
			return
		}

		// Receive user info
		player1 := new(UserInfo)
		err := websocket.JSON.Receive(ws, player1)
		if err != nil {
			log.Println("Error receiving user info:", err)
		}
		log.Println("player1", player1)

		select {
		case player2 := <-MMQueue:
			games, err := client.Games.FindMany().Exec(context.Background())
			if err != nil || len(games) == 0 {
				log.Println("Error fetching games:", err)
				websocket.Message.Send(ws, `{"error": "No games available"}`)
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
				websocket.Message.Send(ws, `{"error": "Error creating match"}`)
				return
			}

			matchMap.Store(match.ID, Match{
				Players: [2]Player{
					{
						UserInfo: *player1,
						Conn:     nil,
						msgQueue: make(chan string, 1),
					},
					{
						UserInfo: player2,
						Conn:     nil,
						msgQueue: make(chan string, 1),
					},
				},
				gameState: GameState{
					duration: randGame.Time,
				},
			})

			websocket.JSON.Send(ws, match)
			isDone <- match
		case MMQueue <- *player1:
			websocket.JSON.Send(ws, `{"msg": "Waiting for player 2"}`)
			match := <-isDone
			data, _ := matchMap.Load(match.ID)
			log.Println("Match created:", data)
			websocket.JSON.Send(ws, match)
		}

	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
