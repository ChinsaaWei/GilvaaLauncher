package downloader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"GilvaaLauncher/logger"
)

type AssetIndex struct {
	Objects map[string]AssetObject `json:"objects"`
}

type AssetObject struct {
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}

type AssetsDownloader struct {
	downloader *Downloader
	assetsDir  string
}

func NewAssetsDownloader(downloader *Downloader, assetsDir string) *AssetsDownloader {
	return &AssetsDownloader{
		downloader: downloader,
		assetsDir:  assetsDir,
	}
}

func (ad *AssetsDownloader) DownloadAssets(assetIndexInfo AssetIndexInfo, virtual bool) error {
	logger.Info("Downloading assets index...")

	indexCacheDir := filepath.Join(ad.assetsDir, "indexes")
	if err := os.MkdirAll(indexCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create indexes directory: %w", err)
	}

	indexPath := filepath.Join(indexCacheDir, assetIndexInfo.ID+".json")

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		if err := ad.downloader.DownloadFile(assetIndexInfo.URL, indexPath, assetIndexInfo.SHA1); err != nil {
			return fmt.Errorf("failed to download asset index: %w", err)
		}
	}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read asset index: %w", err)
	}

	var assetIndex AssetIndex
	if err := json.Unmarshal(data, &assetIndex); err != nil {
		return fmt.Errorf("failed to parse asset index: %w", err)
	}

	logger.Info("Found %d assets to verify/download", len(assetIndex.Objects))

	if virtual {
		return ad.downloadVirtualAssets(&assetIndex)
	}

	return ad.downloadLegacyAssets(&assetIndex)
}

func (ad *AssetsDownloader) downloadVirtualAssets(assetIndex *AssetIndex) error {
	objectsDir := filepath.Join(ad.assetsDir, "objects")
	if err := os.MkdirAll(objectsDir, 0755); err != nil {
		return err
	}

	downloaded := 0
	skipped := 0

	for name, obj := range assetIndex.Objects {
		hashPrefix := obj.Hash[:2]
		hashPath := filepath.Join(objectsDir, hashPrefix, obj.Hash)
		objectURL := fmt.Sprintf("https://resources.download.minecraft.net/%s/%s", hashPrefix, obj.Hash)

		if _, err := os.Stat(hashPath); err == nil {
			if ad.downloader.verifyHash(hashPath, obj.Hash) {
				skipped++
				continue
			}
		}

		if err := ad.downloader.DownloadFile(objectURL, hashPath, obj.Hash); err != nil {
			logger.Warn("Failed to download asset %s: %v", name, err)
			continue
		}

		downloaded++
		if downloaded%100 == 0 {
			logger.Info("Downloaded %d assets, skipped %d", downloaded, skipped)
		}
	}

	logger.Info("Assets download complete: %d downloaded, %d skipped", downloaded, skipped)
	return nil
}

func (ad *AssetsDownloader) downloadLegacyAssets(assetIndex *AssetIndex) error {
	legacyDir := filepath.Join(ad.assetsDir, "legacy")
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		return err
	}

	objectsDir := filepath.Join(ad.assetsDir, "objects")
	if err := os.MkdirAll(objectsDir, 0755); err != nil {
		return err
	}

	downloaded := 0
	skipped := 0

	for name, obj := range assetIndex.Objects {
		hashPrefix := obj.Hash[:2]
		hashPath := filepath.Join(objectsDir, hashPrefix, obj.Hash)
		legacyPath := filepath.Join(legacyDir, name)

		objectURL := fmt.Sprintf("https://resources.download.minecraft.net/%s/%s", hashPrefix, obj.Hash)

		if _, err := os.Stat(hashPath); err == nil {
			if ad.downloader.verifyHash(hashPath, obj.Hash) {
				if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
					if err := os.MkdirAll(filepath.Dir(legacyPath), 0755); err != nil {
						logger.Warn("Failed to create legacy directory: %v", err)
						continue
					}
					if err := os.Symlink(hashPath, legacyPath); err != nil {
						if err := copyFile(hashPath, legacyPath); err != nil {
							logger.Warn("Failed to copy asset %s: %v", name, err)
							continue
						}
					}
				}
				skipped++
				continue
			}
		}

		if err := ad.downloader.DownloadFile(objectURL, hashPath, obj.Hash); err != nil {
			logger.Warn("Failed to download asset %s: %v", name, err)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(legacyPath), 0755); err != nil {
			logger.Warn("Failed to create legacy directory: %v", err)
			continue
		}

		if err := copyFile(hashPath, legacyPath); err != nil {
			logger.Warn("Failed to copy asset %s: %v", name, err)
			continue
		}

		downloaded++
		if downloaded%100 == 0 {
			logger.Info("Downloaded %d assets, skipped %d", downloaded, skipped)
		}
	}

	logger.Info("Legacy assets download complete: %d downloaded, %d skipped", downloaded, skipped)
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
