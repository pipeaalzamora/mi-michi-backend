package firebaseauth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const certsURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"

type Payload struct {
	UID     string
	Email   string
	Name    string
	Picture string
}

var certCache = struct {
	sync.RWMutex
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
}{}

func Verify(ctx context.Context, idToken string) (*Payload, error) {
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		return nil, errors.New("FIREBASE_PROJECT_ID no configurado")
	}

	claims := jwt.MapClaims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithAudience(projectID),
		jwt.WithIssuer("https://securetoken.google.com/"+projectID),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)

	token, err := parser.ParseWithClaims(idToken, claims, func(token *jwt.Token) (interface{}, error) {
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("token sin kid")
		}
		keys, err := getKeys(ctx)
		if err != nil {
			return nil, err
		}
		key := keys[kid]
		if key == nil {
			return nil, errors.New("kid de Firebase no reconocido")
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token Firebase invalido")
	}

	uid, err := claims.GetSubject()
	if err != nil || uid == "" {
		return nil, errors.New("token sin uid")
	}

	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	picture, _ := claims["picture"].(string)

	return &Payload{
		UID:     uid,
		Email:   email,
		Name:    name,
		Picture: picture,
	}, nil
}

func getKeys(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	now := time.Now()

	certCache.RLock()
	if certCache.keys != nil && now.Before(certCache.expiresAt) {
		keys := certCache.keys
		certCache.RUnlock()
		return keys, nil
	}
	certCache.RUnlock()

	certCache.Lock()
	defer certCache.Unlock()

	if certCache.keys != nil && now.Before(certCache.expiresAt) {
		return certCache.keys, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, certsURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("firebase certs status %d: %s", resp.StatusCode, string(body))
	}

	var raw map[string]string
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	keys := make(map[string]*rsa.PublicKey, len(raw))
	for kid, certPEM := range raw {
		block, _ := pem.Decode([]byte(certPEM))
		if block == nil {
			return nil, fmt.Errorf("certificado Firebase invalido para kid %s", kid)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("certificado Firebase sin RSA public key para kid %s", kid)
		}
		keys[kid] = publicKey
	}

	certCache.keys = keys
	certCache.expiresAt = cacheExpiry(resp.Header.Get("Cache-Control"), now)
	return certCache.keys, nil
}

func cacheExpiry(cacheControl string, now time.Time) time.Time {
	for _, part := range strings.Split(cacheControl, ",") {
		part = strings.TrimSpace(part)
		if !strings.HasPrefix(part, "max-age=") {
			continue
		}
		seconds, err := strconv.Atoi(strings.TrimPrefix(part, "max-age="))
		if err == nil && seconds > 0 {
			return now.Add(time.Duration(seconds) * time.Second)
		}
	}
	return now.Add(time.Hour)
}
