package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mi-michi/backend/internal/middleware"
)

func HandleList(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	catID := c.Param("id")
	logs, err := List(c.Request.Context(), catID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func HandleCreate(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	catID := c.Param("id")
	var req CreateLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log, err := Create(c.Request.Context(), catID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, log)
}

func HandleUpdate(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	catID := c.Param("id")
	logID := c.Param("logId")
	var req UpdateLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log, err := Update(c.Request.Context(), logID, catID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if log == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "registro no encontrado"})
		return
	}
	c.JSON(http.StatusOK, log)
}

func HandleDelete(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	catID := c.Param("id")
	logID := c.Param("logId")
	if err := Delete(c.Request.Context(), logID, catID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "eliminado"})
}
