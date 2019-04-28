package main

import (
	"log"
	"mahjong"
	"math/rand"
	"net/http"
	"time"

	"github.com/rs/cors"
)

func main() {
	rand.Seed(time.Now().Unix())

	err := mahjong.NewGameManager()
	if err {
		return
	}
	go mahjong.Exec()

	mahjong.GetServer().On("connection", mahjong.SocketConnect)
	mahjong.GetServer().On("error", mahjong.SocketError)

	mux := http.NewServeMux()
	mux.Handle("/socket.io/", mahjong.GetServer())
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:9000"},
		AllowCredentials: true,
	})
	handler := c.Handler(mux)
	baseURL := "localhost:3000"
	log.Println("Serving at", baseURL, "...")
	log.Fatal(http.ListenAndServe(baseURL, handler))
}
