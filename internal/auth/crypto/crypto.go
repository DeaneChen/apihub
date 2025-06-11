package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
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

// Encrypt 加密明文
func (s *AESCryptoService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", errors.New("plaintext cannot be empty")
	}

	// 创建AES cipher
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// 返回base64编码的结果
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密密文
func (s *AESCryptoService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", errors.New("ciphertext cannot be empty")
	}

	// base64解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// 创建AES cipher
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 检查数据长度
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// 分离nonce和密文
	nonce, cipherData := data[:nonceSize], data[nonceSize:]

	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
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
