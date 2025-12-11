package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/collector/internal/database"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/collector/internal/service"
	"github.com/labstack/echo/v4"
	"gopkg.in/go-playground/validator.v9"
)

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{
		service: service,
	}
}

func (h *TaskHandler) GetTask(c echo.Context) error {
	ctx := c.Request().Context()

	taskID := c.Param("id")
	userID := c.Get(ContextKeyUserID).(string)

	task, err := h.service.GetTask(ctx, taskID, userID)
	if err != nil {
		if errors.Is(err, database.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Task not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) GetTasks(c echo.Context) error {
	ctx := c.Request().Context()

	userID := c.Get(ContextKeyUserID).(string)

	filter := parseTaskFilter(c, userID)

	tasks, err := h.service.GetUserTasks(ctx, userID, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) UpdateTask(c echo.Context) error {
	ctx := c.Request().Context()

	taskID := c.Param("id")
	userID := c.Get(ContextKeyUserID).(string)

	var update database.UpdateTaskRequest
	if err := c.Bind(&update); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	validate := validator.New()
	if err := validate.Struct(update); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := h.service.UpdateTask(ctx, taskID, userID, update); err != nil {
		if errors.Is(err, database.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Task not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Task updated successfully"})
}

func (h *TaskHandler) DeleteTask(c echo.Context) error {
	ctx := c.Request().Context()

	taskID := c.Param("id")
	userID := c.Get(ContextKeyUserID).(string)

	if err := h.service.DeleteTask(ctx, taskID, userID); err != nil {
		if errors.Is(err, database.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Task not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Task deleted successfully"})
}

func (h *TaskHandler) CompleteTask(c echo.Context) error {
	ctx := c.Request().Context()

	taskID := c.Param("id")
	userID := c.Get(ContextKeyUserID).(string)

	if err := h.service.CompleteTask(ctx, taskID, userID); err != nil {
		if errors.Is(err, database.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Task not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Task completed successfully"})
}

func (h *TaskHandler) GetStats(c echo.Context) error {
	ctx := c.Request().Context()

	userID := c.Get(ContextKeyUserID).(string)

	stats, err := h.service.GetStats(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, stats)
}

func (h *TaskHandler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "collector",
	})
}

func parseTaskFilter(c echo.Context, userID string) database.TaskFilter {
	filter := database.TaskFilter{
		UserID: userID,
		Limit:  50,
		Offset: 0,
	}

	if status := c.QueryParam("status"); status != "" {
		filter.Status = &status
	}

	if priority := c.QueryParam("priority"); priority != "" {
		filter.Priority = &priority
	}

	if fromStr := c.QueryParam("from_deadline"); fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.FromDeadline = &from
		}
	}

	if toStr := c.QueryParam("to_deadline"); toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			filter.ToDeadline = &to
		}
	}

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filter.Limit = limit
		}
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	return filter
}
