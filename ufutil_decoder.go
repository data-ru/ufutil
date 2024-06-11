package ufutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

type ApiCardapio []struct {
	Titulo            string `json:"titulo"`
	Local             string `json:"local"`
	Mensagem          string `json:"mensagem"`
	PrincipalAlmoco   string `json:"principal_almoco"`
	VegetarianoAlmoco string `json:"vegetariano_almoco"`
	ArrozAlmoco       string `json:"arroz_almoco"`
	FeijaoAlmoco      string `json:"feijao_almoco"`
	GuarnicaoAlmoco   string `json:"guarnicao_almoco"`
	SaladaAlmoco      string `json:"salada_almoco"`
	SobremesaAlmoco   string `json:"sobremesa_almoco"`
	SucoAlmoco        string `json:"suco_almoco"`
	Data              string `json:"data"`
	PrincipalJantar   string `json:"principal_jantar"`
	VegetarianoJantar string `json:"vegetariano_jantar"`
	ArrozJantar       string `json:"arroz_jantar"`
	FeijaoJantar      string `json:"feijao_jantar"`
	GuarnicaoJantar   string `json:"guarnicao_jantar"`
	SaladaJantar      string `json:"salada_jantar"`
	SobremesaJantar   string `json:"sobremesa_jantar"`
	SucoJantar        string `json:"suco_jantar"`
	Nid               string `json:"nid"`
}

// Gera uma chave usando PBKDF2 com o salt e a palavra-chave
func generateKey(salt, passphrase string, keySize, iterationCount int) ([]byte, error) {
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		return nil, err
	}
	key := pbkdf2.Key([]byte(passphrase), saltBytes, iterationCount, keySize, sha1.New)
	return key, nil
}

// Decrypt the given ciphertext using AES in CBC mode
func decryptAES(salt, iv, passphrase, ciphertext string, keySize, iterationCount int) (string, error) {
	key, err := generateKey(salt, passphrase, keySize, iterationCount)
	if err != nil {
		return "", err
	}
	ivBytes, err := hex.DecodeString(iv)
	if err != nil {
		return "", err
	}
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCDecrypter(block, ivBytes)
	plainBytes := make([]byte, len(ciphertextBytes))
	mode.CryptBlocks(plainBytes, ciphertextBytes)
	plainBytes = pkcs7Unpad(plainBytes, aes.BlockSize)
	return string(plainBytes), nil
}

// Remove PKCS7 padding
func pkcs7Unpad(data []byte, blockSize int) []byte {

	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}

// Decrypt text using the passphrase
func decryptText(encryptedText, passphrase string) (string, error) {
	const (
		salt      = "3FF2EC019C627B945225DEBAD71A01B6985FE84C95A70EB132882F88C0A59A55"
		iv        = "F27D5C9927726BCEFE7510B1BDD3D137"
		keySize   = 16
		iterCount = 10
	)
	return decryptAES(salt, iv, passphrase, encryptedText, keySize, iterCount)
}

// Process the encoded JSON string
func decryptJSON(encodedJSONStr string) (ApiCardapio, error) {
	const markerB64 = "RzJiMVVGWU1Zak5WaVBaWTZiU3B2SG5OWXhI"
	marker, err := base64.StdEncoding.DecodeString(markerB64)
	if err != nil {
		return nil, err
	}
	markerStr := string(marker)
	startIndex := strings.LastIndex(encodedJSONStr, markerStr)
	if startIndex == -1 {
		return nil, nil
	}
	passphrase := encodedJSONStr[startIndex+len(markerStr):]
	if passphrase != "" {
		encryptedText := encodedJSONStr[:startIndex]
		decryptedText, err := decryptText(encryptedText, passphrase)
		if err != nil {
			return nil, err
		}
		if decryptedText == "" {
			return nil, errors.New("problema ao decifrar texto: foi retornado um texto vazio")
		}
		var result ApiCardapio
		err = json.Unmarshal([]byte(decryptedText), &result)
		if err != nil {
			fmt.Println(decryptedText)
			return nil, err
		}
		fmt.Println(decryptedText)
		return result, nil
	}
	return nil, nil
}

func decryptJsonAsString(encodedJSONStr string) (string, error) {
	const markerB64 = "RzJiMVVGWU1Zak5WaVBaWTZiU3B2SG5OWXhI"
	marker, err := base64.StdEncoding.DecodeString(markerB64)
	if err != nil {
		return "", err
	}
	markerStr := string(marker)
	startIndex := strings.LastIndex(encodedJSONStr, markerStr)
	if startIndex == -1 {
		return "", nil
	}
	passphrase := encodedJSONStr[startIndex+len(markerStr):]
	if passphrase != "" {
		encryptedText := encodedJSONStr[:startIndex]
		decryptedText, err := decryptText(encryptedText, passphrase)
		if err != nil {
			return "", err
		}
		if decryptedText == "" {
			return "", errors.New("decryptor returned an empty string")
		}

		return decryptedText, nil
	}
	return "", nil
}
