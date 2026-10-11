package inventory

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	mw "penggajian/server/internal/platform/middleware"
)

// Register mendaftarkan routes fitur inventory ke group /api/v1.
func Register(g *echo.Group, db *gorm.DB, authMW echo.MiddlewareFunc) {
	repo := NewRepository(db)

	master := NewMasterService(repo)
	mutation := NewMutationService(db, repo)
	purchase := NewPurchaseService(db, repo)

	h := NewHandler(master)
	mh := NewMutationHandler(mutation, repo)
	sh := NewStockHandler(repo)
	ph := NewPurchaseHandler(purchase, repo)

	categories := g.Group("/inventory/categories", authMW, mw.RequirePermission("inventory.read"))
	categories.GET("", h.ListCategories)
	categories.POST("", h.CreateCategory, mw.RequirePermission("inventory.create"))
	categories.PUT("/:id", h.UpdateCategory, mw.RequirePermission("inventory.update"))
	categories.DELETE("/:id", h.DeleteCategory, mw.RequirePermission("inventory.delete"))

	locations := g.Group("/inventory/locations", authMW, mw.RequirePermission("inventory.read"))
	locations.GET("", h.ListLocations)
	locations.POST("", h.CreateLocation, mw.RequirePermission("inventory.create"))
	locations.PUT("/:id", h.UpdateLocation, mw.RequirePermission("inventory.update"))
	locations.DELETE("/:id", h.DeleteLocation, mw.RequirePermission("inventory.delete"))

	items := g.Group("/inventory/items", authMW, mw.RequirePermission("inventory.read"))
	items.GET("", h.ListItems)
	items.POST("", h.CreateItem, mw.RequirePermission("inventory.create"))
	items.PUT("/:id", h.UpdateItem, mw.RequirePermission("inventory.update"))
	items.DELETE("/:id", h.DeleteItem, mw.RequirePermission("inventory.delete"))

	mutations := g.Group("/inventory/mutations", authMW, mw.RequirePermission("inventory.read"))
	mutations.GET("", mh.List)
	mutations.POST("", mh.Create, mw.RequirePermission("inventory.create"))

	stocks := g.Group("/inventory/stocks", authMW, mw.RequirePermission("inventory.read"))
	stocks.GET("", sh.List)
	stocks.GET("/summary", sh.Summary)

	purchases := g.Group("/inventory/purchases", authMW, mw.RequirePermission("inventory.read"))
	purchases.GET("", ph.List)
	purchases.GET("/:id", ph.Get)
	purchases.POST("", ph.Create, mw.RequirePermission("inventory.create"))
	purchases.PUT("/:id", ph.Update, mw.RequirePermission("inventory.update"))
	purchases.DELETE("/:id", ph.Delete, mw.RequirePermission("inventory.delete"))
	purchases.POST("/:id/receive", ph.Receive, mw.RequirePermission("inventory.update"))
}
