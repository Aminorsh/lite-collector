package handlers

import (
	"net/http"

	"lite-collector/services"

	"github.com/gin-gonic/gin"
)

// CreateForm handles creating a new form.
// Request body: { "title": "...", "description": "...", "schema": "<json string>" }
func CreateForm(formService *services.FormService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Title       string `json:"title" binding:"required"`
			Description string `json:"description"`
			Schema      string `json:"schema" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		userID := c.MustGet("user_id").(uint64)

		form, err := formService.CreateForm(userID, req.Title, req.Description, []byte(req.Schema))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create form"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":          form.ID,
			"title":       form.Title,
			"description": form.Description,
			"status":      form.Status,
			"created_at":  form.CreatedAt,
		})
	}
}

// GetForms returns all forms owned by the current user.
func GetForms(formService *services.FormService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)

		forms, err := formService.GetFormsByOwner(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get forms"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"forms": forms})
	}
}

// GetForm returns a single form by ID (owner only).
func GetForm(formService *services.FormService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		form, err := formService.GetFormByID(formID, userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":          form.ID,
			"title":       form.Title,
			"description": form.Description,
			"schema":      string(form.Schema),
			"status":      form.Status,
			"created_at":  form.CreatedAt,
			"updated_at":  form.UpdatedAt,
		})
	}
}

// UpdateForm updates an existing form (owner only).
// Request body: { "title": "...", "description": "...", "schema": "<json string>" }
func UpdateForm(formService *services.FormService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Schema      string `json:"schema"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		form, err := formService.UpdateForm(formID, userID, req.Title, req.Description, []byte(req.Schema))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update form"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":          form.ID,
			"title":       form.Title,
			"description": form.Description,
			"status":      form.Status,
			"updated_at":  form.UpdatedAt,
		})
	}
}

// PublishForm publishes a form (owner only).
func PublishForm(formService *services.FormService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		if err := formService.PublishForm(formID, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish form"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "form published successfully"})
	}
}
