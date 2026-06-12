package integrations

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func HandleCatBreeds(c *gin.Context) {
	breeds, err := ListCatBreeds(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, breeds)
}

func HandleCatBreedImages(c *gin.Context) {
	images, err := GetCatBreedImages(c.Request.Context(), c.Param("id"), parseLimit(c.DefaultQuery("limit", "8")))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, images)
}

func HandleCatFact(c *gin.Context) {
	fact, err := GetDailyCatFact(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fact)
}

func HandleCatImage(c *gin.Context) {
	image, err := GetRandomCatImage(c.Request.Context(), c.Query("tag"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, image)
}

func HandleFoodProduct(c *gin.Context) {
	product, err := GetFoodProduct(c.Request.Context(), c.Param("barcode"))
	if err != nil {
		status := http.StatusBadGateway
		if err.Error() == "producto no encontrado" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, product)
}

func HandleFoodSearch(c *gin.Context) {
	products, err := SearchFoodProducts(c.Request.Context(), c.Query("q"), parseLimit(c.DefaultQuery("limit", "10")))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

func HandleAdoptions(c *gin.Context) {
	animals, err := SearchAdoptableCats(
		c.Request.Context(),
		c.Query("location"),
		c.Query("breed"),
		parseLimit(c.DefaultQuery("limit", "10")),
	)
	if err != nil {
		if errors.Is(err, ErrPetfinderNotConfigured) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Configura PETFINDER_CLIENT_ID y PETFINDER_CLIENT_SECRET para activar adopciones"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, animals)
}

func parseLimit(raw string) int {
	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return limit
}
