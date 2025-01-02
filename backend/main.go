package main

import (
	"fmt"
	"log"

	"github.com/Ananth1082/arenas/prisma/db"
)

var client = dbConnect()

func main() {
	fmt.Println("Hello world")

	server()

	defer func() {
		if err := client.Prisma.Disconnect(); err != nil {
			panic(err)
		}
	}()
}

func dbConnect() *db.PrismaClient {
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		log.Println("Error connecting to the db", err)
	}
	return client
}
