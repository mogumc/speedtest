// 加密类
// @author MoGuQAQ
// @version 1.0.0

package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Unpadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	padLength := int(data[length-1])
	if padLength < 1 || padLength > aes.BlockSize {
		return nil, fmt.Errorf("invalid padding length")
	}
	for i := length - padLength; i < length; i++ {
		if data[i] != byte(padLength) {
			return nil, fmt.Errorf("invalid padding value")
		}
	}
	return data[:(length - padLength)], nil
}

func AesEncrypt(data []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	encryptBytes := pkcs7Padding(data, blockSize)
	crypted := make([]byte, len(encryptBytes))
	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(crypted, encryptBytes)
	return crypted, nil
}

func AesDecrypt(crypted []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(crypted))
	blockMode.CryptBlocks(decrypted, crypted)
	decrypted, err = pkcs7Unpadding(decrypted)
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}
