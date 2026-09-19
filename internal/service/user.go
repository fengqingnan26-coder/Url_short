package service

import (
	"context"
	"fmt"

	"github.com/jekyulll/url_shortener/internal/cache"
	"github.com/jekyulll/url_shortener/internal/dto"
	"github.com/jekyulll/url_shortener/internal/model"
	"github.com/jekyulll/url_shortener/internal/repository"
	"github.com/jekyulll/url_shortener/pkg/email"
	"github.com/jekyulll/url_shortener/pkg/hasher"
	"github.com/jekyulll/url_shortener/pkg/jwt"
	"github.com/jekyulll/url_shortener/pkg/randnum"
)

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword string, password string) bool
}

type UserCacher interface {
	GetEmailCode(ctx context.Context, email string) (string, error)
	SetEmailCode(ctx context.Context, email, emailCode string) error
}

type EmailSender interface {
	Send(email, emailCode string) error
}

type NumberRandomer interface {
	Generate() string
}

type JWTer interface {
	Generate(email string, userID int) (string, error)
}

type UserService struct {
	repo           repository.UserRepository
	passwordHasher PasswordHasher
	jwter          JWTer
	userCacher     UserCacher
	emailSender    EmailSender
	numberRandomer NumberRandomer
}

func NewUserService(repo repository.UserRepository, p PasswordHasher, j JWTer, u UserCacher, e EmailSender, n NumberRandomer) *UserService {
	return &UserService{
		repo:           repo,
		passwordHasher: p,
		jwter:          j,
		userCacher:     u,
		emailSender:    e,
		numberRandomer: n,
	}
}

// IsEmailAvailable implements api.UserService.
func (s *UserService) IsEmailAvailable(ctx context.Context, email string) error {
	ok, err := s.repo.IsEmailAvailable(ctx, email)
	if err != nil {
		return err
	}
	if !ok {
		return ErrEmailAleadyExist
	}
	return nil
}

// Register implements api.UserService.
func (s *UserService) Register(ctx context.Context, req dto.RegisterReqeust) (*dto.LoginResponse, error) {
	// 判断验证码
	emailCode, err := s.userCacher.GetEmailCode(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get emailCode from cache: %v", err)
	}
	if emailCode != req.EmailCode {
		return nil, ErrEmailCodeNotEqual
	}
	// hash
	hash, err := s.passwordHasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}
	user := model.User{
		Email:        req.Email,
		PasswordHash: hash,
	}
	// 写入
	if err := s.repo.CreateUser(ctx, &user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}
	// access token
	accessToken, err := s.jwter.Generate(user.Email, int(user.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}
	return &dto.LoginResponse{
		AccessToken: accessToken,
		Email:       user.Email,
		UserID:      int(user.ID),
	}, nil
}

// ResetPassword implements api.UserService.
func (s *UserService) ResetPassword(ctx context.Context, req dto.ForgetPasswordReqeust) (*dto.LoginResponse, error) {
	emailCode, err := s.userCacher.GetEmailCode(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if emailCode != req.EmailCode {
		return nil, ErrEmailCodeNotEqual
	}
	hash, err := s.passwordHasher.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	// 更新 db
	id, err := s.repo.UpdatePasswordByEmail(ctx, hash, req.Email)
	if err != nil {
		return nil, err
	}
	accessToken, err := s.jwter.Generate(req.Email, int(id))
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{
		AccessToken: accessToken,
		Email:       req.Email,
		UserID:      int(id),
	}, nil
}

// SendEmailCode implements api.UserService.
func (s *UserService) SendEmailCode(ctx context.Context, email string) error {
	emailCode := s.numberRandomer.Generate()
	if err := s.emailSender.Send(email, emailCode); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	if err := s.userCacher.SetEmailCode(ctx, email, emailCode); err != nil {
		return fmt.Errorf("failed to set email to cache: %v", err)
	}
	return nil
}

func (s *UserService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %v", err)
	}

	if !s.passwordHasher.ComparePassword(user.PasswordHash, req.Password) {
		return nil, ErrUserNameOrPasswordFailed
	}

	accessToken, err := s.jwter.Generate(user.Email, int(user.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to genrate access token: %v", err)
	}

	return &dto.LoginResponse{
		AccessToken: accessToken,
		Email:       user.Email,
		UserID:      int(user.ID),
	}, nil
}

var _ PasswordHasher = (*hasher.PasswordHash)(nil)
var _ UserCacher = (*cache.RedisCache)(nil)
var _ EmailSender = (*email.EmailSend)(nil)
var _ NumberRandomer = (*randnum.RandNum)(nil)
var _ JWTer = (*jwt.JWT)(nil)
