package auth

import (
	"encoding/hex"

	"github.com/golang-jwt/jwt"
)

// AuthClaims represents the JWT claims used in authentication. These claims should
// be identical to the ones used in our authentication server to create the token.
type AuthClaims struct {
	jwt.StandardClaims
	Uid        uint   `json:"uid"`
	IdProvider string `json:"idp"`
	AppId      string `json:"appid"`
}

var (
	signingSecret []byte
)

func init() {
	secretString := "5779c98432893f4af70e46dc99bbbe66523a7ebd19f75182228eea267fc8fa0e"
	signingSecret, _ = hex.DecodeString(secretString)
}

// VerifyAuthClaims verifies and parses a JWT token string into an AuthClaims
// struct. It takes a JWT token string as input and returns a pointer to an
// AuthClaims struct if the JWT token is valid and successfully parsed, or nil
// if there was an error.
//
// Parameters:
//
//	s (string): The JWT token string to verify and parse.
//
// Returns:
//
//	*AuthClaims: A pointer to an AuthClaims struct if the JWT token is valid
//	and successfully parsed, or nil if there was an error.
func VerifyAuthClaims(s string) *AuthClaims {
	var kf jwt.Keyfunc = func(t *jwt.Token) (interface{}, error) {
		return signingSecret, nil
	}
	t, err := jwt.ParseWithClaims(s, &AuthClaims{}, kf)
	if err != nil {
		return nil
	}
	return t.Claims.(*AuthClaims)
}
