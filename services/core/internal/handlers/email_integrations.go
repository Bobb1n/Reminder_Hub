package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/models"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/rabbitmq"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EmailIntegrationHandler struct {
	integrationRepo *repository.EmailIntegrationRepository
	emailRawRepo    *repository.EmailRawRepository
	rabbitProducer  *rabbitmq.Producer
}

func NewEmailIntegrationHandler(
	integrationRepo *repository.EmailIntegrationRepository,
	emailRawRepo *repository.EmailRawRepository,
	rabbitProducer *rabbitmq.Producer,
) *EmailIntegrationHandler {
	return &EmailIntegrationHandler{
		integrationRepo: integrationRepo,
		emailRawRepo:    emailRawRepo,
		rabbitProducer:  rabbitProducer,
	}
}

// CreateEmailIntegration создает новую почтовую интеграцию
func (h *EmailIntegrationHandler) CreateEmailIntegration(c *gin.Context) {
	var req models.CreateEmailIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	integration := &models.EmailIntegration{
		ID:           uuid.New(),
		UserID:       req.UserID,
		EmailAddress: req.EmailAddress,
		IMAPHost:     req.IMAPHost,
		IMAPPort:     req.IMAPPort,
		UseSSL:       req.UseSSL,
		Password:     req.Password,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.integrationRepo.Create(c.Request.Context(), integration); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Не возвращаем пароль в ответе
	integration.Password = ""
	c.JSON(http.StatusCreated, integration)
}

// GetEmailIntegrations возвращает все интеграции пользователя
func (h *EmailIntegrationHandler) GetEmailIntegrations(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	integrations, err := h.integrationRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Убираем пароли из ответа
	for _, integration := range integrations {
		integration.Password = ""
	}

	c.JSON(http.StatusOK, integrations)
}

// DeleteEmailIntegration удаляет почтовую интеграцию
func (h *EmailIntegrationHandler) DeleteEmailIntegration(c *gin.Context) {
	integrationIDStr := c.Param("id")
	integrationID, err := uuid.Parse(integrationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	// Также удаляем все сырые письма пользователя, связанные с этой интеграцией
	integration, err := h.integrationRepo.FindByID(c.Request.Context(), integrationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	// Удаляем интеграцию
	if err := h.integrationRepo.Delete(c.Request.Context(), integrationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Удаляем связанные письма
	if err := h.emailRawRepo.DeleteByUserID(c.Request.Context(), integration.UserID); err != nil {
		// Логируем ошибку, но не прерываем выполнение
		log.Printf("Warning: failed to delete user emails: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Integration deleted successfully"})
}
