package entity

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for entities
type Handler struct {
	service Service
}

// NewHandler creates a new entity HTTP handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers entity routes onto a gin router group
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	group := rg.Group("/entities")
	{
		group.GET("", h.GetAll)
		group.GET("/:id", h.GetByID)
		group.POST("", h.Create)
		group.PUT("/:id", h.Update)
		group.PUT("", h.Update) // support query parameter: PUT /api/entities?id=...
		group.DELETE("/:id", h.Delete)
		group.DELETE("", h.Delete) // support query parameter: DELETE /api/entities?id=...
	}
}

// GetAll handles GET /api/entities (or GET /api/entities?id=...)
func (h *Handler) GetAll(c *gin.Context) {
	if idStr := c.Query("id"); idStr != "" {
		h.GetByID(c)
		return
	}

	entities, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to retrieve entities", nil))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Entities retrieved successfully", entities))
}

// GetByID handles GET /api/entities/:id (or /api/entities?id=...)
func (h *Handler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		idStr = c.Query("id")
	}

	entityID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid entity ID format", map[string]string{
			"id": "must be a valid UUID",
		}))
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), entityID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, NewErrorResponse("Entity not found", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to retrieve entity", nil))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Entity retrieved successfully", entity))
}

// Create handles POST /api/entities
func (h *Handler) Create(c *gin.Context) {
	var req CreateOrUpdateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid entity data", map[string]string{
			"body": "Malformed or invalid JSON payload: " + err.Error(),
		}))
		return
	}

	if validationErrors := req.Validate(); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid entity data", validationErrors))
		return
	}

	created, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to create entity", nil))
		return
	}

	c.JSON(http.StatusCreated, NewSuccessResponse("Entity created successfully", created))
}

// Update handles PUT /api/entities/:id (or /api/entities?id=...)
func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		idStr = c.Query("id")
	}

	entityID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid entity ID format", map[string]string{
			"id": "must be a valid UUID",
		}))
		return
	}

	var req CreateOrUpdateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid entity data", map[string]string{
			"body": "Malformed or invalid JSON payload: " + err.Error(),
		}))
		return
	}

	if validationErrors := req.Validate(); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid entity data", validationErrors))
		return
	}

	updated, err := h.service.Update(c.Request.Context(), entityID, &req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, NewErrorResponse("Entity not found", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to update entity", nil))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Entity updated successfully", updated))
}

// Delete handles DELETE /api/entities/:id (or /api/entities?id=...)
func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		idStr = c.Query("id")
	}

	entityID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid entity ID format", map[string]string{
			"id": "must be a valid UUID",
		}))
		return
	}

	err = h.service.Delete(c.Request.Context(), entityID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, NewErrorResponse("Entity not found", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to delete entity", nil))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Entity deleted successfully", nil))
}
