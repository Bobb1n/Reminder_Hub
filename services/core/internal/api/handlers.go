package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"core/internal/api/response"
	"core/internal/database"
	"core/internal/security"
	"core/internal/util"
	"reminder-hub/pkg/logger"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	db        database.DBer
	encryptor security.Encryptor
	log       *logger.CurrentLogger
}

func NewHandler(db database.DBer, encryptor security.Encryptor, log *logger.CurrentLogger) *Handler {
	return &Handler{
		db:        db,
		encryptor: encryptor,
		log:       log,
	}
}

func (h *Handler) CreateIntegration(c echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)
	if requestID != "" {
		ctx = logger.WithRequestID(ctx, requestID)
	}

	h.log.Info(ctx, "CreateIntegration called")

	var req database.CreateIntegrationRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	userID := c.Get(ContextKeyUserID).(string)
	h.log.Info(ctx, "Creating integration", "user_id", userID, "email", req.EmailAddress)

	encryptedPassword, err := h.encryptPassword(ctx, req.Password)
	if err != nil {
		return err
	}

	integration, err := h.createIntegrationRecord(ctx, &req, userID, encryptedPassword)
	if err != nil {
		return err
	}

	h.log.Info(ctx, "Saving to database", "integration_id", integration.ID)

	if err := h.db.CreateIntegration(ctx, integration); err != nil {
		h.log.Error(ctx, "Failed to create integration in DB", "error", err)
		if errors.Is(err, database.ErrDuplicateIntegration) {
			return c.JSON(http.StatusConflict, response.ErrorResponse{
				Error: "Integration already exists for this user and email",
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to create integration",
		})
	}

	h.log.Info(ctx, "Integration created successfully", "integration_id", integration.ID)

	return c.JSON(http.StatusCreated, response.CreateIntegrationResponse{
		ID:     integration.ID,
		Status: "created",
	})
} //"+absPath,

func (h *Handler) GetUserIntegrations(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("user_id")

	if _, err := uuid.Parse(userID); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid user ID format",
		})
	}

	integrations, err := h.db.GetUserIntegrations(ctx, userID)
	if err != nil {
		h.log.Error(ctx, "Failed to get integrations for user", "error", err, "user_id", userID)
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to get integrations",
		})
	}

	return c.JSON(http.StatusOK, integrations)
}

func (h *Handler) DeleteIntegration(c echo.Context) error {
	ctx := c.Request().Context()
	integrationID := c.Param("id")
	userID := c.Get(ContextKeyUserID).(string)

	if err := h.db.DeleteIntegration(ctx, userID, integrationID); err != nil {
		h.log.Error(ctx, "Failed to delete integration", "error", err, "integration_id", integrationID, "user_id", userID)
		if errors.Is(err, database.ErrIntegrationNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Integration not found or access denied",
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to delete integration",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) createIntegrationRecord(ctx context.Context, req *database.CreateIntegrationRequest, userID, encryptedPassword string) (*database.EmailIntegration, error) {
	integrationID, err := util.GenerateUUID()
	if err != nil {
		h.log.Error(ctx, "Failed to generate UUID", "error", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate ID")
	}

	integration := &database.EmailIntegration{
		ID:           integrationID,
		UserID:       userID,
		EmailAddress: strings.ToLower(strings.TrimSpace(req.EmailAddress)),
		ImapHost:     req.ImapHost,
		ImapPort:     req.ImapPort,
		UseSSL:       req.UseSSL,
		Password:     encryptedPassword,
	}
	return integration, nil
}

func (h *Handler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":    "healthy",
		"service":   "core",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func bindAndValidate(c echo.Context, req *database.CreateIntegrationRequest) error {
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func (h *Handler) encryptPassword(ctx context.Context, password string) (string, error) {
	encryptedPassword, err := h.encryptor.Encrypt(password)
	if err != nil {
		h.log.Error(ctx, "Failed to encrypt password", "error", err)
		return "", echo.NewHTTPError(http.StatusInternalServerError, "Failed to encrypt password")
	}
	return encryptedPassword, nil
}
