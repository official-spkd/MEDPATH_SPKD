package store

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strconv"
	"strings"
	"sync"
)

// PBKDF2-HMAC-SHA256, salt acak 128-bit; teks sandi tidak disimpan dalam User.
func hashPassword(password string) string {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		panic("Sumber acak tidak tersedia")
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, 600000, 32)
	if err != nil {
		panic("Hash kata sandi gagal")
	}
	return "pbkdf2-sha256$600000$" + hex.EncodeToString(salt) + "$" + hex.EncodeToString(key)
}
func cocokPassword(encoded, password string) bool {
	p := strings.Split(encoded, "$")
	if len(p) != 4 || p[0] != "pbkdf2-sha256" {
		return false
	}
	rounds, err := strconv.Atoi(p[1])
	if err != nil || rounds < 600000 || rounds > 1000000 {
		return false
	}
	salt, err := hex.DecodeString(p[2])
	if err != nil || len(salt) != 16 {
		return false
	}
	expected, err := hex.DecodeString(p[3])
	if err != nil || len(expected) != 32 {
		return false
	}
	actual, err := pbkdf2.Key(sha256.New, password, salt, rounds, 32)
	return err == nil && subtle.ConstantTimeCompare(actual, expected) == 1
}

var demoHashOnce sync.Once
var demoHash string

func demoPasswordHash() string {
	demoHashOnce.Do(func() { demoHash = hashPassword(DemoPassword) })
	return demoHash
}
