package handlers

import (
	"net/http"
	"time"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/models"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReminderHandler struct {
	repo *repository.ReminderRepository
}

func NewReminderHandler(repo *repository.ReminderRepository) *ReminderHandler {
	return &ReminderHandler{repo: repo}
}

// CreateReminder создает напоминание
func (h *ReminderHandler) CreateReminder(c *gin.Context) {
	var req models.CreateReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reminder := &models.Reminder{
		ID:          uuid.New(),
		UserID:      req.UserID,
		SourceType:  req.SourceType,
		EmailRawID:  req.EmailRawID,
		Title:       req.Title,
		Description: req.Description,
		Deadline:    req.Deadline,
		IsCompleted: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.repo.Create(c.Request.Context(), reminder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, reminder)
}

// GetReminders возвращает напоминания пользователя
func (h *ReminderHandler) GetReminders(c *gin.Context) {
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

	reminders, err := h.repo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reminders)
}
