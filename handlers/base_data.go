package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"lite-collector/services"
	"lite-collector/utils"

	"github.com/gin-gonic/gin"
)

// BatchImportBaseData godoc
// @Summary      批量导入底表数据
// @Description  为指定表单批量导入底表数据（用于预填充）。如果 row_key 已存在则更新。仅表单创建者可操作。
// @Tags         底表数据
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int                      true  "表单 ID"
// @Param        body    body      batchImportRequest        true  "底表数据列表"
// @Success      200     {object}  batchImportResponse
// @Failure      400     {object}  errorResponse  "请求参数错误"
// @Failure      401     {object}  errorResponse  "未登录或 token 已过期"
// @Failure      403     {object}  errorResponse  "无权操作该表单"
// @Failure      404     {object}  errorResponse  "表单不存在"
// @Router       /forms/{formId}/base-data [post]
func BatchImportBaseData(formService *services.FormService, baseDataService *services.BaseDataService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		form, err := formService.GetFormByID(formID, userID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		var req batchImportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			e := utils.ErrBadRequest
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: err.Error()}})
			return
		}

		rows := make([]services.BaseDataRow, 0, len(req.Rows))
		for _, r := range req.Rows {
			dataBytes, err := json.Marshal(r.Data)
			if err != nil {
				continue
			}
			rows = append(rows, services.BaseDataRow{RowKey: r.RowKey, Data: dataBytes})
		}

		count, err := baseDataService.BatchImport(form.ID, rows)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusOK, batchImportResponse{Imported: count})
	}
}

// ListBaseData godoc
// @Summary      获取底表数据列表
// @Description  获取指定表单的所有底表数据。仅表单创建者可访问。
// @Tags         底表数据
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int  true  "表单 ID"
// @Success      200     {object}  baseDataListResponse
// @Failure      401     {object}  errorResponse  "未登录或 token 已过期"
// @Failure      403     {object}  errorResponse  "无权访问该表单"
// @Failure      404     {object}  errorResponse  "表单不存在"
// @Router       /forms/{formId}/base-data [get]
func ListBaseData(formService *services.FormService, baseDataService *services.BaseDataService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		form, err := formService.GetFormByID(formID, userID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		list, err := baseDataService.GetByFormID(form.ID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		items := make([]baseDataItem, 0, len(list))
		for _, bd := range list {
			items = append(items, baseDataItem{
				ID:     bd.ID,
				RowKey: bd.RowKey,
				Data:   json.RawMessage(bd.Data),
			})
		}
		c.JSON(http.StatusOK, baseDataListResponse{Rows: items})
	}
}

// LookupBaseData godoc
// @Summary      查询底表数据（预填充）
// @Description  根据 row_key 查询底表数据，用于填写表单时自动预填充字段。任何登录用户均可查询已发布表单的底表数据。
// @Tags         底表数据
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int     true  "表单 ID"
// @Param        row_key query     string  true  "查询键（如工号）"
// @Success      200     {object}  baseDataItem
// @Failure      401     {object}  errorResponse  "未登录或 token 已过期"
// @Failure      404     {object}  errorResponse  "未找到匹配的底表数据"
// @Router       /forms/{formId}/base-data/lookup [get]
func LookupBaseData(formService *services.FormService, baseDataService *services.BaseDataService) gin.HandlerFunc {
	return func(c *gin.Context) {
		formID := c.Param("formId")
		rowKey := c.Query("row_key")

		if rowKey == "" {
			e := utils.ErrBadRequest
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: "row_key is required"}})
			return
		}

		// Verify form is published (any auth'd user can lookup)
		form, err := formService.GetPublishedFormByID(formID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		bd, err := baseDataService.Lookup(form.ID, rowKey)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusOK, baseDataItem{
			ID:     bd.ID,
			RowKey: bd.RowKey,
			Data:   json.RawMessage(bd.Data),
		})
	}
}

// DeleteBaseData godoc
// @Summary      清空底表数据
// @Description  删除指定表单的所有底表数据。仅表单创建者可操作。
// @Tags         底表数据
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int  true  "表单 ID"
// @Success      200     {object}  messageResponse
// @Failure      401     {object}  errorResponse  "未登录或 token 已过期"
// @Failure      403     {object}  errorResponse  "无权操作该表单"
// @Failure      404     {object}  errorResponse  "表单不存在"
// @Router       /forms/{formId}/base-data [delete]
func DeleteBaseData(formService *services.FormService, baseDataService *services.BaseDataService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		form, err := formService.GetFormByID(formID, userID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		_ = form // ownership verified
		fid, _ := strconv.ParseUint(formID, 10, 64)
		if err := baseDataService.DeleteByFormID(fid); err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusOK, messageResponse{Message: "base data cleared"})
	}
}

// Request / response types

type batchImportRequest struct {
	Rows []batchImportRow `json:"rows" binding:"required"`
}

type batchImportRow struct {
	RowKey string                 `json:"row_key" binding:"required" example:"EMP001"`
	Data   map[string]interface{} `json:"data"    binding:"required"`
}

type batchImportResponse struct {
	Imported int `json:"imported" example:"10"`
}

type baseDataItem struct {
	ID     uint64          `json:"id"      example:"1"`
	RowKey string          `json:"row_key" example:"EMP001"`
	Data   json.RawMessage `json:"data"    swaggertype:"object"`
}

type baseDataListResponse struct {
	Rows []baseDataItem `json:"rows"`
}
