package authorize

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type Payload struct {
	UserId    uint
	IsAdmin   bool
	IsSuAdmin bool
	jwt.StandardClaims
}

func GenerateJwt(userId uint, isAdmin bool, isSuAdmin bool, secret []byte) (string, error) {

	expiresAt := time.Now().Add(48 * time.Hour)

	jwtclaims := &Payload{
		UserId:    userId,
		IsAdmin:   isAdmin,
		IsSuAdmin: isSuAdmin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwtclaims)

	tokenString, err := token.SignedString(secret)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenString string, secret []byte) (map[string]interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Payload{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodES256 {
			return nil, fmt.Errorf("invalid token")
		}

		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if token == nil || !token.Valid {
		return nil, fmt.Errorf("token is not valid or its empty")
	}

	claims, ok := token.Claims.(*Payload)

	if !ok {
		return nil, fmt.Errorf("cannot parse claims")
	}

	cred := map[string]interface{}{
		"userId":    claims.UserId,
		"isAdmin":   claims.IsAdmin,
		"isSuAdmin": claims.IsSuAdmin,
	}
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("token expired")
	}
	return cred, nil
}
