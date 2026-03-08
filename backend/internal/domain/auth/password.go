package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(encodedHash string, password string) error
}

type Argon2idHasher struct {
	params Argon2idParams
}

type Argon2idParams struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func DefaultArgon2idParams() Argon2idParams {
	return Argon2idParams{
		Memory:      19 * 1024,
		Iterations:  2,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}
}

func RecommendedArgon2idParams(appEnv string) Argon2idParams {
	params := DefaultArgon2idParams()
	if appEnv != "production" {
		return params
	}

	params.Memory = 64 * 1024
	params.Iterations = 3
	return params
}

func NewArgon2idHasher(params Argon2idParams) Argon2idHasher {
	return Argon2idHasher{params: params}
}

func (h Argon2idHasher) Hash(password string) (string, error) {
	salt := make([]byte, h.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.params.Iterations,
		h.params.Memory,
		h.params.Parallelism,
		h.params.KeyLength,
	)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.params.Memory,
		h.params.Iterations,
		h.params.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (h Argon2idHasher) Verify(encodedHash string, password string) error {
	params, salt, expectedHash, err := parseArgon2idHash(encodedHash)
	if err != nil {
		return ErrInvalidCredentials
	}

	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		uint32(len(expectedHash)),
	)

	if subtle.ConstantTimeCompare(expectedHash, computedHash) != 1 {
		return ErrInvalidCredentials
	}

	return nil
}

func parseArgon2idHash(encodedHash string) (Argon2idParams, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return Argon2idParams{}, nil, nil, fmt.Errorf("unexpected argon2id format")
	}

	if parts[1] != "argon2id" {
		return Argon2idParams{}, nil, nil, fmt.Errorf("unsupported password algorithm")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return Argon2idParams{}, nil, nil, fmt.Errorf("unsupported argon2 version")
	}

	params := Argon2idParams{}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism); err != nil {
		return Argon2idParams{}, nil, nil, fmt.Errorf("invalid argon2 params")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Argon2idParams{}, nil, nil, fmt.Errorf("decode salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Argon2idParams{}, nil, nil, fmt.Errorf("decode hash: %w", err)
	}

	params.SaltLength = uint32(len(salt))
	params.KeyLength = uint32(len(hash))
	return params, salt, hash, nil
}
