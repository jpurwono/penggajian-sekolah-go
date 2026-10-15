package payroll

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/audit"
	"penggajian/server/internal/platform/httpx"
)

type SlipHandler struct {
	svc *SlipService
}

func NewSlipHandler(svc *SlipService) *SlipHandler {
	return &SlipHandler{svc: svc}
}

func (h *SlipHandler) Generate(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	result, err := h.svc.Generate(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.db, "generate", "payroll_periods", id, nil, result)
	return httpx.OKWithMessage(c, result.Message, result)
}

func (h *SlipHandler) ListSlipsByPeriod(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	page, per, q := httpx.ParseListParams(c)
	result, err := h.svc.ListByPeriod(c.Request().Context(), id, q, page, per)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, result.Payslips, result.Meta)
}

func (h *SlipHandler) GetSlip(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	resp, err := h.svc.GetSlip(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return httpx.ErrNotFound(c, "Slip tidak ditemukan")
		}
		return mapError(c, err)
	}
	return httpx.OK(c, resp)
}

func (h *SlipHandler) UpdateSlipNote(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var body struct {
		Note *string `json:"note"`
	}
	if err := c.Bind(&body); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	if err := h.svc.UpdateSlipNote(c.Request().Context(), id, body.Note); err != nil {
		return mapError(c, err)
	}
	resp, err := h.svc.GetSlip(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.db, "update", "payslips", id, nil, resp)
	return httpx.OK(c, resp)
}

func (h *SlipHandler) AddItem(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SaveItemRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.AddItem(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.db, "create", "payslip_items", id, nil, req)
	return httpx.Created(c, resp)
}

func (h *SlipHandler) UpdateItem(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	itemID, _ := httpx.ParseID(c, "item_id")
	if id <= 0 || itemID <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req UpdateItemRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.UpdateItem(c.Request().Context(), id, itemID, req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.db, "update", "payslip_items", itemID, nil, req)
	return httpx.OK(c, resp)
}

func (h *SlipHandler) DeleteItem(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	itemID, _ := httpx.ParseID(c, "item_id")
	if id <= 0 || itemID <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	resp, err := h.svc.DeleteItem(c.Request().Context(), id, itemID)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.db, "delete", "payslip_items", itemID, nil, nil)
	return httpx.OK(c, resp)
}

func (h *SlipHandler) UpdateAttendance(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req BulkAttendanceRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	count, err := h.svc.UpdateAttendance(c.Request().Context(), id, req.Items)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.db, "update", "payroll_periods", id, nil, map[string]any{
		"action": "input-jam", "count": count,
	})
	return httpx.OKWithMessage(c, fmt.Sprintf("Data jam %d slip berhasil disimpan", count), nil)
}

func (h *SlipHandler) ImportAttendance(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	res, err := h.svc.ImportAttendance(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.db, "update", "payroll_periods", id, nil, map[string]any{
		"action": "import-attendance", "slip_count": res.SlipCount,
	})
	return httpx.OKWithMessage(c, res.Message, res)
}

func (h *SlipHandler) Recap(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	recap, err := h.svc.Recap(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}

	if c.QueryParam("format") == "csv" {
		return writeRecapCSV(c, recap)
	}
	return httpx.OK(c, recap)
}

func writeRecapCSV(c echo.Context, recap *RecapResponse) error {
	var buf bytes.Buffer
	buf.WriteRune('\ufeff')
	w := csv.NewWriter(&buf)

	header := []string{"No", "No. Slip", "No. Pegawai", "Nama Pegawai"}
	header = append(header, recap.Columns...)
	header = append(header, "Total Penerimaan", "Total Potongan", "Diterima Bersih")
	if err := w.Write(header); err != nil {
		return err
	}

	for i, row := range recap.Rows {
		record := []string{
			fmt.Sprintf("%d", i+1), row.SlipNo, row.EmployeeNo, row.EmployeeName,
		}
		for _, code := range recap.Columns {
			record = append(record, fmt.Sprintf("%d", row.Components[code]))
		}
		record = append(record,
			fmt.Sprintf("%d", row.TotalEarning),
			fmt.Sprintf("%d", row.TotalDeduction),
			fmt.Sprintf("%d", row.TakeHomePay),
		)
		if err := w.Write(record); err != nil {
			return err
		}
	}

	total := []string{"", "", "", "TOTAL"}
	for _, code := range recap.Columns {
		sum := int64(0)
		for _, row := range recap.Rows {
			sum += row.Components[code]
		}
		total = append(total, fmt.Sprintf("%d", sum))
	}
	total = append(total,
		fmt.Sprintf("%d", recap.GrandEarning),
		fmt.Sprintf("%d", recap.GrandDeduction),
		fmt.Sprintf("%d", recap.GrandTakeHome),
	)
	if err := w.Write(total); err != nil {
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	filename := fmt.Sprintf("rekap-gaji-%d-%02d.csv", recap.Period.PeriodYear, recap.Period.PeriodMonth)
	c.Response().Header().Set(echo.HeaderContentType, "text/csv; charset=utf-8")
	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="`+filename+`"`)
	return c.Blob(200, "text/csv; charset=utf-8", buf.Bytes())
}
