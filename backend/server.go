package main

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/Ananth1082/arenas/prisma/db"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"
)

var MMQueue = make(chan string, 1)
var isDone = make(chan *db.MatchesModel)

func matchMaking(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		// Greet the client
		if err := websocket.Message.Send(ws, "Hello, Client!"); err != nil {
			c.Logger().Error("Error sending greeting:", err)
			return
		}

		// Receive user info
		player1ID := ""
		err := websocket.Message.Receive(ws, &player1ID)
		if err != nil {
			c.Logger().Error(err)
		}

		select {
		case player2ID := <-MMQueue:
			games, err := client.Games.FindMany().Exec(context.Background())
			if err != nil || len(games) == 0 {
				websocket.Message.Send(ws, `{"error": "No games available"}`)
				return
			}
			randGame := games[rand.Intn(len(games))]
			match, err := client.Matches.CreateOne(
				db.Matches.Player1.Link(db.User.ID.Equals(player1ID)),
				db.Matches.Player2.Link(db.User.ID.Equals(player2ID)),
				db.Matches.Game.Link(db.Games.ID.Equals(randGame.ID)),
				db.Matches.Time.Set(db.DateTime(time.Now().Add(5*time.Minute))),
			).Exec(context.Background())
			if err != nil {
				websocket.Message.Send(ws, `{"error": "Error creating match"}`)
				return
			}
			isDone <- match
			websocket.JSON.Send(ws, match)
		case MMQueue <- player1ID:
			match := <-isDone
			websocket.JSON.Send(ws, match)
		}

	}).ServeHTTP(c.Response(), c.Request())
	return nil
}

type User struct {
	Name string `json:"name,omitempty"`
}

func server() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/user", func(c echo.Context) error {
		req := new(User)
		c.Bind(req)
		newUser, err := client.User.CreateOne(db.User.Name.Set(req.Name)).Exec(context.Background())
		if err != nil {
			c.Echo().Logger.Print(err)
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"msg": "username already exists"})
		}
		return c.JSON(http.StatusOK, echo.Map{"msg": "user created", "user": newUser})
	})

	e.GET("/ws/match-making", matchMaking)
	e.Logger.Fatal(e.Start(":8080"))
}
