package vaccines

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mi-michi/backend/internal/middleware"
)

func HandleList(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	catID := c.Param("id")
	vaccines, err := List(c.Request.Context(), catID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vaccines)
}

func HandleCreate(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	catID := c.Param("id")
	var req CreateVaccineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := Create(c.Request.Context(), catID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

func HandleUpdate(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	catID := c.Param("id")
	vacID := c.Param("vacId")
	var req UpdateVaccineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := Update(c.Request.Context(), vacID, catID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if v == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vacuna no encontrada"})
		return
	}
	c.JSON(http.StatusOK, v)
}

func HandleDelete(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	catID := c.Param("id")
	vacID := c.Param("vacId")
	if err := Delete(c.Request.Context(), vacID, catID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "eliminada"})
}
