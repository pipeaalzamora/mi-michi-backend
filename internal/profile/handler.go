package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mi-michi/backend/internal/middleware"
	"github.com/mi-michi/backend/pkg/firebaseauth"
	"github.com/mi-michi/backend/pkg/googleauth"
)

// GoogleLoginRequest recibe el id_token de Google desde Flutter móvil.
type GoogleLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type FirebaseLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

// GoogleDesktopLoginRequest recibe el authorization code del flujo desktop.
type GoogleDesktopLoginRequest struct {
	Code        string `json:"code"         binding:"required"`
	RedirectURI string `json:"redirect_uri" binding:"required"`
	ClientID    string `json:"client_id"    binding:"required"`
}

// HandleGoogleLogin — flujo móvil: verifica id_token directamente.
func HandleGoogleLogin(c *gin.Context) {
	var req GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload, err := googleauth.Verify(c.Request.Context(), req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token de Google inválido: " + err.Error()})
		return
	}

	user, err := UpsertByGoogle(c.Request.Context(), payload.GoogleID, payload.Email, payload.Name, payload.Picture)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al guardar usuario"})
		return
	}

	tokenStr, err := generateJWT(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al generar token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenStr, "user": user})
}

func HandleFirebaseLogin(c *gin.Context) {
	var req FirebaseLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload, err := firebaseauth.Verify(c.Request.Context(), req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token Firebase invalido"})
		return
	}

	user, err := UpsertByFirebase(c.Request.Context(), payload.UID, payload.Email, payload.Name, payload.Picture)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al guardar usuario"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// HandleGoogleDesktopLogin — flujo desktop: intercambia authorization code por tokens.
func HandleGoogleDesktopLogin(c *gin.Context) {
	var req GoogleDesktopLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Intercambiar code por tokens con Google
	idToken, err := exchangeCodeForIDToken(c.Request.Context(), req.Code, req.RedirectURI, req.ClientID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "error al intercambiar code: " + err.Error()})
		return
	}

	payload, err := googleauth.Verify(c.Request.Context(), idToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "id_token inválido: " + err.Error()})
		return
	}

	user, err := UpsertByGoogle(c.Request.Context(), payload.GoogleID, payload.Email, payload.Name, payload.Picture)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al guardar usuario"})
		return
	}

	tokenStr, err := generateJWT(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al generar token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenStr, "user": user})
}

// HandleGetProfile devuelve el perfil del usuario autenticado.
func HandleGetProfile(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	user, err := GetByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// HandleUpdateProfile actualiza el display_name del usuario.
func HandleUpdateProfile(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := UpdateDisplayName(c.Request.Context(), userID, req.DisplayName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// ── Helpers ────────────────────────────────────────────────────────────────

func generateJWT(userID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	expiryHours := 720
	if h := os.Getenv("JWT_EXPIRY_HOURS"); h != "" {
		if v, err := strconv.Atoi(h); err == nil {
			expiryHours = v
		}
	}
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Duration(expiryHours) * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// exchangeCodeForIDToken intercambia un authorization code de Google por un id_token.
// Requiere el client_secret configurado en el backend (GOOGLE_CLIENT_SECRET).
func exchangeCodeForIDToken(ctx context.Context, code, redirectURI, clientID string) (string, error) {
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientSecret == "" {
		return "", fmt.Errorf("GOOGLE_CLIENT_SECRET no configurado")
	}

	form := url.Values{
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://oauth2.googleapis.com/token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google token error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if result.IDToken == "" {
		return "", fmt.Errorf("google no devolvió id_token")
	}
	return result.IDToken, nil
}
