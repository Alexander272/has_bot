package error_bot

import (
	"bytes"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/goccy/go-json"
)

type Message struct {
	Service *Service     `json:"service" binding:"required"`
	Data    *MessageData `json:"data" binding:"required"`
}

type Service struct {
	Id   string `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type MessageData struct {
	Date    string `json:"date" binding:"required"`
	Error   string `json:"error" binding:"required"`
	IP      string `json:"ip" binding:"required"`
	URL     string `json:"url" binding:"required"`
	Request string `json:"request"`
}

func Send(errMsg string, request interface{}) {
	var req []byte
	if request != nil {
		var err error
		req, err = json.MarshalIndent(request, "", "    ")
		if err != nil {
			slog.Error("failed to marshal request body.", slog.String("error", err.Error()))
		}
	}

	data := &MessageData{
		Date:    time.Now().Format("02/01/2006 - 15:04:05"),
		Error:   errMsg,
		Request: string(req),
	}

	message := Message{
		Service: &Service{
			Id:   os.Getenv("SERVICE_ID"),
			Name: os.Getenv("SERVICE_NAME"),
		},
		Data: data,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(message); err != nil {
		slog.Error("failed to encode struct.", slog.String("error", err.Error()))
	}

	url := os.Getenv("ERR_URL")
	if url == "" {
		return
	}

	_, err := http.Post(url, "application/json", &buf)
	if err != nil {
		slog.Error("failed to send error to bot.", slog.String("error", err.Error()))
	}
}