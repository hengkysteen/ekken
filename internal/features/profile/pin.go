package profile

import (
	"golang.org/x/crypto/bcrypt"
)

func hashPIN(pin string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	return string(bytes), err
}

func verifyPINHash(encoded, pin string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encoded), []byte(pin))
	return err == nil
}
