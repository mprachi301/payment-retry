package api

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine, h *JobHandler) {
	r.GET("/HEALTH", func(c *gin.Context) {
		c.JSON(200, gin.H{"Status": "OK"})
	})
	v1 := r.Group("/api/v1")
	{
		v1.POST("/jobs", h.CreateJob)
		v1.GET("/jobs/:id", h.GetJob)
	}
}
