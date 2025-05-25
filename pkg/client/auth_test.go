package client

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken(t *testing.T) {
	client := &Client{
		username: "testuser",
		password: "testpassword",
	}

	token, err := client.GenerateToken()
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Parse and validate the token
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(client.password), nil
	})

	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if !parsedToken.Valid {
		t.Fatal("Token is not valid")
	}

	claims, ok := parsedToken.Claims.(*JWTClaims)
	if !ok {
		t.Fatal("Failed to parse claims")
	}

	if claims.Username != client.username {
		t.Fatalf("Expected username %s, got %s", client.username, claims.Username)
	}

	// Check that the token expires in approximately 24 hours
	expectedExpiry := time.Now().Add(24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time
	timeDiff := actualExpiry.Sub(expectedExpiry)
	if timeDiff > time.Minute || timeDiff < -time.Minute {
		t.Fatalf("Token expiry time is not as expected. Expected around %v, got %v", expectedExpiry, actualExpiry)
	}
}

func TestSetAuthHeader(t *testing.T) {
	client := &Client{
		username: "testuser",
		password: "testpassword",
	}

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	err = client.SetAuthHeader(req)
	if err != nil {
		t.Fatalf("Failed to set auth header: %v", err)
	}

	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		t.Fatal("Authorization header is empty")
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		t.Fatalf("Authorization header should start with 'Bearer ', got: %s", authHeader)
	}

	// Extract and validate the token
	token := authHeader[7:]
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(client.password), nil
	})

	if err != nil {
		t.Fatalf("Failed to parse token from header: %v", err)
	}

	if !parsedToken.Valid {
		t.Fatal("Token from header is not valid")
	}

	claims, ok := parsedToken.Claims.(*JWTClaims)
	if !ok {
		t.Fatal("Failed to parse claims from header token")
	}

	if claims.Username != client.username {
		t.Fatalf("Expected username %s, got %s", client.username, claims.Username)
	}
}
