package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Ananth1082/arenas/prisma/db"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"
)

func hello(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			err := websocket.Message.Send(ws, "Hello, Client!")
			if err != nil {
				c.Logger().Error(err)
			}

			msg := ""
			err = websocket.Message.Receive(ws, &msg)
			if err != nil {
				c.Logger().Error(err)
			}
			fmt.Printf("%s\n", msg)
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

	e.GET("/ws", hello)

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

	e.Logger.Fatal(e.Start(":8080"))
}
