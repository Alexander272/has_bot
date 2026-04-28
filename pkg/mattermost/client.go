package mattermost

import (
	"fmt"
	"log"

	"github.com/mattermost/mattermost-server/v6/model"
)

type Config struct {
	ServerLink string
	Token     string
}

type Client struct {
	Http   *model.Client4
	Socket *model.WebSocketClient
	apiUrl string
	token  string
}

func NewClient(conf Config) *Client {
	httpClient := model.NewAPIv4Client("https://"+conf.ServerLink)
	httpClient.SetToken(conf.Token)

	return &Client{
		Http:   httpClient,
		apiUrl: conf.ServerLink,
		token:  conf.Token,
	}
}

func (m *Client) Connect() bool {
	if m.Socket != nil {
		m.Socket.Close()
	}

	socket, err := model.NewWebSocketClient("wss://"+m.apiUrl, m.token)
	if err != nil {
		log.Printf("[!] Error connecting to the Mattermost WS: %s\n", err.Error())
		return false
	}
	m.Socket = socket
	m.Socket.Listen()
	log.Println("[+] Mattermost Websocket connection established")

	return true
}

func (m *Client) IsConnected() bool {
	if m.Socket != nil && m.Socket.ListenError != nil {
		log.Printf("[!] Error: Lost connect to the Mattermost WS: %s\n", m.Socket.ListenError.Error())
		return false
	}
	return true
}

func (m *Client) SendMessage(channelID, message string) error {
	post := &model.Post{
		ChannelId: channelID,
		Message:  message,
	}

	_, _, err := m.Http.CreatePost(post)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}

	return nil
}