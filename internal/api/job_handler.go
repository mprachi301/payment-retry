package api

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/mprachi301/payment-retry/internal/service"
)

type JobHandler struct {
	Service *service.JobService
}

func NewJobHandler(service *service.JobService) *JobHandler {
	return &JobHandler{
		Service: service,
	}
}

type CreateJobRequest struct {
	PaymentID string  `json:"payment_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
	Currency  string  `json:"currency" binding:"required"`
}

func (h *JobHandler) CreateJob(c *gin.Context) {
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error ": err.Error()})
		return
	}
	job, err := h.Service.Create(req.PaymentID, req.Amount, req.Currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, job)
}

func (h *JobHandler) GetJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Job ID"})
		return
	}
	job, err := h.Service.GetJob(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "job not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}
