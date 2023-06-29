package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/VinukaThejana/auth/backend/errors"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/schemas"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
		LoginAt:         now,
		IPAddress:       reqData.IPAddress,
		Location:        reqData.Location,
		OS:              reqData.OS,
		Device:          reqData.Device,
		AccessTokenUUID: reqData.AccessTokenUUID,
	}

	tokenVal, err := json.Marshal(refreshTokenDetails)
	if err != nil {
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

// ValidateToken is a function that is used to validate the passed token
func (Token) ValidateToken(h *initialize.H, token, publicKey string, isAccessToken bool) (*TokenDetails, error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, err
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(decodedPublicKey)
	if err != nil {
		return nil, err
	}

	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected method : %s", t.Header["alg"])
		}

		return key, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("Validate : invalid token")
	}

	td := &TokenDetails{
		TokenUUID: fmt.Sprint(claims["token_uuid"]),
		UserID:    fmt.Sprint(claims["sub"]),
	}

	ctx := context.TODO()
	val := h.R.RS.Get(ctx, td.TokenUUID).Val()
	if val == "" {
		return nil, errors.ErrUnauthorized
	}

	if isAccessToken {
		if val == td.UserID {
			return td, nil
		}

		return nil, errors.ErrUnauthorized
	}

	var refreshTokenDetails schemas.RefreshTokenDetails
	err = json.Unmarshal([]byte(val), &refreshTokenDetails)
	if err != nil {
		return nil, errors.ErrInternalServerError
	}

	if refreshTokenDetails.UserID == td.UserID {
		return td, nil
	}

	return nil, errors.ErrUnauthorized
}

// DeleteToken is a function to delete a token
func (Token) DeleteToken(h *initialize.H, token string) error {
	ctx := context.Background()
	err := h.R.RS.Del(ctx, token).Err()
	if err != nil {
		return err
	}

	return nil
}
