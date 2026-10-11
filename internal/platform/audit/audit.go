// Package audit menyediakan pencatatan audit log lintas fitur (best-effort).
package audit

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/middleware"
)

func toJSON(v any) []byte {
	if v == nil {
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	if string(data) == "null" {
		return nil
	}
	return data
}

// Log mencatat aksi ke audit_logs; kegagalan tidak menggagalkan request.
func Log(c echo.Context, db *gorm.DB, action, entity string, entityID int64, oldValues, newValues any) {
	err := db.Exec(`
		INSERT INTO audit_logs (user_id, action, entity, entity_id, old_values, new_values, ip)
		VALUES (?, ?, ?, ?, ?::jsonb, ?::jsonb, ?)
	`, middleware.CurrentUserID(c), action, entity, entityID, toJSON(oldValues), toJSON(newValues), c.RealIP()).Error
	if err != nil {
		c.Logger().Warnf("gagal menulis audit log: %v", err)
	}
}
