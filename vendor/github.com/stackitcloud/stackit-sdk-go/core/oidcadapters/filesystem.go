package oidcadapters

import (
	"context"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var (
	parser *jwt.Parser = jwt.NewParser()
)

func ReadJWTFromFileSystem(tokenFilePath string) OIDCTokenFunc {
	return func(context.Context) (string, error) {
		token, err := os.ReadFile(tokenFilePath)
		if err != nil {
			return "", err
		}
		tokenStr := string(token)
		_, _, err = parser.ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			return "", err
		}
		return tokenStr, nil
	}
}
