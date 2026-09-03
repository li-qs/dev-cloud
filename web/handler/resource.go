package handler

import (
	"devcloud/service"
	"devcloud/web/dto"
	"devcloud/web/reqctx"
	"devcloud/web/response"
	"strconv"

	"github.com/labstack/echo/v5"
)

type Resource struct {
	resourceSrv *service.Resource
}

func NewResource(resourceSrv *service.Resource) *Resource {
	return &Resource{resourceSrv: resourceSrv}
}

func (r *Resource) List(c *echo.Context) error {
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

	rs, count, err := r.resourceSrv.List(c.Request().Context(), user.ID, req.Page, req.PageSize)
	if err != nil {
		return err
	}

	items := make([]dto.ResourceResponse, count)
	for i, r := range rs {
		items[i] = dto.ResourceResponse{
			ID:       r.ID,
			Name:     r.Name,
			Type:     r.Type,
			Provider: r.Provider,
			Status:   r.Status.String(),
		}
	}

	return response.JsonList(c, items, req.Page, req.PageSize, count)
}

func (r *Resource) Get(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	rc, err := r.resourceSrv.Get(c.Request().Context(), user.ID, id)
	if err != nil {
		return err
	}
	return response.JsonData(c, dto.ResourceResponse{
		ID:       rc.ID,
		Name:     rc.Name,
		Type:     rc.Type,
		Provider: rc.Provider,
		Status:   rc.Status.String(),
	})
}

func (r *Resource) Create(c *echo.Context) error {
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

	rc, t, err := r.resourceSrv.Create(
		c.Request().Context(),
		user.ID,
		req.Name,
		req.Provider,
		req.Config,
	)
	if err != nil {
		return err
	}

	return response.JsonData(c, dto.CreateResourceResponse{
		Resource: dto.ResourceResponse{
			ID:       rc.ID,
			Name:     rc.Name,
			Type:     rc.Type,
			Provider: rc.Provider,
			Status:   rc.Status.String(),
		},
		TaskID: t.ID,
	})
}

func (r *Resource) Delete(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	err = r.resourceSrv.Delete(c.Request().Context(), user.ID, id)
	if err != nil {
		return err
	}

	return response.JsonSuccess(c)
}

func (r *Resource) Start(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	err = r.resourceSrv.Start(c.Request().Context(), user.ID, id)
	if err != nil {
		return err
	}

	return response.JsonSuccess(c)
}

func (r *Resource) Stop(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	err = r.resourceSrv.Stop(c.Request().Context(), user.ID, id)
	if err != nil {
		return err
	}

	return response.JsonSuccess(c)
}

func (r *Resource) Restart(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	err = r.resourceSrv.Restart(c.Request().Context(), user.ID, id)
	if err != nil {
		return err
	}

	return response.JsonSuccess(c)
}
