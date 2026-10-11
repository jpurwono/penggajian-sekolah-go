package rbac

import (
	"errors"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/httpx"
	"penggajian/server/internal/platform/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return httpx.ErrNotFound(c, err.Error())
	case errors.Is(err, ErrRoleInUse), errors.Is(err, ErrCannotEditSelf):
		return httpx.ErrBadRequest(c, err.Error())
	case errors.Is(err, gorm.ErrRecordNotFound):
		return httpx.ErrNotFound(c, "Data tidak ditemukan")
	default:
		return httpx.ErrBadRequest(c, err.Error())
	}
}

func (h *Handler) ListUsers(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)

	filter := ListUsersFilter{Page: page, PerPage: per, Q: q}
	if v := c.QueryParam("role_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.RoleID = &id
		}
	}
	if v := c.QueryParam("is_active"); v == "true" || v == "false" {
		b := v == "true"
		filter.Active = &b
	}

	result, err := h.svc.ListUsers(c.Request().Context(), filter)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, result.Users, result.Meta)
}

func (h *Handler) GetUser(c echo.Context) error {
	id, err := httpx.ParseID(c, "id")
	if err != nil {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	user, err := h.svc.GetUser(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, user)
}

func (h *Handler) CreateUser(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	user, err := h.svc.CreateUser(c.Request().Context(), req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.Created(c, user)
}

func (h *Handler) UpdateUser(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	user, err := h.svc.UpdateUser(c.Request().Context(), id, middleware.CurrentUserID(c), req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, user)
}

func (h *Handler) UpdateUserPassword(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req UpdatePasswordRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	if err := h.svc.UpdateUserPassword(c.Request().Context(), id, req.NewPassword); err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMessage(c, "Password berhasil diperbarui", nil)
}

func (h *Handler) DeleteUser(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.svc.DeleteUser(c.Request().Context(), id, middleware.CurrentUserID(c)); err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMessage(c, "User berhasil dihapus", nil)
}

func (h *Handler) ListRoles(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)
	result, err := h.svc.ListRoles(c.Request().Context(), ListRolesFilter{Page: page, PerPage: per, Q: q})
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, result.Roles, result.Meta)
}

func (h *Handler) GetRole(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	role, err := h.svc.GetRole(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, role)
}

func (h *Handler) CreateRole(c echo.Context) error {
	var req SaveRoleRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	role, err := h.svc.CreateRole(c.Request().Context(), req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.Created(c, role)
}

func (h *Handler) UpdateRole(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SaveRoleRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	role, err := h.svc.UpdateRole(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, role)
}

func (h *Handler) DeleteRole(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.svc.DeleteRole(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMessage(c, "Role berhasil dihapus", nil)
}

func (h *Handler) ListPermissions(c echo.Context) error {
	perms, err := h.svc.ListPermissions(c.Request().Context())
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, perms)
}
