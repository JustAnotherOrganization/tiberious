package types

import (
	"crypto/rand"
	"crypto/sha256"

	"github.com/pborman/uuid"

	"golang.org/x/crypto/pbkdf2"
)

//User ...
type User struct {
	ID         uuid.UUID
	Type       string
	Username   string
	LoginName  string
	Email      string
	Password   string
	Salt       string
	Connected  bool
	Authorized bool
	Rooms      []string
}

//HashPassword ..
func HashPassword(password, salt string) string {
	return string(pbkdf2.Key([]byte(password), []byte(salt), 4096, 32, sha256.New))
}

func (user *User) isPassword(passwordTest string) bool {
	return HashPassword(passwordTest, user.Salt) == user.Password
}

//NewSalt returns a new, 7 character salt. TODO make this a common method of sorts?
func NewSalt() string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, 7)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
