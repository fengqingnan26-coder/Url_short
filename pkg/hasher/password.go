package hasher

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHash struct{}

func NewPassworkHash() *PasswordHash {
	return &PasswordHash{}
}

// 使用 bcrypt 算法，自动加盐
func (p *PasswordHash) HashPassword(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashBytes), nil
}

func (p *PasswordHash) ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return nil == err
}
