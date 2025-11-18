package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

// Password hashing parameters (reasonable defaults)
var (
	argonTime    uint32 = 1
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32
)

// HashPassword returns a string which encodes the parameters, salt and hash.
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	// store as: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	encoded := strings.Join([]string{"$argon2id", "v=19", "m=" + strconv.Itoa(int(argonMemory)) + ",t=" + strconv.Itoa(int(argonTime)) + ",p=" + strconv.Itoa(int(argonThreads)), b64Salt, b64Hash}, "$")
	return encoded, nil
}

// VerifyPassword checks password against encoded hash
func VerifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false
	}
	// parts: "", "argon2id", "v=19", "m=...,t=...,p=...", salt, hash
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}
	// parse params
	params := parts[3]
	// example: m=65536,t=1,p=4
	var m, t, p uint64
	for _, kv := range strings.Split(params, ",") {
		kvp := strings.Split(kv, "=")
		if len(kvp) != 2 {
			continue
		}
		switch kvp[0] {
		case "m":
			m, _ = strconv.ParseUint(kvp[1], 10, 32)
		case "t":
			t, _ = strconv.ParseUint(kvp[1], 10, 32)
		case "p":
			p, _ = strconv.ParseUint(kvp[1], 10, 8)
		}
	}
	derived := argon2.IDKey([]byte(password), salt, uint32(t), uint32(m), uint8(p), uint32(len(hash)))
	return subtleConstantTimeCompare(hash, derived)
}

// constant time compare
func subtleConstantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff uint8 = 0
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

// JWT helpers
var jwtSecret []byte

func getJWTSecret() []byte {
	if jwtSecret != nil {
		return jwtSecret
	}
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		// fallback: generate ephemeral secret (not for production)
		tmp := make([]byte, 32)
		_, _ = rand.Read(tmp)
		s = base64.RawStdEncoding.EncodeToString(tmp)
	}
	jwtSecret = []byte(s)
	return jwtSecret
}

func GenerateToken(userID uint, username string, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"name": username,
		"role": role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

func ParseToken(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return getJWTSecret(), nil
	})
}

// AuthMiddleware enforces a valid JWT in Authorization header: "Bearer <token>"
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}
		tokStr := parts[1]
		token, err := ParseToken(tokStr)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		// attach claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user", claims)
		}
		c.Next()
	}
}
