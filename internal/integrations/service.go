package integrations

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ErrPetfinderNotConfigured = errors.New("Petfinder no configurado")

func ListCatBreeds(ctx context.Context) ([]CatBreed, error) {
	var breeds []CatBreed
	err := cacheJSON(ctx, "thecatapi:breeds:v1", 24*time.Hour, &breeds, func() (any, error) {
		var fetched []CatBreed
		if err := getJSON(ctx, theCatAPIURL("/breeds", nil), theCatAPIHeaders(), &fetched); err != nil {
			return nil, err
		}
		return fetched, nil
	})
	return breeds, err
}

func GetCatBreedImages(ctx context.Context, breedID string, limit int) ([]CatImage, error) {
	breedID = strings.TrimSpace(breedID)
	if breedID == "" {
		return nil, errors.New("breed_id requerido")
	}
	if limit < 1 || limit > 20 {
		limit = 8
	}

	var images []CatImage
	cacheKey := fmt.Sprintf("thecatapi:breed-images:%s:%d", breedID, limit)
	err := cacheJSON(ctx, cacheKey, 12*time.Hour, &images, func() (any, error) {
		params := url.Values{}
		params.Set("breed_ids", breedID)
		params.Set("limit", strconv.Itoa(limit))
		params.Set("size", "med")
		params.Set("mime_types", "jpg,png")
		params.Set("format", "json")

		var fetched []CatImage
		if err := getJSON(ctx, theCatAPIURL("/images/search", params), theCatAPIHeaders(), &fetched); err != nil {
			return nil, err
		}
		return fetched, nil
	})
	return images, err
}

func GetDailyCatFact(ctx context.Context) (*CatFact, error) {
	var fact CatFact
	key := "catfact:daily:" + time.Now().Format("2006-01-02")
	err := cacheJSON(ctx, key, 24*time.Hour, &fact, func() (any, error) {
		var fetched CatFact
		if err := getJSON(ctx, "https://catfact.ninja/fact", nil, &fetched); err != nil {
			return nil, err
		}
		return fetched, nil
	})
	if err != nil {
		return nil, err
	}
	return &fact, nil
}

