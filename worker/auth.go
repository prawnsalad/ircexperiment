package worker

import (
	"errors"
	"math/rand"
	"strconv"
	"time"
)

func RegisterAccount(data *DataRpcWrapper, username, password string) error {
	existingUserID := byteAsInt(data.HashGet("users", username))
	if existingUserID != 0 {
		return errors.New("Username is already in use")
	}

	userID := 0

	// TODO: Better ID generation
	rand.Seed(int64(rand.Intn(10000)))
	for i := 0; i < 200; i++ {
		userID += rand.Intn(1000)
	}

	userKey := "user:" + strconv.Itoa(userID)
	created := int(time.Now().Unix())

	data.HashSet("users", username, intAsByte(userID))
	data.HashSet(userKey, "created", intAsByte(created))
	data.HashSet(userKey, "username", []byte(username))
	data.HashSet(userKey, "password", []byte(password))

	return nil
}

func AuthAccount(client *Client, username, password string) bool {
	println("Auth()", username, password)
	data := client.DataWrapperClient.data
	userIDBytes := data.HashGet("users", username)
	println(username, password, userIDBytes)
	if len(userIDBytes) == 0 {
		return false
	}

	userID := byteAsInt(userIDBytes)
	if userID == 0 {
		return false
	}

	dbPasswordBytes := data.HashGet("user:"+strconv.Itoa(userID), "password")
	dbPassword := string(dbPasswordBytes)
	// TODO: hash this password
	if dbPassword != password {
		return false
	}

	// User has now authed from this point. Username and password matches

	client.SetUserID(userID)
	return true
}
