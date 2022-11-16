package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
)

type Event struct {
	Message       chan string
	NewClients    chan chan string
	ClosedClients chan chan string
	TotalClients  map[chan string]bool
}

type ClientChan chan string

type Message struct {
	Word string `json:"word"`
}

var msg = Message{Word: "First message"}

func main() {
	router := gin.Default()
	stream := NewServer()

	go func() {
		// Передаем слово клиентам с интервалом в 1 секунду
		s := gocron.NewScheduler(time.UTC)
		_, _ = s.Every(1).Seconds().Do(func() {
			stream.Message <- msg.Word
		})
		s.StartAsync()
	}()

	router.GET("/listen", HeadersMiddleware(), stream.serveHTTP(), getStreamMessages)
	router.POST("/say", changeWord)

	router.StaticFile("/", "./public/index.html")

	_ = router.Run(":8080")
}

// NewServer инициализирует событие и стартует прослушку запросов
func NewServer() (event *Event) {
	event = &Event{
		Message:       make(chan string),
		NewClients:    make(chan chan string),
		ClosedClients: make(chan chan string),
		TotalClients:  make(map[chan string]bool),
	}

	go event.listen()

	return
}

// Прослушка и обработка всех входящие запросов клиентов
func (stream *Event) listen() {
	for {
		select {
		// Добавление нового клиента
		case client := <-stream.NewClients:
			stream.TotalClients[client] = true
			log.Printf("Добавлен новый клиент. Всего зарегистрированных клиентов: %d", len(stream.TotalClients))

		// Удаление клиента
		case client := <-stream.ClosedClients:
			delete(stream.TotalClients, client)
			close(client)
			log.Printf("Клиент удален. Всего зарегистрированных клиентов: %d", len(stream.TotalClients))

		// Ващание сообщения всем зарегистрированным клиентам
		case eventMsg := <-stream.Message:
			for clientMessageChan := range stream.TotalClients {
				clientMessageChan <- eventMsg
			}
		}
	}
}

func (stream *Event) serveHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientChan := make(ClientChan)

		stream.NewClients <- clientChan

		defer func() {
			stream.ClosedClients <- clientChan
		}()

		c.Set("clientChan", clientChan)

		c.Next()
	}
}

func HeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}

func getStreamMessages(c *gin.Context) {
	v, ok := c.Get("clientChan")
	if !ok {
		return
	}
	clientChan, ok := v.(ClientChan)
	if !ok {
		return
	}
	c.Stream(func(w io.Writer) bool {
		// Stream message to client from message channel
		if msg, ok := <-clientChan; ok {
			c.SSEvent("message", msg)
			return true
		}
		return false
	})
}

func changeWord(c *gin.Context) {
	var newMessage Message
	if err := c.BindJSON(&newMessage); err != nil {
		return
	}
	msg = newMessage
	c.IndentedJSON(http.StatusCreated, newMessage)
}
