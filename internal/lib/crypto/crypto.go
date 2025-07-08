package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
)

var (
	iv  = []byte("1234567890123456")
	key []byte
)

func InitCrypto(secret string) {

	hash := sha256.Sum256([]byte(secret))
	key = hash[:]
}

func pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

// Unpad удаляет PKCS7 паддинг
func unpad(src []byte) ([]byte, error) {
	length := len(src)
	if length == 0 {
		return nil, fmt.Errorf("некорректная длина для unpad")
	}
	padding := int(src[length-1])
	if padding > aes.BlockSize || padding == 0 {
		return nil, fmt.Errorf("неверный padding")
	}
	return src[:length-padding], nil
}

func HashTgID(tgID int64) (string, error) {
	plain := pad([]byte(strconv.FormatInt(tgID, 10)))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(plain))
	mode.CryptBlocks(encrypted, plain)

	return hex.EncodeToString(encrypted), nil
}

// DecryptTgID расшифровывает user_id обратно
func DecryptTgID(cipherHex string) (int64, error) {
	cipherBytes, err := hex.DecodeString(cipherHex)
	if err != nil {
		return 0, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(cipherBytes))
	mode.CryptBlocks(decrypted, cipherBytes)

	unpadded, err := unpad(decrypted)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(string(unpadded), 10, 64)
}
