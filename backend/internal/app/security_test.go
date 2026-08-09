package app

import (
	"strings"
	"testing"
	"time"
)

func TestPasswordRoundTrip(t *testing.T) {
	hash, err := hashPassword("Demo12345*")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "pbkdf2_sha256$") {
		t.Fatalf("formato inesperado: %s", hash)
	}
	if !verifyPassword(hash, "Demo12345*") {
		t.Fatal("la clave correcta no fue validada")
	}
	if verifyPassword(hash, "incorrecta") {
		t.Fatal("una clave incorrecta fue aceptada")
	}
}

func TestAccessToken(t *testing.T) {
	tok, err := signAccessToken("secret-for-test", "user-1", "demo@example.com", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := parseAccessToken("secret-for-test", tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Sub != "user-1" || claims.Email != "demo@example.com" {
		t.Fatalf("claims incorrectos: %#v", claims)
	}
	if _, err := parseAccessToken("wrong-secret", tok); err == nil {
		t.Fatal("se acepto una firma incorrecta")
	}
}

func TestRefreshTokenHash(t *testing.T) {
	plain, hash, err := randomToken()
	if err != nil {
		t.Fatal(err)
	}
	if plain == "" || hash == "" {
		t.Fatal("token vacio")
	}
	if tokenHash(plain) != hash {
		t.Fatal("hash inconsistente")
	}
}
