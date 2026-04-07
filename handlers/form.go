package handlers

import (
	"net/http"
	"time"

	"lite-collector/services"
	"lite-collector/utils"

	"github.com/gin-gonic/gin"
)

// CreateForm godoc
// @Summary      创建表单
// @Description  创建一个新的草稿表单，归属于当前登录用户。创建后需调用发布接口才能开放填写。
// @Tags         表单
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      createFormRequest  true  "表单基本信息"
// @Success      201   {object}  formResponse
// @Failure      400   {object}  errorResponse  "请求参数错误"
// @Failure      401   {object}  errorResponse  "未登录或 token 已过期"
// @Failure      500   {object}  errorResponse  "服务器内部错误"
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
// @Summary      获取我的表单列表
// @Description  返回当前登录用户创建的所有表单（草稿和已发布均包含）。
// @Tags         表单
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  formListResponse
// @Failure      401  {object}  errorResponse  "未登录或 token 已过期"
// @Failure      500  {object}  errorResponse  "服务器内部错误"
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
// @Summary      获取表单详情
// @Description  根据 ID 获取单个表单的完整信息，包含 schema 字段结构。仅表单创建者可访问。
// @Tags         表单
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int  true  "表单 ID"
// @Success      200     {object}  formDetailResponse
// @Failure      401     {object}  errorResponse  "未登录或 token 已过期"
// @Failure      403     {object}  errorResponse  "无权访问该表单"
// @Failure      404     {object}  errorResponse  "表单不存在"
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
// @Summary      更新表单
// @Description  修改草稿表单的标题、描述和字段结构（schema）。仅表单创建者可操作。
// @Tags         表单
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int               true  "表单 ID"
// @Param        body    body      updateFormRequest  true  "需要更新的字段"
// @Success      200     {object}  formResponse
// @Failure      400     {object}  errorResponse  "请求参数错误"
// @Failure      401     {object}  errorResponse  "未登录或 token 已过期"
// @Failure      403     {object}  errorResponse  "无权操作该表单"
// @Failure      404     {object}  errorResponse  "表单不存在"
// @Failure      500     {object}  errorResponse  "服务器内部错误"
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
// @Summary      发布表单
// @Description  将表单状态从草稿（0）改为已发布（1）。发布后填写人才可提交数据。仅表单创建者可操作。
// @Tags         表单
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int  true  "表单 ID"
// @Success      200     {object}  messageResponse
// @Failure      401     {object}  errorResponse  "未登录或 token 已过期"
// @Failure      403     {object}  errorResponse  "无权操作该表单"
// @Failure      404     {object}  errorResponse  "表单不存在"
// @Failure      500     {object}  errorResponse  "服务器内部错误"
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
	Title       string `json:"title"       binding:"required" example:"2024年度部门报表"`
	Description string `json:"description"                    example:"请于周五前填写完毕"`
	Schema      string `json:"schema"      binding:"required" example:"{\"fields\":[{\"key\":\"f_001\",\"label\":\"姓名\",\"type\":\"text\",\"required\":true}]}"`
}

type updateFormRequest struct {
	Title       string `json:"title"       example:"更新后的标题"`
	Description string `json:"description" example:"更新后的描述"`
	Schema      string `json:"schema"      example:"{\"fields\":[]}"`
}

type formResponse struct {
	ID          uint64    `json:"id"          example:"42"`
	Title       string    `json:"title"       example:"2024年度部门报表"`
	Description string    `json:"description" example:"请于周五前填写完毕"`
	Status      int8      `json:"status"      example:"0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type formDetailResponse struct {
	ID          uint64    `json:"id"          example:"42"`
	Title       string    `json:"title"       example:"2024年度部门报表"`
	Description string    `json:"description" example:"请于周五前填写完毕"`
	Schema      string    `json:"schema"      example:"{\"fields\":[{\"key\":\"f_001\",\"label\":\"姓名\",\"type\":\"text\",\"required\":true}]}"`
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
