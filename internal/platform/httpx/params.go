package httpx

import (
	"math"
	"strconv"

	"github.com/labstack/echo/v4"
)

// ParseListParams membaca query param paginasi & pencarian standar:
// ?page=1&per_page=15&q=kata-kunci
func ParseListParams(c echo.Context) (page, per int64, q string) {
	page = parseIntDefault(c.QueryParam("page"), 1)
	if page < 1 {
		page = 1
	}

	per = parseIntDefault(c.QueryParam("per_page"), 15)
	if per < 1 {
		per = 15
	}
	if per > 100 {
		per = 100
	}

	return page, per, c.QueryParam("q")
}

func NewMeta(page, per, total int64) *Meta {
	totalPage := int64(0)
	if per > 0 {
		totalPage = int64(math.Ceil(float64(total) / float64(per)))
	}
	return &Meta{Page: page, PerPage: per, Total: total, TotalPage: totalPage}
}

func ParseID(c echo.Context, param string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(param), 10, 64)
	if err != nil || id <= 0 {
		return 0, echo.NewHTTPError(400, "ID tidak valid")
	}
	return id, nil
}

func parseIntDefault(value string, fallback int64) int64 {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
