package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword parolni bcrypt bilan hashlaydi.
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword berilgan parol hash bilan mos kelishini tekshiradi.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
