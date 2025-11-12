package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(pass string) (string, error) {
	hasPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hasPass), err
}
func ReverseHash(pass string, user string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user), []byte(pass))
	if err != nil {
		return err
	}
	return nil
}
