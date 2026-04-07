package services

import (
	"time"

	"lite-collector/models"
	"lite-collector/repository"
	"lite-collector/utils"

	"github.com/golang-jwt/jwt/v5"
)

// UserService handles user-related operations
type UserService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo repository.UserRepository, jwtSecret []byte) *UserService {
	return &UserService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Login exchanges a WeChat code for a JWT token, creating the user if needed.
// Returns the signed token string and the user record.
func (s *UserService) Login(code string) (string, *models.User, error) {
	// TODO Phase 3: replace with real WeChat code exchange (cfg.Wechat.AppID/AppSecret)
	openid := "simulated_openid_" + code
	nickname := "WeChat User"
	avatarURL := ""

	user, err := s.userRepo.FindByOpenID(openid)
	if err != nil {
		// User not found — create a new one
		user = &models.User{
			OpenID:    openid,
			Nickname:  nickname,
			AvatarURL: avatarURL,
		}
		if err := s.userRepo.Create(user); err != nil {
			return "", nil, utils.ErrLoginFailed
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"openid":  user.OpenID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, utils.ErrLoginFailed
	}

	return tokenString, user, nil
}

// FindByID finds a user by ID
func (s *UserService) FindByID(id uint64) (*models.User, error) {
	return s.userRepo.FindByID(id)
}