func GetRandomCatImage(ctx context.Context, tag string) (*CataasImage, error) {
	var image CataasImage
	params := url.Values{}
	params.Set("json", "true")
	if tag = strings.TrimSpace(tag); tag != "" {
		params.Set("tags", tag)
	}
	err := cacheJSON(ctx, "cataas:random:"+tag, 10*time.Minute, &image, func() (any, error) {
		var fetched CataasImage
		if err := getJSON(ctx, "https://cataas.com/cat?"+params.Encode(), nil, &fetched); err != nil {
			return nil, err
		}
		if fetched.URL != "" && strings.HasPrefix(fetched.URL, "/") {
			fetched.URL = "https://cataas.com" + fetched.URL
		}
		return fetched, nil
	})
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func GetFoodProduct(ctx context.Context, barcode string) (*FoodProduct, error) {
	barcode = strings.TrimSpace(barcode)
	if barcode == "" {
		return nil, errors.New("barcode requerido")
	}

	var product FoodProduct
	cacheKey := "openpetfoodfacts:product:" + barcode
	err := cacheJSON(ctx, cacheKey, 12*time.Hour, &product, func() (any, error) {
		var raw struct {
			Code    string         `json:"code"`
			Status  any            `json:"status"`
			Product map[string]any `json:"product"`
		}
		endpoint := fmt.Sprintf("%s/api/v3/product/%s.json", openPetFoodFactsBaseURL(), url.PathEscape(barcode))
		if err := getJSON(ctx, endpoint, openFoodFactsHeaders(), &raw); err != nil {
			return nil, err
		}
		if len(raw.Product) == 0 || isZeroStatus(raw.Status) {
			return nil, fmt.Errorf("producto no encontrado")
		}
		return foodProductFromMap(barcode, raw.Product), nil
	})
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func SearchFoodProducts(ctx context.Context, query string, limit int) ([]FoodProduct, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []FoodProduct{}, nil
	}
	if limit < 1 || limit > 20 {
		limit = 10
	}

	var products []FoodProduct
	cacheKey := fmt.Sprintf("openpetfoodfacts:search:%s:%d", strings.ToLower(query), limit)
	err := cacheJSON(ctx, cacheKey, 30*time.Minute, &products, func() (any, error) {
		params := url.Values{}
		params.Set("search_terms", query)
		params.Set("search_simple", "1")
		params.Set("action", "process")
		params.Set("json", "1")
		params.Set("page_size", strconv.Itoa(limit))

		var raw struct {
			Products []map[string]any `json:"products"`
		}
		if err := getJSON(ctx, openPetFoodFactsBaseURL()+"/cgi/search.pl?"+params.Encode(), openFoodFactsHeaders(), &raw); err != nil {
			return nil, err
		}

		results := make([]FoodProduct, 0, len(raw.Products))
		for _, item := range raw.Products {
			barcode := stringFromMap(item, "code", "_id")
			results = append(results, foodProductFromMap(barcode, item))
		}
		return results, nil
	})
	return products, err
}

func SearchAdoptableCats(ctx context.Context, location, breed string, limit int) ([]AdoptionAnimal, error) {
	if limit < 1 || limit > 20 {
		limit = 10
	}

	token, err := petfinderBearerToken(ctx)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("type", "Cat")
	params.Set("limit", strconv.Itoa(limit))
	if location = strings.TrimSpace(location); location != "" {
		params.Set("location", location)
	}
	if breed = strings.TrimSpace(breed); breed != "" {
		params.Set("breed", breed)
	}

	var raw struct {
		Animals []struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Age         string `json:"age"`
			Gender      string `json:"gender"`
			Size        string `json:"size"`
			Status      string `json:"status"`
			Description string `json:"description"`
			URL         string `json:"url"`
			Breeds      struct {
				Primary   string `json:"primary"`
				Secondary string `json:"secondary"`
				Mixed     bool   `json:"mixed"`
			} `json:"breeds"`
			Photos []struct {
				Small  string `json:"small"`
				Medium string `json:"medium"`
				Large  string `json:"large"`
				Full   string `json:"full"`
			} `json:"photos"`
			Contact struct {
				Email string `json:"email"`
				Phone string `json:"phone"`
			} `json:"contact"`
		} `json:"animals"`
	}

	err = getJSON(ctx, "https://api.petfinder.com/v2/animals?"+params.Encode(), map[string]string{
		"Authorization": "Bearer " + token,
	}, &raw)
	if err != nil {
		return nil, err
	}

	animals := make([]AdoptionAnimal, 0, len(raw.Animals))
	for _, item := range raw.Animals {
		breeds := []string{}
		if item.Breeds.Primary != "" {
			breeds = append(breeds, item.Breeds.Primary)
		}
		if item.Breeds.Secondary != "" {
			breeds = append(breeds, item.Breeds.Secondary)
		}
		if item.Breeds.Mixed {
			breeds = append(breeds, "Mix")
		}

		photos := []string{}
		for _, photo := range item.Photos {
			switch {
			case photo.Medium != "":
				photos = append(photos, photo.Medium)
			case photo.Large != "":
				photos = append(photos, photo.Large)
			case photo.Full != "":
				photos = append(photos, photo.Full)
			case photo.Small != "":
				photos = append(photos, photo.Small)
			}
		}

		contact := item.Contact.Email
		if contact == "" {
			contact = item.Contact.Phone
		}

		animals = append(animals, AdoptionAnimal{
			ID:          item.ID,
			Name:        item.Name,
			Age:         item.Age,
			Gender:      item.Gender,
			Size:        item.Size,
			Status:      item.Status,
			Description: item.Description,
			URL:         item.URL,
			Breeds:      breeds,
			Photos:      photos,
			Contact:     contact,
		})
	}

	return animals, nil
}

func theCatAPIURL(path string, params url.Values) string {
	base := strings.TrimRight(firstEnv("THE_CAT_API_BASE_URL"), "/")
	if base == "" {
		base = "https://api.thecatapi.com/v1"
	}
	endpoint := base + path
	if params != nil && len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	return endpoint
}

