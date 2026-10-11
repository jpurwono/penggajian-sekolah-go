package inventory

import (
	"github.com/labstack/echo/v4"

	"penggajian/server/internal/platform/httpx"
)

type MutationHandler struct {
	svc  *MutationService
	repo *Repository
}

func NewMutationHandler(svc *MutationService, repo *Repository) *MutationHandler {
	return &MutationHandler{svc: svc, repo: repo}
}

func (h *MutationHandler) List(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)

	var itemID *int64
	if v := c.QueryParam("item_id"); v != "" {
		if n := queryInt64(c, "item_id"); n != nil {
			itemID = n
		}
	}

	data, total, err := h.repo.ListMutations(
		c.Request().Context(), q, c.QueryParam("type"), itemID,
		c.QueryParam("date_from"), c.QueryParam("date_to"), per, (page-1)*per,
	)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, data, httpx.NewMeta(page, per, total))
}

func (h *MutationHandler) Create(c echo.Context) error {
	var req CreateMutationRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.Create(c.Request().Context(), c, req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.Created(c, resp)
}

type StockHandler struct {
	repo *Repository
}

func NewStockHandler(repo *Repository) *StockHandler {
	return &StockHandler{repo: repo}
}

func (h *StockHandler) List(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)

	data, total, err := h.repo.ListStocks(
		c.Request().Context(), q,
		queryInt64(c, "item_id"), queryInt64(c, "location_id"),
		c.QueryParam("condition"), per, (page-1)*per,
	)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, data, httpx.NewMeta(page, per, total))
}

func (h *StockHandler) Summary(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)
	belowMin := c.QueryParam("below_min") == "true"

	data, total, err := h.repo.StockSummary(c.Request().Context(), q, belowMin, per, (page-1)*per)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, data, httpx.NewMeta(page, per, total))
}

type PurchaseHandler struct {
	svc  *PurchaseService
	repo *Repository
}

func NewPurchaseHandler(svc *PurchaseService, repo *Repository) *PurchaseHandler {
	return &PurchaseHandler{svc: svc, repo: repo}
}

func (h *PurchaseHandler) List(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)
	rows, total, err := h.svc.List(c.Request().Context(), q, c.QueryParam("status"), per, (page-1)*per)
	if err != nil {
		return mapError(c, err)
	}
	resp := make([]*PurchaseResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toPurchaseResponse(row))
	}
	return httpx.OKWithMeta(c, resp, httpx.NewMeta(page, per, total))
}

func (h *PurchaseHandler) Get(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	purchase, items, err := h.svc.Find(c.Request().Context(), id)
	if err != nil {
		return mapError(c, mapNotFound(err))
	}
	return httpx.OK(c, toPurchaseDetail(purchase, items))
}

func (h *PurchaseHandler) Create(c echo.Context) error {
	var req SavePurchaseRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	id, err := h.svc.Create(c.Request().Context(), c, req)
	if err != nil {
		return mapError(c, err)
	}
	purchase, items, err := h.svc.Find(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.Created(c, toPurchaseDetail(purchase, items))
}

func (h *PurchaseHandler) Update(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SavePurchaseRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	if err := h.svc.Update(c.Request().Context(), c, id, req); err != nil {
		return mapError(c, err)
	}
	purchase, items, err := h.svc.Find(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, toPurchaseDetail(purchase, items))
}

func (h *PurchaseHandler) Delete(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMessage(c, "Pengadaan berhasil dihapus", nil)
}

func (h *PurchaseHandler) Receive(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.svc.Receive(c.Request().Context(), c, id); err != nil {
		return mapError(c, err)
	}
	purchase, items, err := h.svc.Find(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMessage(c, "Pengadaan diterima — stok bertambah", toPurchaseDetail(purchase, items))
}
