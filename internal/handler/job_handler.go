package handler

import (
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type JobHandler struct {
	jobService service.JobService
}

func NewJobHandler(jobService service.JobService) *JobHandler {
	return &JobHandler{jobService: jobService}
}

func (h *JobHandler) GetJob(c *gin.Context) {
	jobUUID := c.Param("job_uuid")
	if jobUUID == "" {
		response.Fail(c, errors.CodeInvalidParams, "Invalid job UUID")
		return
	}

	job, err := h.jobService.GetJobByUUID(jobUUID)
	if err != nil {
		response.Fail(c, errors.CodeJobNotFound, "Job not found")
		return
	}

	userID := getUserIDFromContext(c)
	if job.UserID != userID {
		response.Fail(c, errors.CodeJobAccessDenied, "Access denied")
		return
	}

	response.SuccessWithData(c, job.ToPublic())
}

func (h *JobHandler) CancelJob(c *gin.Context) {
	jobUUID := c.Param("job_uuid")
	if jobUUID == "" {
		response.Fail(c, errors.CodeInvalidParams, "Invalid job UUID")
		return
	}

	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "Unauthorized")
		return
	}

	job, err := h.jobService.CancelJob(userID, jobUUID)
	if err != nil {
		if err.Error() == "access denied" {
			response.Fail(c, errors.CodeJobAccessDenied, "Access denied")
			return
		}
		response.Fail(c, errors.CodeJobCancelFailed, "Failed to cancel job")
		return
	}

	response.SuccessWithData(c, job.ToPublic())
}
