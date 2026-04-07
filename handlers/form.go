package handlers

import (
	"net/http"

	"lite-collector/services"
	"lite-collector/repository"

	"github.com/gin-gonic/gin"
)

// CreateForm handles creating a new form
func CreateForm(formRepo repository.FormRepository) gin.HandlerFunc {
	formService := services.NewFormService(formRepo)
	return func(c *gin.Context) {
		var req struct {
			Title       string `json:"title" binding:"required"`
			Description string `json:"description"`
			Schema      string `json:"schema" binding:"required"` // JSON string
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Create form
		form, err := formService.CreateForm(userID.(uint64), req.Title, req.Description, []byte(req.Schema))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create form"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":        form.ID,
			"title":     form.Title,
			"description": form.Description,
			"status":    form.Status,
			"created_at": form.CreatedAt,
		})
	}
}

// GetForms handles getting list of forms for the current user
func GetForms(formRepo repository.FormRepository) gin.HandlerFunc {
	formService := services.NewFormService(formRepo)
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Get forms
		forms, err := formService.GetFormsByOwner(userID.(uint64))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get forms"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"forms": forms,
		})
	}
}

// GetForm handles getting a specific form by ID
func GetForm(formRepo repository.FormRepository) gin.HandlerFunc {
	formService := services.NewFormService(formRepo)
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		formID := c.Param("formId")
		if formID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Form ID required"})
			return
		}

		// Get form
		form, err := formService.GetFormByID(formID, userID.(uint64))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":        form.ID,
			"title":     form.Title,
			"description": form.Description,
			"schema":    string(form.Schema),
			"status":    form.Status,
			"created_at": form.CreatedAt,
			"updated_at": form.UpdatedAt,
		})
	}
}

// UpdateForm handles updating an existing form
func UpdateForm(formRepo repository.FormRepository) gin.HandlerFunc {
	formService := services.NewFormService(formRepo)
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		formID := c.Param("formId")
		if formID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Form ID required"})
			return
		}

		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Schema      string `json:"schema"` // JSON string
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Update form
		form, err := formService.UpdateForm(formID, userID.(uint64), req.Title, req.Description, []byte(req.Schema))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update form"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":        form.ID,
			"title":     form.Title,
			"description": form.Description,
			"status":    form.Status,
			"updated_at": form.UpdatedAt,
		})
	}
}

// PublishForm handles publishing a form (changing status to published)
func PublishForm(formRepo repository.FormRepository) gin.HandlerFunc {
	formService := services.NewFormService(formRepo)
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		formID := c.Param("formId")
		if formID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Form ID required"})
			return
		}

		// Publish form
		err := formService.PublishForm(formID, userID.(uint64))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish form"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Form published successfully",
		})
	}
}