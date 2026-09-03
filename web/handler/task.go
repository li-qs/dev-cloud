package handler

import (
	"devcloud/service"
	"devcloud/web/handler/dto"
	"devcloud/web/handler/reqctx"
	"devcloud/web/handler/response"
	"strconv"

	"github.com/labstack/echo/v5"
)

type Task struct {
	taskSrv *service.Task
}

func NewTask(taskSrv *service.Task) *Task {
	return &Task{taskSrv: taskSrv}
}

func (t *Task) Get(c *echo.Context) error {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	task, err := t.taskSrv.GetTask(c.Request().Context(), user.ID, taskID)
	if err != nil {
		return err
	}

	return response.JsonData(c, dto.TaskInfoResponse{
		ID:           task.ID,
		ResourceID:   task.ResourceID,
		Type:         task.Type.String(),
		Status:       task.Status.String(),
		Attempts:     task.Attempts,
		MaxAttempts:  task.MaxAttempts,
		ErrorMessage: task.ErrorMessage,
		StartedAt:    task.StartedAt,
		FinishedAt:   task.FinishedAt,
		CreatedAt:    task.CreatedAt,
	})
}
