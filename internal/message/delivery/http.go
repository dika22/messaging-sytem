package delivery

import (
	"multi-tenant-service/internal/message/usecase"
	"multi-tenant-service/package/response"
	"multi-tenant-service/package/structs"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MessageHandler struct {
	messageUsecase usecase.IMessageUsecase
}

// PublishMessage godoc
// @Summary Publish a message
// @Description Publish a message to a tenant's queue
// @Tags messages
// @Accept json
// @Produce json
// @Param message body structs.CreateMessageRequest true "Message data"
// @Success 202 {object} structs.Response
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /messages [post]
func (h *MessageHandler) PublishMessage(c echo.Context) error {
	ctx := c.Request().Context()
	var req structs.CreateMessageRequest
	if err := c.Bind(&req); err != nil {
		return response.JSONResponse(c, http.StatusBadRequest, false, err.Error(), nil)
	}

	if err := h.messageUsecase.PublishMessage(ctx, req); err != nil {
		return response.JSONResponse(c, http.StatusInternalServerError, false, err.Error(), nil)
	}
	return response.JSONResponse(c, http.StatusAccepted, true, "Message published successfully", nil)
}

// GetMessages godoc
// @Summary Get messages with cursor pagination
// @Description Get messages for a tenant with cursor-based pagination
// @Tags messages
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param cursor query string false "Cursor for pagination"
// @Param limit query int false "Limit number of results" default(10)
// @Success      200      {object}  structs.Response
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /messages [get]
func (h *MessageHandler) GetMessages(c echo.Context) error {
	ctx := c.Request().Context()
	tenantIDStr := c.QueryParam("tenant_id")
	if tenantIDStr == "" {
		return response.JSONResponse(c, http.StatusBadRequest, false, "Missing tenant_id", nil)
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return response.JSONResponse(c, http.StatusBadRequest, false, "Invalid tenant_id", nil)
	}

	cursor := c.QueryParam("cursor")
	var cursorPtr *string
	if cursor != "" {
		cursorPtr = &cursor
	}

	limitStr := c.QueryParam("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	requestGetMessage := structs.RequestGetMessage{
		TenantID: tenantID,
		Cursor:   cursorPtr,
		Limit:    limit,
	}

	messages, err := h.messageUsecase.GetMessages(ctx, requestGetMessage)
	if err != nil {
		return response.JSONResponse(c, http.StatusInternalServerError, false, err.Error(), nil)
	}
	return response.JSONSuccess(c, messages, "Messages retrieved successfully")
}

func NewMessageHandler(e *echo.Group, messageUsecase usecase.IMessageUsecase) *MessageHandler {
	return &MessageHandler{
		messageUsecase: messageUsecase,
	}
}

func NewMessageHTTPHandler(r *echo.Group, messageUsecase usecase.IMessageUsecase)  {
	h := &MessageHandler{
		messageUsecase: messageUsecase,
	}
	r.POST("/messages", h.PublishMessage).Name = "PublishMessage"
	r.GET("/messages", h.GetMessages).Name = "GetMessages"
}