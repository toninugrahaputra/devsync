package service

import (
	"context"
	"crypto/rand"
	"devsync/internal/model"
	"devsync/internal/repository"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

var ErrGoogleNotConfigured = errors.New("google sign-in is not configured on this server")

type AuthService interface {
	Register(name, email, password string) (*model.LoginResponse, error)
	Login(email, password string) (*model.LoginResponse, error)
	LoginWithGoogle(ctx context.Context, credential string) (*model.LoginResponse, error)
	ValidateToken(tokenString string) (*model.User, error)
}

type authService struct {
	userRepo       repository.UserRepository
	jwtSecret      []byte
	googleClientID string
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "devsync-secret-key"
	}
	return &authService{
		userRepo:       userRepo,
		jwtSecret:      []byte(secret),
		googleClientID: os.Getenv("GOOGLE_CLIENT_ID"),
	}
}

func (s *authService) Register(name, email, password string) (*model.LoginResponse, error) {
	if existing, _ := s.userRepo.GetByEmail(email); existing != nil {
		return nil, errors.New("email already registered")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{Name: name, Email: email, Password: string(hashed), AuthProvider: "password"}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return s.issueToken(user)
}

func (s *authService) Login(email, password string) (*model.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return s.issueToken(user)
}

// LoginWithGoogle verifies the ID token (a signed JWT) that Google Identity
// Services hands the frontend after a successful Google sign-in, and either
// logs in the matching account by email or creates one on the fly. There's
// no password for a Google-created account, so we store an unusable random
// hash — such users always come back through this path, never /auth/login.
func (s *authService) LoginWithGoogle(ctx context.Context, credential string) (*model.LoginResponse, error) {
	if s.googleClientID == "" {
		return nil, ErrGoogleNotConfigured
	}

	payload, err := idtoken.Validate(ctx, credential, s.googleClientID)
	if err != nil {
		return nil, errors.New("invalid google credential")
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	if email == "" {
		return nil, errors.New("google account has no email")
	}
	if name == "" {
		name = email
	}

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		randomPassword := make([]byte, 32)
		if _, err := rand.Read(randomPassword); err != nil {
			return nil, err
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(hex.EncodeToString(randomPassword)), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user = &model.User{Name: name, Email: email, Password: string(hashed), AuthProvider: "google"}
		if err := s.userRepo.Create(user); err != nil {
			return nil, err
		}
	}

	return s.issueToken(user)
}

func (s *authService) issueToken(user *model.User) (*model.LoginResponse, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token: tokenString,
		User:  *user,
	}, nil
}

func (s *authService) ValidateToken(tokenString string) (*model.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID := uint(claims["user_id"].(float64))
	return s.userRepo.GetByID(userID)
}
