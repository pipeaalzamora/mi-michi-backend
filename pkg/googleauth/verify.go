package googleauth

import (
	"context"
	"errors"
	"os"

	"google.golang.org/api/idtoken"
)

// Payload contiene los datos del usuario extraídos del token de Google.
type Payload struct {
	GoogleID string
	Email    string
	Name     string
	Picture  string
}

// Verify valida el id_token de Google y devuelve los datos del usuario.
func Verify(ctx context.Context, idToken string) (*Payload, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		return nil, errors.New("GOOGLE_CLIENT_ID no configurado")
	}

	payload, err := idtoken.Validate(ctx, idToken, clientID)
	if err != nil {
		return nil, err
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	if email == "" {
		return nil, errors.New("token sin email")
	}

	return &Payload{
		GoogleID: payload.Subject,
		Email:    email,
		Name:     name,
		Picture:  picture,
	}, nil
}
