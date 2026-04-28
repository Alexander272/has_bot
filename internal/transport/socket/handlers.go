package socket

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Alexander272/HasBot/internal/services"
	"github.com/Alexander272/HasBot/pkg/error_bot"
	"github.com/Alexander272/HasBot/pkg/logger"
	"github.com/goccy/go-json"
	"github.com/mattermost/mattermost-server/v6/model"
)

type Handler struct {
	socket   *model.WebSocketClient
	user    *model.User
	service *services.Service
	eventSem chan struct{}
}

func NewHandler(socket *model.WebSocketClient, user *model.User, service *services.Service) *Handler {
	return &Handler{
		socket:   socket,
		user:    user,
		service: service,
		eventSem: make(chan struct{}, 20),
	}
}

func (h *Handler) Run(ctx context.Context) {
	logger.Info("socket worker started")
	defer logger.Info("socket worker stopped")

	backoff := time.Second * 2
	maxBackoff := time.Minute

	for {
		select {
		case <-ctx.Done():
			h.socket.Close()
			return
		default:
		}

		if err := h.runConnection(ctx, &backoff, maxBackoff); err != nil {
			logger.Debug("connection loop exited", logger.ErrAttr(err))
		}
	}
}

func (h *Handler) runConnection(ctx context.Context, backoff *time.Duration, maxBackoff time.Duration) error {
	logger.Info("connecting to websocket...")

	h.socket.Listen()

	if h.socket.ListenError != nil {
		logger.Error("listen error", logger.ErrAttr(h.socket.ListenError))
		h.applyBackoff(ctx, backoff, maxBackoff)
		return h.socket.ListenError
	}

	logger.Info("websocket connected")
	*backoff = time.Second * 2

	if err := h.consumeEvents(ctx); err != nil {
		logger.Info("event loop stopped", logger.ErrAttr(err))
	}

	logger.Info("reconnecting...")
	h.applyBackoff(ctx, backoff, maxBackoff)
	return nil
}

func (h *Handler) applyBackoff(ctx context.Context, backoff *time.Duration, maxBackoff time.Duration) {
	select {
	case <-time.After(*backoff):
	case <-ctx.Done():
		return
	}
	*backoff = minDuration((*backoff)*2, maxBackoff)
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func (h *Handler) Listen() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h.Run(ctx)
}

func (h *Handler) Close() {
	h.socket.Close()
}

func (h *Handler) consumeEvents(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event, ok := <-h.socket.EventChannel:
			if !ok {
				return errors.New("event channel closed")
			}

			if event.EventType() != "posted" {
				continue
			}

			select {
			case h.eventSem <- struct{}{}:
				go func() {
					defer func() { <-h.eventSem }()
					h.safeHandleEvent(event)
				}()
			default:
				logger.Info("event dropped: too many concurrent handlers")
			}
		}
	}
}

func (h *Handler) safeHandleEvent(event *model.WebSocketEvent) {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			logger.Error("panic recovered in event handler", logger.AnyAttr("recover", r))
		}
	}()

	h.handleEvent(event)
}

func (h *Handler) handleEvent(event *model.WebSocketEvent) {
	channelId := event.GetBroadcast().ChannelId

	post, err := h.parsePost(event)
	if err != nil {
		logger.Error("failed to parse post", logger.ErrAttr(err))
		error_bot.Send("failed to parse post", post)
		return
	}

	if post.UserId == h.user.Id {
		logger.Debug("skipping own message")
		return
	}

	if post.Type != "" {
		logger.Debug("skipping system message", logger.AnyAttr("type", post.Type))
		return
	}

	post.Message = strings.TrimSpace(post.Message)
	logger.Debug("message received", logger.AnyAttr("message", post.Message), logger.AnyAttr("channel", channelId))

	if err := h.service.HandleMessage(channelId, post.Message); err != nil {
		logger.Error("command error", logger.ErrAttr(err))
		error_bot.Send(err.Error(), post.Message)
	}
}

func (h *Handler) parsePost(event *model.WebSocketEvent) (*model.Post, error) {
	postData, ok := event.GetData()["post"]
	if !ok {
		return nil, errors.New("post key missing in event data")
	}

	postStr, ok := postData.(string)
	if !ok {
		return nil, fmt.Errorf("post is not a string, got %T", postData)
	}

	var post model.Post
	if err := json.Unmarshal([]byte(postStr), &post); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	return &post, nil
}
