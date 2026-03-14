package util

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CalculateSHA1(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func VerifySHA1(filePath, expectedHash string) (bool, error) {
	actualHash, err := CalculateSHA1(filePath)
	if err != nil {
		return false, err
	}

	return actualHash == expectedHash, nil
}

func CalculateFileHash(filePath string, algorithm string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var hashValue []byte
	switch algorithm {
	case "sha1":
		hash := sha1.New()
		if _, err := io.Copy(hash, file); err != nil {
			return "", fmt.Errorf("failed to calculate hash: %w", err)
		}
		hashValue = hash.Sum(nil)
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	return hex.EncodeToString(hashValue), nil
}

func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}
	return info.Size(), nil
}

func EnsureDirectory(dirPath string) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func RemoveFileIfExists(filePath string) error {
	if FileExists(filePath) {
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
	}
	return nil
}

func CopyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

func MoveFile(src, dst string) error {
	if err := CopyFile(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}
