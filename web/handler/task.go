package handler

import (
	"devcloud/ent/resource"
	"devcloud/web/handler/dto"
	"devcloud/web/handler/reqctx"
	"devcloud/web/handler/response"
	"devcloud/web/service"
	"strconv"

	"github.com/labstack/echo/v5"
)

type Task struct {
	taskSrv *service.Task
}

func NewTask(taskSrv *service.Task) *Task {
	return &Task{taskSrv: taskSrv}
}

func (h *Task) List(c *echo.Context) error {
	var req dto.PaginationRequest
	if err := c.Bind(&req); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}
	req.SetDefaults()

	if err := c.Validate(&req); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	rows, count, err := h.taskSrv.List(
		c.Request().Context(),
		user.ID,
		req.Page,
		req.PageSize,
	)
	if err != nil {
		return err
	}

	var tasks []dto.TaskInfoResponse
	for _, item := range rows {
		tasks = append(tasks, dto.TaskInfoResponse{
			ID:           item.ID,
			ResourceID:   item.ResourceID,
			Type:         item.Type.String(),
			Status:       item.Status.String(),
			Attempts:     item.Attempts,
			MaxAttempts:  item.MaxAttempts,
			ErrorMessage: item.ErrorMessage,
			StartedAt:    item.StartedAt,
			FinishedAt:   item.FinishedAt,
			CreatedAt:    item.CreatedAt,
		})
	}

	return response.JsonList(c, tasks, req.Page, req.PageSize, count)
}

func (h *Task) Get(c *echo.Context) error {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	t, err := h.taskSrv.Get(c.Request().Context(), user.ID, taskID)
	if err != nil {
		return err
	}

	return response.JsonData(c, dto.TaskInfoResponse{
		ID:           t.ID,
		ResourceID:   t.ResourceID,
		Type:         t.Type.String(),
		Status:       t.Status.String(),
		Attempts:     t.Attempts,
		MaxAttempts:  t.MaxAttempts,
		ErrorMessage: t.ErrorMessage,
		StartedAt:    t.StartedAt,
		FinishedAt:   t.FinishedAt,
		CreatedAt:    t.CreatedAt,
	})
}

func (h *Task) CreateResource(c *echo.Context) error {
	var req dto.CreateResourceRequest
	if err := c.Bind(&req); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}
	if err := c.Validate(&req); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	rsc, t, err := h.taskSrv.CreateResource(
		c.Request().Context(),
		user.ID,
		req.Name,
		resource.Provider(req.Provider),
		req.Image,
		req.Config,
	)
	if err != nil {
		return err
	}

	return response.JsonData(c, dto.CreateResourceResponse{
		Resource: dto.ResourceResponse{
			ID:       rsc.ID,
			Name:     rsc.Name,
			Image:    rsc.Image,
			Provider: rsc.Provider.String(),
			Status:   rsc.Status.String(),
		},
		TaskID: t.ID,
	})
}

func (h *Task) RemoveResource(c *echo.Context) error {
	resourceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	t, err := h.taskSrv.RemoveResource(
		c.Request().Context(),
		user.ID,
		resourceID,
	)
	if err != nil {
		return err
	}

	return response.JsonData(c, dto.ControlResourceResponse{
		TaskID: t.ID,
	})
}

func (h *Task) StartResource(c *echo.Context) error {
	resourceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	t, err := h.taskSrv.StartResource(
		c.Request().Context(),
		user.ID,
		resourceID,
	)
	if err != nil {
		return err
	}

	return response.JsonData(c, dto.ControlResourceResponse{
		TaskID: t.ID,
	})
}

func (h *Task) StopResource(c *echo.Context) error {
	resourceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	t, err := h.taskSrv.StopResource(
		c.Request().Context(),
		user.ID,
		resourceID,
	)
	if err != nil {
		return err
	}

	return response.JsonData(c, dto.ControlResourceResponse{
		TaskID: t.ID,
	})
}

func (h *Task) RestartResource(c *echo.Context) error {
	resourceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	t, err := h.taskSrv.RestartResource(
		c.Request().Context(),
		user.ID,
		resourceID,
	)
	if err != nil {
		return err
	}

	return response.JsonData(c, dto.ControlResourceResponse{
		TaskID: t.ID,
	})
}
