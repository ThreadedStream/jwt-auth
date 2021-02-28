package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"time"
)

func (a *App) PutRefreshTokenIntoDatabase(userGuid, refreshToken, checksum string, expiresAt int64) error {
	//Hashing token prior to storing it into database
	bcryptedToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	updateStmt := bson.M{"$set": bson.M{"refresh_token": string(bcryptedToken), "expires_at": expiresAt, "checksum": checksum}}

	_, err = a.UsersCollection.UpdateOne(a.Context,
		bson.D{{"guid", userGuid}}, updateStmt)

	if err != nil {
		return err
	}

	return nil
}

func (a *App) FindUser(userGuid string) bool {
	var data []bson.M

	cur, err := a.UsersCollection.Find(a.Context, bson.M{"guid": userGuid})
	if err != nil {
		return false
	}

	err = cur.All(a.Context, &data)
	if err != nil {
		return false
	}

	if len(data) == 0{
		return false
	}

	//At this point, we are in good position!
	return true
}

//Finds refresh token in database and checks its validity
func (a *App) FindRefreshTokenAndCheckValidity(checksum, refreshToken string) (bson.M, bool) {
	var data []bson.M

	//bcryptedToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)

	cur, err := a.UsersCollection.Find(a.Context, bson.M{"checksum": checksum})
	if err != nil {
		return nil, false
	}

	err = cur.All(a.Context, &data)
	if err != nil {
		return nil, false
	}

	if len(data) == 0 {
		//We don't want to get into trouble accessing fields of an empty map
		return nil, false
	}

	token, ok := data[0]["refresh_token"]
	tokenStr := fmt.Sprintf("%v", token)
	if !ok {
		return nil, false
	} else {
		err = bcrypt.CompareHashAndPassword([]byte(tokenStr), []byte(refreshToken))
		if err != nil {
			return nil, false
		}
		expiresAtStr := fmt.Sprintf("%v", data[0]["expires_at"])
		expiresAt, err := strconv.ParseInt(expiresAtStr, 10, 64)
		if err != nil {
			return nil, false
		}
		if expiresAt < time.Now().Unix() {
			return nil, false
		}
	}

	return data[0], true
}
