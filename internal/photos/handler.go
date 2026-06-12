package photos

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mi-michi/backend/internal/cats"
	"github.com/mi-michi/backend/internal/middleware"
	"github.com/mi-michi/backend/internal/storage"
)

func HandleList(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	catID := c.Param("id")
	if ok := ensureCatOwner(c, catID, userID); !ok {
		return
	}
	photos, err := List(c.Request.Context(), userID, catID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, photos)
}

func HandleUpload(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	catID := c.Param("id")
	if ok := ensureCatOwner(c, catID, userID); !ok {
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

	caption := c.PostForm("caption")
	photo, err := Create(c.Request.Context(), userID, catID, photoKey, caption)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, photo)
}

func HandleDelete(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	if err := Delete(c.Request.Context(), userID, c.Param("photoId")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "eliminado"})
}

func ensureCatOwner(c *gin.Context, catID, userID string) bool {
	cat, err := cats.GetByID(c.Request.Context(), catID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return false
	}
	if cat == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "gato no encontrado"})
		return false
	}
	return true
}
