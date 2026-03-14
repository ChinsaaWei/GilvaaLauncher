package downloader

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mclauncher/logger"
)

type DownloadProgress struct {
	TotalBytes      int64
	DownloadedBytes int64
	Speed           float64
}

type DownloadCallback func(progress DownloadProgress)

type Downloader struct {
	client    *http.Client
	callback  DownloadCallback
	userAgent string
}

func NewDownloader() *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 30 * time.Minute,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Minecraft Launcher/1.0.0",
	}
}

func (d *Downloader) SetCallback(callback DownloadCallback) {
	d.callback = callback
}

func (d *Downloader) DownloadFile(url, destPath string, expectedHash string) error {
	logger.Debug("Downloading %s to %s", url, destPath)

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if expectedHash != "" {
		if d.verifyHash(destPath, expectedHash) {
			logger.Info("File already exists and matches hash: %s", filepath.Base(destPath))
			return nil
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", d.userAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	tempPath := destPath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	progress := DownloadProgress{
		TotalBytes: resp.ContentLength,
	}

	startTime := time.Now()
	var lastUpdate time.Time
	var lastBytes int64

	buffer := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("failed to write file: %w", writeErr)
			}

			progress.DownloadedBytes += int64(n)

			now := time.Now()
			if now.Sub(lastUpdate) > 500*time.Millisecond {
				elapsed := now.Sub(startTime).Seconds()
				if elapsed > 0 {
					progress.Speed = float64(progress.DownloadedBytes-lastBytes) / (now.Sub(lastUpdate).Seconds()) / 1024 / 1024
				}
				lastUpdate = now
				lastBytes = progress.DownloadedBytes

				if d.callback != nil {
					d.callback(progress)
				}
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
	}

	if expectedHash != "" {
		if err := file.Close(); err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}

		if !d.verifyHash(tempPath, expectedHash) {
			os.Remove(tempPath)
			return fmt.Errorf("hash mismatch for downloaded file")
		}
	}

	if err := os.Rename(tempPath, destPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	logger.Info("Downloaded: %s", filepath.Base(destPath))
	return nil
}

func (d *Downloader) verifyHash(filePath, expectedHash string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false
	}

	actualHash := strings.ToLower(hex.EncodeToString(hash.Sum(nil)))
	expectedHash = strings.ToLower(expectedHash)

	return actualHash == expectedHash
}

func (d *Downloader) DownloadJSON(url string, target interface{}) error {
	logger.Debug("Downloading JSON from %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", d.userAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download JSON: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}
