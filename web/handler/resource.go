package handler

import (
	"devcloud/web/handler/dto"
	"devcloud/web/handler/reqctx"
	"devcloud/web/handler/response"
	"devcloud/web/service"
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

	items := make([]dto.ResourceResponse, len(rs))
	for i, r := range rs {
		items[i] = dto.ResourceResponse{
			ID:       r.ID,
			Name:     r.Name,
			Image:    r.Image,
			Provider: r.Provider.String(),
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
		Image:    rc.Image,
		Provider: rc.Provider.String(),
		Status:   rc.Status.String(),
	})
}
