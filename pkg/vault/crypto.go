// sailor
// Copyright (C) 2025 SailorHQ and Ashish Shekar (codekidX)

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/hkdf"
)

type SecretRecord struct {
	EncryptedSecret string `json:"encrypted_secret"`
	EncryptedDEK    string `json:"encrypted_dek"`
}

const (
	DEKLength = 32 // AES-256
	NonceSize = 12 // GCM nonce
)

func GenerateDEK() ([]byte, error) {
	dek := make([]byte, DEKLength)
	_, err := rand.Read(dek)
	return dek, err
}

// Derive KEK from a passphrase using HKDF (could also be PBKDF2)
func DeriveKEK(passphrase string, salt []byte) ([]byte, error) {
	hkdf := hkdf.New(sha256.New, []byte(passphrase), salt, nil)
	kek := make([]byte, 32)
	if _, err := io.ReadFull(hkdf, kek); err != nil {
		return nil, err
	}
	return kek, nil
}

// Encrypts secret with DEK using AES-GCM
func EncryptWithDEK(secret string, dek []byte) (cipherTextBase64 string, nonce []byte, err error) {
	block, err := aes.NewCipher(dek)
	if err != nil {
		return "", nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", nil, err
	}

	nonce = make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return "", nil, err
	}

	cipherText := aesgcm.Seal(nil, nonce, []byte(secret), nil)
	return base64.StdEncoding.EncodeToString(append(nonce, cipherText...)), nonce, nil
}

// Decrypts secret with DEK
func DecryptWithDEK(cipherTextBase64 string, dek []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cipherTextBase64)
	if err != nil {
		return "", err
	}
	if len(data) < NonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce := data[:NonceSize]
	ciphertext := data[NonceSize:]

	block, err := aes.NewCipher(dek)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plain, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	return string(plain), err
}

// Encrypts DEK with KEK (AES-GCM)
func EncryptDEK(dek, kek []byte) (string, error) {
	block, err := aes.NewCipher(kek)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	enc := aesgcm.Seal(nonce, nonce, dek, nil)
	return base64.StdEncoding.EncodeToString(enc), nil
}

// Decrypts DEK with KEK (AES-GCM)
func DecryptDEK(encryptedDEK string, kek []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedDEK)
	if err != nil {
		return nil, err
	}
	nonce := data[:NonceSize]
	ciphertext := data[NonceSize:]

	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return aesgcm.Open(nil, nonce, ciphertext, nil)
}
