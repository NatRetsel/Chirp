package internal

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenType string

const TokenAccess TokenType = "chirpy"

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("couldn't hash password: %v", err)
	}
	return string(hash), nil
}

func CheckPasswordHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})
	signedJWT, err := jwtToken.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("error signing secret token: %v", err)
	}
	return signedJWT, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimsStruct := jwt.RegisteredClaims{}
	jwtToken, err := jwt.ParseWithClaims(tokenString, &claimsStruct, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("error parsing token string: %v", err)
	}
	idString, err := jwtToken.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("error retrieving id from token: %v", err)
	}
	issuer, err := jwtToken.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != string(TokenAccess) {
		return uuid.Nil, fmt.Errorf("invalid issuer")
	}
	id, err := uuid.Parse(idString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %v", err)
	}
	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authorizationKV := headers.Get("Authorization")
	if len(authorizationKV) == 0 {
		return "", fmt.Errorf("key Authorization does not exist in http header")
	}
	authSplit := strings.Split(authorizationKV, " ")
	if len(authSplit) < 2 || authSplit[0] != "Bearer" {
		return "", fmt.Errorf("malformed authorization header")
	}
	return authSplit[1], nil

}