func theCatAPIHeaders() map[string]string {
	return map[string]string{
		"x-api-key": os.Getenv("THE_CAT_API_KEY"),
	}
}

func openFoodFactsHeaders() map[string]string {
	userAgent := os.Getenv("OPEN_PET_FOOD_FACTS_USER_AGENT")
	if userAgent == "" {
		userAgent = defaultUserAgent()
	}
	return map[string]string{"User-Agent": userAgent}
}

func openPetFoodFactsBaseURL() string {
	base := strings.TrimRight(os.Getenv("OPEN_PET_FOOD_FACTS_BASE_URL"), "/")
	if base == "" {
		base = "https://world.openpetfoodfacts.org"
	}
	return base
}

func isZeroStatus(value any) bool {
	switch v := value.(type) {
	case float64:
		return v == 0
	case int:
		return v == 0
	case string:
		return v == "0"
	default:
		return false
	}
}

func foodProductFromMap(barcode string, product map[string]any) FoodProduct {
	nutriments := map[string]any{}
	if raw, ok := product["nutriments"].(map[string]any); ok {
		nutriments = raw
	}
	if barcode == "" {
		barcode = stringFromMap(product, "code", "_id")
	}
	nova := stringFromMap(product, "nova_group")
	if nova == "" {
		nova = stringFromMap(product, "nova_groups")
	}

	return FoodProduct{
		Barcode:         barcode,
		Name:            stringFromMap(product, "product_name", "product_name_es", "product_name_en"),
		GenericName:     stringFromMap(product, "generic_name", "generic_name_es", "generic_name_en"),
		Brands:          stringFromMap(product, "brands"),
		Quantity:        stringFromMap(product, "quantity"),
		Categories:      stringFromMap(product, "categories"),
		IngredientsText: stringFromMap(product, "ingredients_text", "ingredients_text_es", "ingredients_text_en"),
		ImageURL:        stringFromMap(product, "image_front_url", "image_url"),
		NutriScore:      stringFromMap(product, "nutriscore_grade", "nutrition_grade_fr"),
		NovaGroup:       nova,
		Nutriments:      nutriments,
		SourceURL:       fmt.Sprintf("%s/product/%s", openPetFoodFactsBaseURL(), barcode),
	}
}

func stringFromMap(values map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := values[key]
		if !ok || value == nil {
			continue
		}
		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		case float64:
			if v == float64(int64(v)) {
				return strconv.FormatInt(int64(v), 10)
			}
			return strconv.FormatFloat(v, 'f', -1, 64)
		case int:
			return strconv.Itoa(v)
		case []any:
			parts := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
					parts = append(parts, strings.TrimSpace(s))
				}
			}
			if len(parts) > 0 {
				return strings.Join(parts, ", ")
			}
		}
	}
	return ""
}

type petfinderTokenCache struct {
	token     string
	expiresAt time.Time
}

var (
	petfinderMu    sync.Mutex
	petfinderToken petfinderTokenCache
)

func petfinderBearerToken(ctx context.Context) (string, error) {
	clientID := os.Getenv("PETFINDER_CLIENT_ID")
	clientSecret := os.Getenv("PETFINDER_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", ErrPetfinderNotConfigured
	}

	petfinderMu.Lock()
	defer petfinderMu.Unlock()

	if petfinderToken.token != "" && time.Now().Before(petfinderToken.expiresAt.Add(-1*time.Minute)) {
		return petfinderToken.token, nil
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	var raw struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := postFormJSON(ctx, "https://api.petfinder.com/v2/oauth2/token", form, nil, &raw); err != nil {
		return "", err
	}
	if raw.AccessToken == "" {
		return "", errors.New("Petfinder no devolvió access_token")
	}
	expiresIn := raw.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}
	petfinderToken = petfinderTokenCache{
		token:     raw.AccessToken,
		expiresAt: time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
	return petfinderToken.token, nil
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}
