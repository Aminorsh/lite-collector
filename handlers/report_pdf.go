package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"lite-collector/services"
	"lite-collector/utils"

	"github.com/gin-gonic/gin"
)

// GetReportPDF godoc
// @Summary      下载报告 PDF
// @Description  将指定的已完成报告任务（generate_report）的 markdown 输出渲染为 PDF 并下载。仅任务所属用户可访问。前端应使用 wx.downloadFile + wx.openDocument 打开。
// @Tags         AI任务
// @Produce      application/pdf
// @Security     BearerAuth
// @Param        jobId  path      int  true  "任务 ID"
// @Success      200    {file}    binary
// @Failure      400    {object}  errorResponse  "非 generate_report 类型任务"
// @Failure      401    {object}  errorResponse  "未登录或 token 已过期"
// @Failure      403    {object}  errorResponse  "无权访问该任务"
// @Failure      404    {object}  errorResponse  "任务不存在"
// @Failure      409    {object}  errorResponse  "任务尚未成功完成"
// @Failure      500    {object}  errorResponse  "PDF 渲染失败"
// @Failure      503    {object}  errorResponse  "当前服务未启用 PDF 渲染"
// @Router       /jobs/{jobId}/pdf [get]
func GetReportPDF(formService *services.FormService, aiJobService *services.AIJobService, pdfService *services.PDFService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)

		jobID, err := strconv.ParseUint(c.Param("jobId"), 10, 64)
		if err != nil {
			e := utils.ErrBadRequest
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: "invalid job id"}})
			return
		}

		job, err := aiJobService.GetJobStatus(jobID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}
		if job.UserID != userID {
			e := utils.ErrForbidden
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}
		if job.JobType != "generate_report" {
			e := utils.ErrJobNotReportable
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}
		if job.Status != 2 {
			e := utils.ErrJobNotCompleted
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		title := "数据分析报告"
		if job.FormID != nil {
			if form, err := formService.GetFormByID(strconv.FormatUint(*job.FormID, 10), userID); err == nil {
				title = strings.TrimSpace(form.Title) + " · 数据分析报告"
			}
		}

		pdfBytes, err := pdfService.Render(title, job.Output)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		filename := fmt.Sprintf("report-%d.pdf", jobID)
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
		c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))
		c.Data(http.StatusOK, "application/pdf", pdfBytes)
	}
}
