package auth

import (
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	token := strings.Split(auth, " ")
	return token[1], nil
}
func HashPassword(password string) (hash string, err error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}
func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
