package handler

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/samsirama/go-wms-saas/internal/core/domain"
	"github.com/samsirama/go-wms-saas/internal/core/services"
)

type StockHandler struct {
	svc      *services.StockService
	validate *validator.Validate
}

func NewStockHandler(svc *services.StockService) *StockHandler {
	return &StockHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

type ReserveStockRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int64  `json:"quantity" validate:"required,gt=0"`
	OrderID   string `json:"order_id" validate:"required"`
}

type ReleaseStockRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int64  `json:"quantity" validate:"required,gt=0"`
	OrderID   string `json:"order_id" validate:"required"`
}

func (h *StockHandler) ReserveStock(c *fiber.Ctx) error {
	var req ReserveStockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"data":    nil,
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
	}

	err := h.svc.ReserveStock(c.Context(), req.ProductID, req.Quantity, req.OrderID)
	if err != nil {
		if errors.Is(err, domain.ErrOptimisticLock) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "Stock was modified by another transaction, please retry",
				"data":    nil,
			})
		}
		if errors.Is(err, services.ErrInsufficientStock) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Insufficient stock available",
				"data":    nil,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to reserve stock",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Stock reserved successfully",
		"data": fiber.Map{
			"product_id": req.ProductID,
			"quantity":   req.Quantity,
			"order_id":   req.OrderID,
		},
	})
}

func (h *StockHandler) ReleaseStock(c *fiber.Ctx) error {
	var req ReleaseStockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"data":    nil,
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
	}

	err := h.svc.ReleaseStock(c.Context(), req.ProductID, req.Quantity, req.OrderID)
	if err != nil {
		if errors.Is(err, domain.ErrOptimisticLock) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "Stock was modified by another transaction, please retry",
				"data":    nil,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to release stock",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Stock released successfully",
		"data": fiber.Map{
			"product_id": req.ProductID,
			"quantity":   req.Quantity,
			"order_id":   req.OrderID,
		},
	})
}

func (h *StockHandler) GetStockLevel(c *fiber.Ctx) error {
	productID := c.Params("id")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Product ID is required",
			"data":    nil,
		})
	}

	stock, err := h.svc.GetStockLevel(c.Context(), productID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Stock not found for this product",
				"data":    nil,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch stock level",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Stock level retrieved successfully",
		"data":    stock,
	})
}

func (h *StockHandler) GetMutationHistory(c *fiber.Ctx) error {
	productID := c.Params("id")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Product ID is required",
			"data":    nil,
		})
	}

	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	mutations, err := h.svc.GetMutationHistory(c.Context(), productID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch mutation history",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Mutation history retrieved successfully",
		"data":    mutations,
	})
}
