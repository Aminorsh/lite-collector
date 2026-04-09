package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	appID     string
	appSecret string
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo repository.UserRepository, jwtSecret []byte, appID, appSecret string) *UserService {
	return &UserService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		appID:     appID,
		appSecret: appSecret,
	}
}

// wxSessionResponse is the JSON structure returned by WeChat jscode2session API
type wxSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// Login exchanges a WeChat code for a JWT token, creating the user if needed.
// Returns the signed token string and the user record.
func (s *UserService) Login(code string) (string, *models.User, error) {
	openid, err := s.exchangeCode(code)
	if err != nil {
		return "", nil, err
	}

	user, err := s.userRepo.FindByOpenID(openid)
	if err != nil {
		user = &models.User{
			OpenID:   openid,
			Nickname: "WeChat User",
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

// exchangeCode calls WeChat jscode2session API to get the user's openid.
// If AppID/AppSecret are not configured, falls back to simulated openid for development.
func (s *UserService) exchangeCode(code string) (string, error) {
	if s.appID == "" || s.appSecret == "" {
		// Fallback: simulated openid for local development without WX credentials
		return "simulated_openid_" + code, nil
	}

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.appID, s.appSecret, code,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", utils.ErrWxCodeExchangeFail
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", utils.ErrWxCodeExchangeFail
	}

	var session wxSessionResponse
	if err := json.Unmarshal(body, &session); err != nil {
		return "", utils.ErrWxCodeExchangeFail
	}

	if session.ErrCode != 0 || session.OpenID == "" {
		return "", utils.ErrWxCodeExchangeFail
	}

	return session.OpenID, nil
}

// FindByID finds a user by ID
func (s *UserService) FindByID(id uint64) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

// UpdateProfile updates the user's nickname and avatar URL.
func (s *UserService) UpdateProfile(userID uint64, nickname, avatarURL string) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, utils.ErrNotFound
	}
	if nickname != "" {
		user.Nickname = nickname
	}
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}
	if err := s.userRepo.Update(user); err != nil {
		return nil, utils.ErrInternal
	}
	return user, nil
}
