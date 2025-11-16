package user

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/muhammadheryan/e-commerce/cmd/config"
	"github.com/muhammadheryan/e-commerce/constant"
	"github.com/muhammadheryan/e-commerce/model"
	redisrepo "github.com/muhammadheryan/e-commerce/repository/redis"
	userrepo "github.com/muhammadheryan/e-commerce/repository/user"
	"github.com/muhammadheryan/e-commerce/utils/errors"
	"github.com/muhammadheryan/e-commerce/utils/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserApp interface {
	Register(ctx context.Context, req *model.RegisterRequest) (*model.RegisterResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
	ValidateToken(ctx context.Context, tokenString string) (uint64, error)
}

type UserAppImpl struct {
	config    *config.Config
	userRepo  userrepo.UserRepository
	redisRepo redisrepo.Repository
}

func NewUserApp(config *config.Config, userRepo userrepo.UserRepository, redisRepo redisrepo.Repository) UserApp {
	return &UserAppImpl{
		config:    config,
		userRepo:  userRepo,
		redisRepo: redisRepo,
	}
}

func (s *UserAppImpl) Register(ctx context.Context, req *model.RegisterRequest) (*model.RegisterResponse, error) {
	// Check if user exists by email or phone
	existingUser, err := s.userRepo.Get(ctx, &model.UserFilter{Email: req.Email})
	if err != nil {
		logger.Error("[Register] err userRepo.Get email", zap.String("error", err.Error()))
		return nil, errors.SetCustomError(constant.ErrInternal)
	}

	if existingUser != nil {
		return nil, errors.SetCustomError(constant.ErrCredentialExists)
	}

	existingUser, err = s.userRepo.Get(ctx, &model.UserFilter{Phone: req.Phone})
	if err != nil {
		logger.Error("[Register] err userRepo.Get phone", zap.String("error", err.Error()))
		return nil, errors.SetCustomError(constant.ErrInternal)
	}
	if existingUser != nil {
		return nil, errors.SetCustomError(constant.ErrCredentialExists)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("[Register] err bcrypt.GenerateFromPassword", zap.String("error", err.Error()))
		return nil, errors.SetCustomError(constant.ErrInternal)
	}

	// Create user entity
	userEntity := &model.UserEntity{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hashedPassword),
	}

	// Save to database
	userEntity, err = s.userRepo.Create(ctx, userEntity)
	if err != nil {
		logger.Error("[Register] err userRepo.Create", zap.String("error", err.Error()))
		return nil, errors.SetCustomError(constant.ErrInternal)
	}

	return &model.RegisterResponse{
		Name:  userEntity.Name,
		Email: userEntity.Email,
	}, nil
}

func (s *UserAppImpl) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	// Find user by email or phone
	filter := &model.UserFilter{}
	if isEmail(req.Identifier) {
		filter.Email = req.Identifier
	} else {
		filter.Phone = req.Identifier
	}

	user, err := s.userRepo.Get(ctx, filter)
	if err != nil {
		logger.Error("[Login] err userRepo.Get", zap.String("error", err.Error()))
		return nil, errors.SetCustomError(constant.ErrInternal)
	}

	if user == nil {
		return nil, errors.SetCustomError(constant.ErrNotFound)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.SetCustomError(constant.ErrInvalidPassword)
	}

	// Generate JWT token
	token, jti, err := s.generateJWT(user.ID)
	if err != nil {
		logger.Error("[Login] err generateJWT", zap.String("error", err.Error()))
		return nil, errors.SetCustomError(constant.ErrInternal)
	}

	// Store session in Redis
	err = s.redisRepo.SetSession(ctx, jti, user.ID, s.config.Auth.SessionExpTime)
	if err != nil {
		logger.Error("[Login] err SetSession", zap.String("error", err.Error()))
		return nil, errors.SetCustomError(constant.ErrInternal)
	}

	return &model.LoginResponse{
		Name:  user.Name,
		Email: user.Email,
		Token: token,
	}, nil
}

func (s *UserAppImpl) ValidateToken(ctx context.Context, tokenString string) (uint64, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.Auth.JWTSecret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid claims")
	}

	// Extract userID from Subject
	userIDStr := claims.Subject
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user id in token")
	}

	// Extract JTI (Token ID)
	jti := claims.ID
	if jti == "" {
		return 0, fmt.Errorf("token missing jti")
	}

	// Check Redis session key
	redisUserID, err := s.redisRepo.GetSession(ctx, jti)
	if err != nil {
		return 0, fmt.Errorf("invalid or expired session")
	}

	// Compare Redis userID with claims.Subject
	if redisUserID != userID {
		return 0, fmt.Errorf("token does not match user session")
	}

	return userID, nil
}

// generateJWT creates a JWT token for the user
func (s *UserAppImpl) generateJWT(userID uint64) (string, string, error) {
	newUUID, _ := uuid.NewRandom()
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.Auth.JWTExpiration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        newUUID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.Auth.JWTSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, claims.ID, nil
}

// isEmail checks if identifier looks like an email
func isEmail(identifier string) bool {
	for _, r := range identifier {
		if r == '@' {
			return true
		}
	}
	return false
}
