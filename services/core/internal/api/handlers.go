package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/api/response"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/database"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/security"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	db        database.DBer
	encryptor security.Encryptor
}

func NewHandler(db database.DBer, encryptor security.Encryptor) *Handler {
	return &Handler{
		db:        db,
		encryptor: encryptor,
	}
}



func (h *Handler) CreateIntegration(c echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)
	logger := log.With().Str("request_id", requestID).Logger()

	logger.Info().Msg("CreateIntegration called")

	var req database.CreateIntegrationRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	userID := c.Get(ContextKeyUserID).(string)
	logger.Info().
		Str("user_id", userID).
		Str("email", req.EmailAddress).
		Msg("Creating integration")

	encryptedPassword, err := h.encryptPassword(req.Password, logger)
	if err != nil {
		return err
	}

	integration, err := h.createIntegrationRecord(&req, userID, encryptedPassword, logger)
	if err != nil {
		return err
	}

	logger.Info().Str("integration_id", integration.ID).Msg("Saving to database")

	if err := h.db.CreateIntegration(ctx, integration); err != nil {
		logger.Error().Err(err).Msg("Failed to create integration in DB")
		if errors.Is(err, database.ErrDuplicateIntegration) {
			return c.JSON(http.StatusConflict, response.ErrorResponse{
				Error: "Integration already exists for this user and email",
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to create integration",
		})
	}

	logger.Info().Str("integration_id", integration.ID).Msg("Integration created successfully")

	return c.JSON(http.StatusCreated, response.CreateIntegrationResponse{
		ID:     integration.ID,
		Status: "created",
	})
}

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
		log.Error().Err(err).Msgf("Failed to get integrations for user %s", userID)
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
		log.Error().Err(err).Msgf("Failed to delete integration %s for user %s", integrationID, userID)
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

func (h *Handler) createIntegrationRecord(req *database.CreateIntegrationRequest, userID, encryptedPassword string, logger zerolog.Logger) (*database.EmailIntegration, error) {
	integrationID, err := util.GenerateUUID()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate UUID")
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

func (h *Handler) encryptPassword(password string, logger zerolog.Logger) (string, error) {
	encryptedPassword, err := h.encryptor.Encrypt(password)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to encrypt password")
		return "", echo.NewHTTPError(http.StatusInternalServerError, "Failed to encrypt password")
	}
	return encryptedPassword, nil
}
