package httpapi

import (
	"net/http"

	"github.com/Alexander272/HasBot/internal/models"
	"github.com/Alexander272/HasBot/internal/services"
	"github.com/Alexander272/HasBot/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *services.Service
}

func NewHandler(service *services.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(r *gin.Engine) {
	api := r.Group("/api")
	api.GET("/channels", h.getChannels)
	api.POST("/channels", h.replaceChannels)
	api.DELETE("/channels/:channel_id", h.deleteChannel)
}

type channelsRequest struct {
	Channels []models.ChannelConfig `json:"channels"`
}

func (h *Handler) getChannels(c *gin.Context) {
	channels := h.service.GetChannels()
	if channels == nil {
		channels = []models.ChannelConfig{}
	}
	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

func (h *Handler) replaceChannels(c *gin.Context) {
	var req channelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
		return
	}

	if err := h.service.ReplaceChannels(req.Channels); err != nil {
		logger.Error("failed to replace channels", logger.ErrAttr(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save channels: " + err.Error()})
		return
	}

	logger.Info("channels replaced via API", logger.IntAttr("channels count", len(req.Channels)))
	c.JSON(http.StatusOK, gin.H{"channels": req.Channels})
}

func (h *Handler) deleteChannel(c *gin.Context) {
	channelID := c.Param("channel_id")

	if err := h.service.DeleteChannel(channelID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	logger.Info("channel deleted via API", logger.AnyAttr("channel_id", channelID))
	c.JSON(http.StatusOK, gin.H{"message": "channel deleted"})
}
