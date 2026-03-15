package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"mclauncher/downloader"
	"mclauncher/logger"
)

type Manager struct {
	gameDir       string
	downloader    *downloader.VersionDownloader
	manifest      *downloader.VersionManifest
	manifestMutex sync.RWMutex
}

func NewManager(gameDir string, versionDownloader *downloader.VersionDownloader) *Manager {
	return &Manager{
		gameDir:    gameDir,
		downloader: versionDownloader,
	}
}

func (m *Manager) RefreshManifest() error {
	m.manifestMutex.Lock()
	defer m.manifestMutex.Unlock()

	manifest, err := m.downloader.FetchVersionManifest()
	if err != nil {
		return fmt.Errorf("failed to refresh manifest: %w", err)
	}

	m.manifest = manifest
	logger.Info("Version manifest refreshed successfully")
	return nil
}

func (m *Manager) GetManifest() (*downloader.VersionManifest, error) {
	m.manifestMutex.RLock()
	if m.manifest == nil {
		m.manifestMutex.RUnlock()
		if err := m.RefreshManifest(); err != nil {
			return nil, err
		}
		m.manifestMutex.RLock()
	}
	defer m.manifestMutex.RUnlock()

	return m.manifest, nil
}

func (m *Manager) ListAvailableVersions(versionType string) ([]downloader.VersionEntry, error) {
	manifest, err := m.GetManifest()
	if err != nil {
		return nil, err
	}

	versions := make([]downloader.VersionEntry, 0)
	for _, v := range manifest.Versions {
		if versionType == "" || v.Type == versionType {
			versions = append(versions, v)
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].ReleaseTime > versions[j].ReleaseTime
	})

	return versions, nil
}

func (m *Manager) ListInstalledVersions() ([]string, error) {
	versionsDir := filepath.Join(m.gameDir, "versions")
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read versions directory: %w", err)
	}

	versions := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			jsonPath := filepath.Join(versionsDir, entry.Name(), entry.Name()+".json")
			if _, err := os.Stat(jsonPath); err == nil {
				versions = append(versions, entry.Name())
			}
		}
	}

	sort.Strings(versions)
	return versions, nil
}

func (m *Manager) IsVersionInstalled(versionID string) bool {
	versionDir := filepath.Join(m.gameDir, "versions", versionID)
	jsonPath := filepath.Join(versionDir, versionID+".json")
	jarPath := filepath.Join(versionDir, versionID+".jar")

	_, err1 := os.Stat(jsonPath)
	_, err2 := os.Stat(jarPath)

	return err1 == nil && err2 == nil
}

func (m *Manager) GetVersionInfo(versionID string) (*downloader.VersionInfo, error) {
	versionDir := filepath.Join(m.gameDir, "versions", versionID)
	jsonPath := filepath.Join(versionDir, versionID+".json")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read version JSON: %w", err)
	}

	var versionInfo downloader.VersionInfo
	if err := json.Unmarshal(data, &versionInfo); err != nil {
		return nil, fmt.Errorf("failed to parse version JSON: %w", err)
	}

	return &versionInfo, nil
}

func (m *Manager) InstallVersion(versionID string) (*downloader.VersionInfo, error) {
	logger.Info("Installing version: %s", versionID)

	if m.IsVersionInstalled(versionID) {
		logger.Info("Version %s is already installed", versionID)
		return m.GetVersionInfo(versionID)
	}

	versionInfo, err := m.downloader.DownloadVersion(versionID, m.gameDir)
	if err != nil {
		return nil, fmt.Errorf("failed to download version: %w", err)
	}

	// Download assets and libraries
	dl := downloader.NewDownloader()
	assetsDownloader := downloader.NewAssetsDownloader(dl, filepath.Join(m.gameDir, "assets"))
	librariesDownloader := downloader.NewLibrariesDownloader(dl, filepath.Join(m.gameDir, "libraries"))

	if err := assetsDownloader.DownloadAssets(versionInfo.AssetIndex, true); err != nil {
		return nil, fmt.Errorf("failed to download assets: %w", err)
	}

	nativesDir := filepath.Join(m.gameDir, "versions", versionID, "natives")
	libraries := downloader.GetLibraries(versionInfo)
	if _, err := librariesDownloader.DownloadLibraries(libraries, nativesDir); err != nil {
		return nil, fmt.Errorf("failed to download libraries: %w", err)
	}

	logger.Info("Version %s installed successfully with all assets and libraries", versionID)
	return versionInfo, nil
}

func (m *Manager) UninstallVersion(versionID string) error {
	logger.Info("Uninstalling version: %s", versionID)

	versionDir := filepath.Join(m.gameDir, "versions", versionID)
	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		return fmt.Errorf("version %s is not installed", versionID)
	}

	if err := os.RemoveAll(versionDir); err != nil {
		return fmt.Errorf("failed to remove version directory: %w", err)
	}

	logger.Info("Version %s uninstalled successfully", versionID)
	return nil
}

func (m *Manager) GetLatestRelease() (string, error) {
	manifest, err := m.GetManifest()
	if err != nil {
		return "", err
	}
	return manifest.Latest.Release, nil
}

func (m *Manager) GetLatestSnapshot() (string, error) {
	manifest, err := m.GetManifest()
	if err != nil {
		return "", err
	}
	return manifest.Latest.Snapshot, nil
}

func (m *Manager) SearchVersions(query string) ([]downloader.VersionEntry, error) {
	manifest, err := m.GetManifest()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	results := make([]downloader.VersionEntry, 0)

	for _, v := range manifest.Versions {
		if strings.Contains(strings.ToLower(v.ID), query) {
			results = append(results, v)
		}
	}

	return results, nil
}
