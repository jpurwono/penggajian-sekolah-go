package spp

import (
	"errors"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/audit"
	"penggajian/server/internal/platform/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return httpx.ErrNotFound(c, ErrNotFound.Error())
	default:
		return httpx.ErrBadRequest(c, err.Error())
	}
}

func queryIntPtr(c echo.Context, key string) *int64 {
	v := c.QueryParam(key)
	if v == "" {
		return nil
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		return nil
	}
	return &n
}

func queryInt(c echo.Context, key string) int {
	n, err := strconv.Atoi(c.QueryParam(key))
	if err != nil {
		return 0
	}
	return n
}

// ---------- Siswa ----------

func (h *Handler) ListStudents(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)

	var active *bool
	if v := c.QueryParam("is_active"); v == "true" || v == "false" {
		b := v == "true"
		active = &b
	}

	result, err := h.svc.ListStudents(c.Request().Context(), q, c.QueryParam("class_name"), active, page, per)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, result.Students, result.Meta)
}

func (h *Handler) CreateStudent(c echo.Context) error {
	var req SaveStudentRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.CreateStudent(c.Request().Context(), req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.db, "create", "students", resp.ID, nil, resp)
	return httpx.Created(c, resp)
}

func (h *Handler) UpdateStudent(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SaveStudentRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.UpdateStudent(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.db, "update", "students", id, nil, resp)
	return httpx.OK(c, resp)
}

func (h *Handler) DeleteStudent(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.svc.DeleteStudent(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.db, "delete", "students", id, nil, nil)
	return httpx.OKWithMessage(c, "Siswa berhasil dihapus", nil)
}

// ---------- Tagihan ----------

func (h *Handler) ListBills(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)

	var active *bool
	_ = active
	filter := BillFilter{
		Q:           q,
		StudentID:   queryIntPtr(c, "student_id"),
		PeriodMonth: queryInt(c, "period_month"),
		PeriodYear:  queryInt(c, "period_year"),
		Status:      c.QueryParam("status"),
		ClassName:   c.QueryParam("class_name"),
	}

	bills, meta, billed, paid, err := h.svc.ListBills(c.Request().Context(), filter, page, per)
	if err != nil {
		return mapError(c, err)
	}
	return c.JSON(200, echo.Map{
		"success": true,
		"data":    bills.Bills,
		"meta":    meta,
		"totals": echo.Map{
			"billed":     billed,
			"collected":  paid,
			"outstanding": billed - paid,
		},
	})
}

func (h *Handler) GetBill(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	resp, err := h.svc.GetBill(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, resp)
}

func (h *Handler) CreateBill(c echo.Context) error {
	var req SaveBillRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.CreateBill(c.Request().Context(), c, req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.Created(c, resp)
}

func (h *Handler) UpdateBill(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SaveBillRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.UpdateBill(c.Request().Context(), c, id, req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, resp)
}

func (h *Handler) DeleteBill(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.svc.DeleteBill(c.Request().Context(), c, id); err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMessage(c, "Tagihan berhasil dihapus", nil)
}

func (h *Handler) GenerateBills(c echo.Context) error {
	var req GenerateBillsRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	result, err := h.svc.Generate(c.Request().Context(), c, req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMessage(c, result.Message, result)
}

func (h *Handler) ListBillPayments(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	rows, err := h.svc.ListBillPayments(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, rows)
}

// ---------- Pembayaran ----------

func (h *Handler) CreatePayment(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req CreatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.CreatePayment(c.Request().Context(), c, id, req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.Created(c, resp)
}

func (h *Handler) ListPayments(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)
	filter := PaymentFilter{
		Q:         q,
		StudentID: queryIntPtr(c, "student_id"),
		DateFrom:  c.QueryParam("date_from"),
		DateTo:    c.QueryParam("date_to"),
	}
	result, err := h.svc.ListPayments(c.Request().Context(), filter, page, per)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, result.Payments, result.Meta)
}

func (h *Handler) GetPayment(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	resp, err := h.svc.GetPayment(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, resp)
}

func (h *Handler) YearSummary(c echo.Context) error {
	year := queryInt(c, "year")
	if year == 0 {
		year = timeNowYear()
	}
	resp, err := h.svc.YearSummary(c.Request().Context(), year)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, resp)
}
