package downloader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"mclauncher/logger"
)

const (
	VersionManifestURL = "https://launchermeta.mojang.com/mc/game/version_manifest.json"
)

type VersionManifest struct {
	Latest   LatestVersions `json:"latest"`
	Versions []VersionEntry `json:"versions"`
}

type LatestVersions struct {
	Release  string `json:"release"`
	Snapshot string `json:"snapshot"`
}

type VersionEntry struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	Time        string `json:"time"`
	ReleaseTime string `json:"releaseTime"`
	Sha1        string `json:"sha1"`
}

type VersionInfo struct {
	ID                 string                  `json:"id"`
	Type               string                  `json:"type"`
	Time               string                  `json:"time"`
	ReleaseTime        string                  `json:"releaseTime"`
	MainClass          string                  `json:"mainClass"`
	Arguments          *Arguments              `json:"arguments"`
	MinecraftArguments string                  `json:"minecraftArguments"`
	Libraries          []Library               `json:"libraries"`
	Downloads          map[string]FileDownload `json:"downloads"`
	AssetIndex         AssetIndexInfo          `json:"assetIndex"`
	Assets             string                  `json:"assets"`
	JavaVersion        *JavaVersion            `json:"javaVersion"`
}

type Arguments struct {
	Game []interface{} `json:"game"`
	JVM  []interface{} `json:"jvm"`
}

type ArgumentRule struct {
	Rules []Rule      `json:"rules,omitempty"`
	Value interface{} `json:"value"`
}

type Rule struct {
	Action   string          `json:"action"`
	OS       *OSRule         `json:"os,omitempty"`
	Features map[string]bool `json:"features,omitempty"`
}

type OSRule struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	Arch    string `json:"arch,omitempty"`
}

type Library struct {
	Name      string                     `json:"name"`
	Downloads map[string]LibraryDownload `json:"downloads"`
	Rules     []Rule                     `json:"rules,omitempty"`
	Natives   map[string]string          `json:"natives,omitempty"`
}

type LibraryDownload struct {
	Path string `json:"path"`
	SHA1 string `json:"sha1"`
	Size int64  `json:"size"`
	URL  string `json:"url"`
}

type FileDownload struct {
	SHA1 string `json:"sha1"`
	Size int64  `json:"size"`
	URL  string `json:"url"`
}

type AssetIndexInfo struct {
	ID        string `json:"id"`
	SHA1      string `json:"sha1"`
	Size      int64  `json:"size"`
	TotalSize int64  `json:"totalSize"`
	URL       string `json:"url"`
}

type JavaVersion struct {
	Component    string `json:"component"`
	MajorVersion int    `json:"majorVersion"`
}

type VersionDownloader struct {
	downloader *Downloader
	cacheDir   string
}

func NewVersionDownloader(downloader *Downloader, cacheDir string) *VersionDownloader {
	return &VersionDownloader{
		downloader: downloader,
		cacheDir:   cacheDir,
	}
}

func (vd *VersionDownloader) FetchVersionManifest() (*VersionManifest, error) {
	cachePath := filepath.Join(vd.cacheDir, "version_manifest.json")

	if _, err := os.Stat(cachePath); err == nil {
		var manifest VersionManifest
		data, err := os.ReadFile(cachePath)
		if err == nil {
			if err := json.Unmarshal(data, &manifest); err == nil {
				logger.Debug("Using cached version manifest")
				return &manifest, nil
			}
		}
	}

	var manifest VersionManifest
	if err := vd.downloader.DownloadJSON(VersionManifestURL, &manifest); err != nil {
		return nil, fmt.Errorf("failed to fetch version manifest: %w", err)
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		logger.Warn("Failed to cache version manifest: %v", err)
	}

	return &manifest, nil
}

func (vd *VersionDownloader) FetchVersionInfo(versionURL string) (*VersionInfo, error) {
	var versionInfo VersionInfo
	if err := vd.downloader.DownloadJSON(versionURL, &versionInfo); err != nil {
		return nil, fmt.Errorf("failed to fetch version info: %w", err)
	}

	return &versionInfo, nil
}

func (vd *VersionDownloader) DownloadVersion(versionID, gameDir string) (*VersionInfo, error) {
	manifest, err := vd.FetchVersionManifest()
	if err != nil {
		return nil, err
	}

	var versionEntry *VersionEntry
	for i := range manifest.Versions {
		if manifest.Versions[i].ID == versionID {
			versionEntry = &manifest.Versions[i]
			break
		}
	}

	if versionEntry == nil {
		return nil, fmt.Errorf("version %s not found", versionID)
	}

	versionInfo, err := vd.FetchVersionInfo(versionEntry.URL)
	if err != nil {
		return nil, err
	}

	versionDir := filepath.Join(gameDir, "versions", versionID)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create version directory: %w", err)
	}

	jsonPath := filepath.Join(versionDir, versionID+".json")
	data, err := json.MarshalIndent(versionInfo, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write version JSON: %w", err)
	}

	jarPath := filepath.Join(versionDir, versionID+".jar")
	if clientJar, ok := versionInfo.Downloads["client"]; ok {
		if err := vd.downloader.DownloadFile(clientJar.URL, jarPath, clientJar.SHA1); err != nil {
			return nil, fmt.Errorf("failed to download client JAR: %w", err)
		}
	}

	logger.Info("Version %s downloaded successfully", versionID)
	return versionInfo, nil
}
