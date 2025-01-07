package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Ananth1082/arenas/prisma/db"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"
)

const BUFF_TIME = 4 * time.Second

var MMQueue = make(chan UserInfo, 1)
var isDone = make(chan *db.MatchesModel)

type ConnMap struct {
	sync.Map
}

func (c *ConnMap) Store(key string, value *Match) {
	c.Map.Store(key, value)
}

func (c *ConnMap) Load(key string) (*Match, bool) {
	val, ok := c.Map.Load(key)
	if !ok {
		return nil, false
	}
	return val.(*Match), true
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
	Players [2]Player
	game    Game //server held game state
}

type Game struct {
	issueTime time.Time
	duration  int
	endTime   time.Time
}

var matchMap ConnMap
var matchSync sync.RWMutex

func tugOfWar() {

}

func handleGameConnection(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		done := make(chan struct{})
		defer func() {

			log.Println("Closing connection")
			ws.Close()
		}()

		if err := websocket.JSON.Send(ws, Message{
			Type: 0,
			Data: "Hello, welcome to the tug of war game",
		}); err != nil {
			c.Logger().Error("Error sending greeting:", err)
			return
		}

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
			websocket.Message.Send(ws, Message{
				Type: 2,
				Data: "Match not found",
			})
			return
		}

		if match.Players[0].ID == userData.UserID {
			match.Players[0].Conn = ws
			i = 0
		} else {
			match.Players[1].Conn = ws
			i = 1
		}

		ws.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		for {
			var msg Message
			if err := websocket.JSON.Receive(ws, &msg); err != nil {
				break
			}
		}
		count := 0
		ws.SetReadDeadline(match.game.issueTime.Add(BUFF_TIME))
		time.Sleep(match.game.issueTime.Add(BUFF_TIME).Sub(time.Now()))
		ws.SetReadDeadline(time.Time{}) //reset read deadline

		endDuration := time.Until(match.game.endTime)
		log.Println("End duration:", endDuration)
		timer := time.NewTimer(endDuration)

		//message dequeue
		// sends the opps game state
		go func() {
			for {
				select {
				case msg := <-match.Players[1-i].msgQueue:
					if msg != "" {
						log.Println("Received message from player", 1-i, ":", msg)

						err := websocket.JSON.Send(match.Players[i].Conn, Message{
							Type: 3,
							Data: msg,
						})
						if err == io.EOF {
							log.Println("Exiting message dequeue")
							return
						}
					}
				case <-done:
					log.Println("DONE")
					return
				}
			}

		}()

		//message enqueue
		// controlls the user game state
		go func() {
			for {
				msg := ""
				if err := websocket.Message.Receive(ws, &msg); err != nil {
					log.Println("Error sending message:", err)
					if err == io.EOF {
						return
					}
				}
				if msg == "PING" {
					count++
				}
				select {
				case <-done:
					return
				case match.Players[i].msgQueue <- fmt.Sprint(count):
					log.Println("Sent message to player", i)
				}
			}
		}()

		<-timer.C
		close(done)
		match.Players[1-i].msgQueue <- fmt.Sprint(count)
		oppCount, _ := strconv.Atoi(<-match.Players[1-i].msgQueue)
		goodbyeMessage := Message{
			Type: 0,
			Data: "Draw",
		}
		if count > oppCount {
			goodbyeMessage.Data = "You Win"
		} else if count < oppCount {
			goodbyeMessage.Data = "You Lose"
		}
		if err := websocket.JSON.Send(ws, goodbyeMessage); err != nil {
			log.Println("Error sending goodbye message:", err)
		} else {
			log.Println("Goodbye message sent successfully")
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}

type User struct {
	Name string `json:"name,omitempty"`
}

func server() {
	e := echo.New()
	// e.Use(middleware.Logger())
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
	e.GET("/ws/tug-of-war", handleGameConnection)
	e.GET("/ws/match-making", matchMaking)

	e.Logger.Fatal(e.Start(":8080"))
}
