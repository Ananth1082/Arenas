// TODO: Implement websocket handler for the type testing game
package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Ananth1082/arenas/prisma/db"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"
)

var MMQueue = make(chan UserInfo, 1)
var isDone = make(chan *db.MatchesModel)

type ConnMap struct {
	sync.Map
}

func (c *ConnMap) Store(key string, value Match) {
	c.Map.Store(key, value)
}

func (c *ConnMap) Load(key string) (Match, bool) {
	val, ok := c.Map.Load(key)
	if !ok {
		return Match{}, false
	}
	return val.(Match), true
}

type UserInfo struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type GameInfo struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Duration int    `json:"duration,omitempty"`
}

type Player struct {
	UserInfo
	Conn     *websocket.Conn
	msgQueue chan string
}
type Match struct {
	Players   [2]Player
	gameState GameState //server held game state
}

type GameState struct {
	issueTime time.Time
	duration  int
	endTime   time.Time
}

var matchMap ConnMap
var matchSync sync.RWMutex

func wsReturnMsg(ws *websocket.Conn) string {
	msg := ""
	if err := websocket.Message.Receive(ws, &msg); err != nil {
		log.Println("Error sending message:", err)
	}
	return msg
}

func tugOfWar(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		// Greet the client
		if err := websocket.Message.Send(ws, "Hello, Client!"); err != nil {
			c.Logger().Error("Error sending greeting:", err)
			return
		}

		// Recieve match id
		userData := new(struct {
			UserID  string `json:"userId"`
			MatchID string `json:"matchId"`
		})
		err := websocket.JSON.Receive(ws, userData)
		if err != nil {
			log.Fatal(err)
		}
		match, ok := matchMap.Load(userData.MatchID)

		//identify user number
		i := 0
		if !ok {
			websocket.Message.Send(ws, `{"error": "Match has not started or has expired"}`)
			return
		}
		if match.Players[0].ID == userData.UserID {
			match.Players[0].Conn = ws
			i = 0
		} else {
			match.Players[1].Conn = ws
			i = 1
		}
		match.gameState.issueTime = time.Now()
		// give 10 seconds buffer time
		match.gameState.endTime = time.Now().Add(time.Duration(match.gameState.duration+10) * time.Second)
		websocket.JSON.Send(ws, map[string]any{
			"issueTime": match.gameState.issueTime,
			"startTime": match.gameState.issueTime.Add(10 * time.Second),
			"endTime":   match.gameState.endTime,
		})
		//sleep untill the game starts
		time.Sleep(time.Now().Sub(match.gameState.issueTime.Add(10 * time.Second)))

		//send game state to both players
		for {
			select {
			case msg := <-match.Players[1-i].msgQueue:
				//send game state
				websocket.Message.Send(match.Players[i].Conn, msg)
			case match.Players[i].msgQueue <- wsReturnMsg(ws):
			case <-time.After(match.gameState.endTime.Sub(time.Now())):
				//game over
				websocket.Message.Send(match.Players[i].Conn, "game over")
				return
			}
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
	e.GET("/ws/tug-of-war", tugOfWar)
	e.GET("/ws/match-making", matchMaking)

	e.Logger.Fatal(e.Start(":8080"))
}
