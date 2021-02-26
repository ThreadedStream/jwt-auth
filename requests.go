package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"os"
	"time"
)

const (
	privateKeyPath = "key256.key" //Private key for signing tokens
)

const (
	//Taking advantage of the fact that we have only one user, by virtue of absence of signup request
	//It's needed in refresh token request, since the latter does not accept guid parameter -- only access and refresh tokens
	UserGuid = "37e3f55c-7c34-439c-ab6d-60644d23cc7f"
)

var (
	signKey = []byte(os.Getenv("SECRET_KEY")) // Might be generated with an aid of openssl
)

func (a *App) ObtainTokenPairApi(w http.ResponseWriter, r *http.Request) {
	var userCreds RequestEntities

	err := json.NewDecoder(r.Body).Decode(&userCreds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok := a.FindUser(userCreds.Guid)
	if !ok {
		http.Error(w, "No such user in database", http.StatusBadRequest)
	}

	accessToken := jwt.New(jwt.SigningMethodHS512)
	refreshToken := jwt.New(jwt.SigningMethodHS512)

	//Cast to MapClaims
	accessExpiresAt := time.Now().Add(time.Minute * 15).Unix()

	accessClaims := accessToken.Claims.(jwt.MapClaims)
	accessClaims["guid"] = userCreds.Guid
	accessClaims["exp"] = accessExpiresAt

	//Generate accessTokenString from accessToken using private key generated earlier
	accessTokenString, err := accessToken.SignedString(signKey)
	if err != nil {
		http.Error(w, "Failed to sign access token string", http.StatusBadRequest)
		return
	}
	refreshExpiresAt := time.Now().Add(time.Hour * 72).Unix()
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["sub"] = 1
	refreshClaims["exp"] = refreshExpiresAt

	refreshTokenString, err := refreshToken.SignedString(signKey)
	if err != nil {
		http.Error(w, "Failed to sign refresh token string", http.StatusBadRequest)
		return
	}

	//Checksum is needed for ensuring that access token sent in request is related with refresh token
	checksum := fmt.Sprintf("%v", sha256.Sum256([]byte(accessTokenString+refreshTokenString)))

	err = a.PutRefreshTokenIntoDatabase(userCreds.Guid, refreshTokenString, checksum, refreshExpiresAt)
	if err != nil {
		ResponseShortcut(w, http.StatusUnauthorized, err.Error())
	}

	ResponseShortcut(w, http.StatusOK, map[string]string{
		"access_token":  accessTokenString,
		"refresh_token": refreshTokenString,
		"guid":          userCreds.Guid,
	})
}

func (a *App) Refresh(w http.ResponseWriter, r *http.Request) {
	var reqEntities RequestEntities

	err := json.NewDecoder(r.Body).Decode(&reqEntities)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	checksum := fmt.Sprintf("%v", sha256.Sum256([]byte(reqEntities.AccessToken+reqEntities.RefreshToken)))

	data, ok := a.FindRefreshTokenAndCheckValidity(checksum, reqEntities.RefreshToken)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
	}

	//Stage when a client has granted permission to obtain necessary tokens
	accessToken := jwt.New(jwt.SigningMethodHS512)
	refreshToken := jwt.New(jwt.SigningMethodHS512)

	//Cast to MapClaims
	accessExpiresAt := time.Now().Add(time.Minute * 15).Unix()

	accessClaims := accessToken.Claims.(jwt.MapClaims)
	accessClaims["guid"] = data["guid"]
	accessClaims["exp"] = accessExpiresAt

	accessTokenString, err := accessToken.SignedString(signKey)

	if err != nil {
		http.Error(w, "Failed to sign access token string", http.StatusBadRequest)
		return
	}

	refreshExpiresAt := time.Now().Add(time.Hour * 72).Unix()
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["sub"] = 1
	refreshClaims["exp"] = refreshExpiresAt

	refreshTokenString, err := refreshToken.SignedString(signKey)
	if err != nil {
		http.Error(w, "Failed to sign refresh token string", http.StatusBadRequest)
		return
	}

	//Generating checksum once again :)
	checksum = fmt.Sprintf("%v", sha256.Sum256([]byte(accessTokenString+refreshTokenString)))
	err = a.PutRefreshTokenIntoDatabase(UserGuid, refreshTokenString, checksum, refreshExpiresAt)
	if err != nil {
		http.Error(w, "Failed to put data into database", http.StatusBadRequest)
		return
	}

	ResponseShortcut(w, http.StatusOK, map[string]string{
		"access_token":  accessTokenString,
		"refresh_token": refreshTokenString,
	})
}
