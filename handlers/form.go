package handlers

import (
	"net/http"
	"time"

	"lite-collector/services"
	"lite-collector/utils"

	"github.com/gin-gonic/gin"
)

// CreateForm godoc
// @Summary      Create a form
// @Description  Creates a new draft form owned by the authenticated user.
// @Tags         forms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      createFormRequest  true  "Form data"
// @Success      201   {object}  formResponse
// @Failure      400   {object}  errorResponse
// @Failure      401   {object}  errorResponse
// @Failure      500   {object}  errorResponse
// @Router       /forms [post]
func CreateForm(formService *services.FormService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createFormRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			e := utils.ErrBadRequest
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: err.Error()}})
			return
		}

		userID := c.MustGet("user_id").(uint64)

		form, err := formService.CreateForm(userID, req.Title, req.Description, []byte(req.Schema))
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusCreated, formResponse{
			ID:          form.ID,
			Title:       form.Title,
			Description: form.Description,
			Status:      form.Status,
			CreatedAt:   form.CreatedAt,
			UpdatedAt:   form.UpdatedAt,
		})
	}
}

// GetForms godoc
// @Summary      List my forms
// @Description  Returns all forms owned by the authenticated user.
// @Tags         forms
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  formListResponse
// @Failure      401  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /forms [get]
func GetForms(formService *services.FormService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)

		forms, err := formService.GetFormsByOwner(userID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		items := make([]formResponse, 0, len(forms))
		for _, f := range forms {
			items = append(items, formResponse{
				ID:          f.ID,
				Title:       f.Title,
				Description: f.Description,
				Status:      f.Status,
				CreatedAt:   f.CreatedAt,
				UpdatedAt:   f.UpdatedAt,
			})
		}
		c.JSON(http.StatusOK, formListResponse{Forms: items})
	}
}

// GetForm godoc
// @Summary      Get a form
// @Description  Returns a single form by ID. Only the owner can access it.
// @Tags         forms
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int  true  "Form ID"
// @Success      200     {object}  formDetailResponse
// @Failure      401     {object}  errorResponse
// @Failure      403     {object}  errorResponse
// @Failure      404     {object}  errorResponse
// @Router       /forms/{formId} [get]
func GetForm(formService *services.FormService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		form, err := formService.GetFormByID(formID, userID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusOK, formDetailResponse{
			ID:          form.ID,
			Title:       form.Title,
			Description: form.Description,
			Schema:      string(form.Schema),
			Status:      form.Status,
			CreatedAt:   form.CreatedAt,
			UpdatedAt:   form.UpdatedAt,
		})
	}
}

// UpdateForm godoc
// @Summary      Update a form
// @Description  Updates title, description, and schema of a draft form. Owner only.
// @Tags         forms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int               true  "Form ID"
// @Param        body    body      updateFormRequest  true  "Updated form data"
// @Success      200     {object}  formResponse
// @Failure      400     {object}  errorResponse
// @Failure      401     {object}  errorResponse
// @Failure      403     {object}  errorResponse
// @Failure      404     {object}  errorResponse
// @Failure      500     {object}  errorResponse
// @Router       /forms/{formId} [put]
func UpdateForm(formService *services.FormService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		var req updateFormRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			e := utils.ErrBadRequest
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: err.Error()}})
			return
		}

		form, err := formService.UpdateForm(formID, userID, req.Title, req.Description, []byte(req.Schema))
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusOK, formResponse{
			ID:          form.ID,
			Title:       form.Title,
			Description: form.Description,
			Status:      form.Status,
			CreatedAt:   form.CreatedAt,
			UpdatedAt:   form.UpdatedAt,
		})
	}
}

// PublishForm godoc
// @Summary      Publish a form
// @Description  Changes form status from draft to published. Submitters can then fill it in. Owner only.
// @Tags         forms
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int  true  "Form ID"
// @Success      200     {object}  messageResponse
// @Failure      401     {object}  errorResponse
// @Failure      403     {object}  errorResponse
// @Failure      404     {object}  errorResponse
// @Failure      500     {object}  errorResponse
// @Router       /forms/{formId}/publish [post]
func PublishForm(formService *services.FormService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		if err := formService.PublishForm(formID, userID); err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusOK, messageResponse{Message: "form published successfully"})
	}
}

// Request / response types

type createFormRequest struct {
	Title       string `json:"title"       binding:"required" example:"2024 Annual Report"`
	Description string `json:"description"                    example:"Please fill in before Friday"`
	Schema      string `json:"schema"      binding:"required" example:"{\"fields\":[{\"key\":\"f_001\",\"label\":\"姓名\",\"type\":\"text\",\"required\":true}]}"`
}

type updateFormRequest struct {
	Title       string `json:"title"       example:"Updated title"`
	Description string `json:"description" example:"Updated description"`
	Schema      string `json:"schema"      example:"{\"fields\":[]}"`
}

type formResponse struct {
	ID          uint64    `json:"id"          example:"42"`
	Title       string    `json:"title"       example:"2024 Annual Report"`
	Description string    `json:"description" example:"Please fill in before Friday"`
	Status      int8      `json:"status"      example:"0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type formDetailResponse struct {
	ID          uint64    `json:"id"          example:"42"`
	Title       string    `json:"title"       example:"2024 Annual Report"`
	Description string    `json:"description" example:"Please fill in before Friday"`
	Schema      string    `json:"schema"      example:"{\"fields\":[]}"`
	Status      int8      `json:"status"      example:"1"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type formListResponse struct {
	Forms []formResponse `json:"forms"`
}

type messageResponse struct {
	Message string `json:"message" example:"form published successfully"`
}
