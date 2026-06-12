package cats

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mi-michi/backend/internal/middleware"
	"github.com/mi-michi/backend/internal/storage"
)

func HandleList(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	cats, err := List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cats)
}

func HandleCreate(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	var req CreateCatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat, err := Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

func HandleGet(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	cat, err := GetByID(c.Request.Context(), c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if cat == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "gato no encontrado"})
		return
	}
	c.JSON(http.StatusOK, cat)
}

func HandleUpdate(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	var req UpdateCatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat, err := Update(c.Request.Context(), c.Param("id"), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if cat == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "gato no encontrado"})
		return
	}
	c.JSON(http.StatusOK, cat)
}

func HandleDelete(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	if err := Delete(c.Request.Context(), c.Param("id"), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "eliminado"})
}

// HandleUploadPhoto recibe multipart/form-data con campo "photo".
func HandleUploadPhoto(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	catID := c.Param("id")

	// Verificar que el gato pertenece al usuario
	cat, err := GetByID(c.Request.Context(), catID, userID)
	if err != nil || cat == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "gato no encontrado"})
		return
	}

	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'photo' requerido"})
		return
	}
	defer file.Close()

	photoKey, err := storage.UploadCatPhoto(c.Request.Context(), userID, catID, header.Filename, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := UpdatePhoto(c.Request.Context(), catID, userID, photoKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}
