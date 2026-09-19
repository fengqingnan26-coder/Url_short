package service

import "errors"

var (
	ErrShortCodeTaken = errors.New("short code already taken")
)

var ErrUserNameOrPasswordFailed = errors.New("username or password failed")
var ErrEmailAleadyExist = errors.New("email already exist")
var ErrEmailCodeNotEqual = errors.New("email code not equal")
