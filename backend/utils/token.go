package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/VinukaThejana/auth/backend/errors"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/models"
	"github.com/VinukaThejana/auth/backend/schemas"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Token is a struct that gorups all the token related operations
type Token struct{}

// TokenDetails is a struct that contains the data that need to be used when
// creating tokens
type TokenDetails struct {
	Token     *string
	TokenUUID string
	UserID    string
	ExpiresIn *int64
}

// CreateRefreshToken is a function that is used to create a refresh token
func (Token) CreateRefreshToken(h *initialize.H, userID, privateKey string, ttl time.Duration, reqData struct {
	IPAddress       string
	Location        string
	Device          string
	OS              string
	AccessTokenUUID string
},
) (*TokenDetails, error) {
	var refreshTokenDetails schemas.RefreshTokenDetails

	uid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	td := &TokenDetails{
		ExpiresIn: new(int64),
		Token:     new(string),
	}

	*td.ExpiresIn = now.Add(ttl).Unix()
	td.TokenUUID = uid.String()
	td.UserID = userID

	decodePrivateKey, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(decodePrivateKey)
	if err != nil {
		return nil, err
	}

	claims := make(jwt.MapClaims)
	claims["sub"] = userID
	claims["token_uuid"] = td.TokenUUID
	claims["exp"] = td.ExpiresIn
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()

	*td.Token, err = jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if err != nil {
		return nil, err
	}

	refreshTokenDetails = schemas.RefreshTokenDetails{
		UserID:          userID,
		AccessTokenUUID: reqData.AccessTokenUUID,
	}

	tokenVal, err := json.Marshal(refreshTokenDetails)
	if err != nil {
		return nil, err
	}

	userUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	err = h.DB.DB.Create(&models.Sessions{
		UserID:    userUID,
		TokenID:   uid,
		IPAddress: "",
		Location:  "",
		OS:        "",
		Device:    "",
		LoginAt:   time.Now().UTC(),
		ExpiresAt: *td.ExpiresIn,
	}).Error
	if err != nil {
		if ok := (errors.CheckDBError{}.DuplicateKey(err)); !ok {
			return nil, errors.ErrUnauthorized
		}

		return nil, err
	}

	ctx := context.TODO()
	h.R.RS.Set(ctx, td.TokenUUID, string(tokenVal), time.Unix(*td.ExpiresIn, 0).Sub(now))

	return td, nil
}

// CreateAccessToken is a function that is used to create a access token
func (Token) CreateAccessToken(h *initialize.H, userID, privateKey string, ttl time.Duration) (*TokenDetails, error) {
	uid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	td := &TokenDetails{
		ExpiresIn: new(int64),
		Token:     new(string),
	}

	*td.ExpiresIn = now.Add(ttl).Unix()
	td.TokenUUID = uid.String()
	td.UserID = userID

	decodePrivateKey, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(decodePrivateKey)
	if err != nil {
		return nil, err
	}

	claims := make(jwt.MapClaims)
	claims["sub"] = userID
	claims["token_uuid"] = td.TokenUUID
	claims["exp"] = td.ExpiresIn
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()

	*td.Token, err = jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()
	h.R.RS.Set(ctx, td.TokenUUID, userID, time.Unix(*td.ExpiresIn, 0).Sub(now))

	return td, nil
}

// ValidateRefreshToken is a fucntion that is used to validate the refresh token
func (Token) ValidateRefreshToken(h *initialize.H, token, publicKey string) (*TokenDetails, *schemas.RefreshTokenDetails, error) {
	td, val, err := validateToken(h, token, publicKey)
	if err != nil {
		return nil, nil, err
	} else if val == nil {
		return nil, nil, errors.ErrInternalServerError
	}

	var refreshTokenDetails schemas.RefreshTokenDetails
	err = json.Unmarshal([]byte(*val), &refreshTokenDetails)
	if err != nil {
		return nil, nil, errors.ErrInternalServerError
	}

	if refreshTokenDetails.UserID == td.UserID {
		return td, &refreshTokenDetails, nil
	}

	return nil, nil, errors.ErrUnauthorized
}

// ValidateAccessToken is a function that is  used to validate the access token
func (Token) ValidateAccessToken(h *initialize.H, token, publicKey string) (*TokenDetails, error) {
	td, val, err := validateToken(h, token, publicKey)
	if err != nil {
		return nil, err
	} else if val == nil {
		return nil, errors.ErrInternalServerError
	}

	if *val == *&td.UserID {
		return td, nil
	}

	return nil, errors.ErrUnauthorized
}

// DeleteToken is a function to delete a token
func (Token) DeleteToken(h *initialize.H, refreshTokenUUID, accessTokenUUID string) error {
	uid, err := uuid.Parse(refreshTokenUUID)
	if err != nil {
		return err
	}
	err = h.DB.DB.Delete(&models.Sessions{
		TokenID: uid,
	}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUnauthorized
		}

		return err
	}

	ctx := context.TODO()

	pipe := h.R.RS.Pipeline()
	pipe.Del(ctx, refreshTokenUUID)
	pipe.Del(ctx, accessTokenUUID)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func validateToken(h *initialize.H, token, publicKey string) (*TokenDetails, *string, error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, nil, err
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(decodedPublicKey)
	if err != nil {
		return nil, nil, err
	}

	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected method : %s", t.Header["alg"])
		}

		return key, nil
	})
	if err != nil {
		return nil, nil, err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, nil, fmt.Errorf("Validate : invalid token")
	}

	td := &TokenDetails{
		TokenUUID: fmt.Sprint(claims["token_uuid"]),
		UserID:    fmt.Sprint(claims["sub"]),
	}

	ctx := context.TODO()
	val := h.R.RS.Get(ctx, td.TokenUUID).Val()
	if val == "" {
		return nil, nil, errors.ErrUnauthorized
	}

	return td, &val, nil
}

// DeleteExpiredTokens is a function that is used to delete expired session tokens
func (Token) DeleteExpiredTokens(h *initialize.H, userID string) {
	var sessions []models.Sessions
	err := h.DB.DB.Where("user_id = ? AND expires_at <= ?", userID, time.Now().UTC().Unix()).Find(&sessions).Error
	if err != nil {
		log.Error(err, nil)
		return
	}
	if len(sessions) == 0 {
		return
	}

	err = h.DB.DB.Where("1 = 1").Delete(&sessions).Error
	if err != nil {
		log.Error(err, nil)
		return
	}
}
