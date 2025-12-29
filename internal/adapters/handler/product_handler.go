package handler

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/samsirama/go-wms-saas/internal/core/domain"
	"github.com/samsirama/go-wms-saas/internal/core/services"
)

type ProductHandler struct {
	svc      *services.ProductService
	validate *validator.Validate
}

func NewProductHandler(svc *services.ProductService) *ProductHandler {
	return &ProductHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

type CreateProductRequest struct {
	SKU         string `json:"sku" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type ListProductsQuery struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

func (h *ProductHandler) Create(c *fiber.Ctx) error {
	var req CreateProductRequest
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

	product := &domain.Product{
		SKU:         req.SKU,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.svc.Create(c.Context(), product); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create product",
			"data":    nil,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Product created successfully",
		"data":    product,
	})
}

func (h *ProductHandler) List(c *fiber.Ctx) error {
	var query ListProductsQuery
	if err := c.QueryParser(&query); err != nil {
		query.Limit = 20
		query.Offset = 0
	}

	if query.Limit == 0 {
		query.Limit = 20
	}

	products, err := h.svc.List(c.Context(), query.Limit, query.Offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch products",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Products retrieved successfully",
		"data":    products,
	})
}

func (h *ProductHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Product ID is required",
			"data":    nil,
		})
	}

	product, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Product not found",
				"data":    nil,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch product",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Product retrieved successfully",
		"data":    product,
	})
}
