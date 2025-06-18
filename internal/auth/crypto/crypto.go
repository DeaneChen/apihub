package crypto

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

// CryptoService 加密服务接口
type CryptoService interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// AESCryptoService AES加密服务实现
type AESCryptoService struct {
	key []byte
}

// NewAESCryptoService 创建AES加密服务
func NewAESCryptoService(secretKey string) *AESCryptoService {
	// 使用SHA256生成32字节的密钥
	hash := sha256.Sum256([]byte(secretKey))
	return &AESCryptoService{
		key: hash[:],
	}
}

// pkcs7Padding 添加PKCS7填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// pkcs7UnPadding 移除PKCS7填充
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("empty data")
	}

	padding := int(data[length-1])
	if padding > length {
		return nil, errors.New("invalid padding size")
	}

	return data[:length-padding], nil
}

// Encrypt 使用AES-ECB模式加密明文
func (s *AESCryptoService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", errors.New("明文不能为空")
	}

	// 创建AES cipher
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("创建cipher失败: %w", err)
	}

	// 对数据进行PKCS7填充
	plainBytes := []byte(plaintext)
	plainBytes = pkcs7Padding(plainBytes, block.BlockSize())

	// 加密
	ciphertext := make([]byte, len(plainBytes))
	blockSize := block.BlockSize()

	// ECB模式加密
	for i := 0; i < len(plainBytes); i += blockSize {
		block.Encrypt(ciphertext[i:i+blockSize], plainBytes[i:i+blockSize])
	}

	// 返回base64编码的结果
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 使用AES-ECB模式解密密文
func (s *AESCryptoService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", errors.New("密文不能为空")
	}

	// base64解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64解码失败: %w", err)
	}

	// 创建AES cipher
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("创建cipher失败: %w", err)
	}

	// 检查数据长度
	blockSize := block.BlockSize()
	if len(data)%blockSize != 0 {
		return "", errors.New("密文长度不是块大小的整数倍")
	}

	// 解密
	plaintext := make([]byte, len(data))

	// ECB模式解密
	for i := 0; i < len(data); i += blockSize {
		block.Decrypt(plaintext[i:i+blockSize], data[i:i+blockSize])
	}

	// 移除填充
	plaintext, err = pkcs7UnPadding(plaintext)
	if err != nil {
		return "", fmt.Errorf("移除填充失败: %w", err)
	}

	return string(plaintext), nil
}

// GenerateSecretKey 生成随机密钥
func GenerateSecretKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate secret key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
