package downloader

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"GilvaaLauncher/logger"
)

type LibrariesDownloader struct {
	downloader   *Downloader
	librariesDir string
}

func NewLibrariesDownloader(downloader *Downloader, librariesDir string) *LibrariesDownloader {
	return &LibrariesDownloader{
		downloader:   downloader,
		librariesDir: librariesDir,
	}
}

func (ld *LibrariesDownloader) DownloadLibraries(libraries []Library, nativesDir string) ([]string, error) {
	classpath := []string{}
	downloaded := 0
	skipped := 0

	for _, lib := range libraries {
		if !ld.shouldDownloadLibrary(lib) {
			continue
		}

		if lib.Natives != nil {
			if err := ld.downloadNativeLibrary(lib, nativesDir); err != nil {
				logger.Warn("Failed to download native library %s: %v", lib.Name, err)
				continue
			}
			downloaded++
		} else {
			artifact, ok := lib.Downloads["artifact"]
			if !ok {
				logger.Warn("No artifact found for library %s", lib.Name)
				continue
			}

			libPath := filepath.Join(ld.librariesDir, artifact.Path)
			if _, err := os.Stat(libPath); err == nil {
				if ld.downloader.verifyHash(libPath, artifact.SHA1) {
					classpath = append(classpath, libPath)
					skipped++
					continue
				}
			}

			if err := ld.downloader.DownloadFile(artifact.URL, libPath, artifact.SHA1); err != nil {
				logger.Warn("Failed to download library %s: %v", lib.Name, err)
				continue
			}

			classpath = append(classpath, libPath)
			downloaded++
		}

		if downloaded%10 == 0 {
			logger.Info("Downloaded %d libraries, skipped %d", downloaded, skipped)
		}
	}

	logger.Info("Libraries download complete: %d downloaded, %d skipped", downloaded, skipped)
	return classpath, nil
}

func (ld *LibrariesDownloader) shouldDownloadLibrary(lib Library) bool {
	if len(lib.Rules) == 0 {
		return true
	}

	allowed := false
	for _, rule := range lib.Rules {
		if ld.matchesRule(rule) {
			allowed = rule.Action == "allow"
		}
	}

	return allowed
}

func (ld *LibrariesDownloader) matchesRule(rule Rule) bool {
	if rule.OS == nil {
		return true
	}

	if rule.OS.Name != "" {
		osName := runtime.GOOS
		if osName == "windows" {
			osName = "windows"
		} else if osName == "darwin" {
			osName = "osx"
		}

		if rule.OS.Name != osName {
			return false
		}
	}

	if rule.OS.Arch != "" {
		arch := runtime.GOARCH
		if rule.OS.Arch != arch {
			return false
		}
	}

	return true
}

func (ld *LibrariesDownloader) downloadNativeLibrary(lib Library, nativesDir string) error {
	osName := runtime.GOOS
	if osName == "windows" {
		osName = "windows"
	} else if osName == "darwin" {
		osName = "osx"
	} else {
		osName = "linux"
	}

	nativeClassifier, ok := lib.Natives[osName]
	if !ok {
		return fmt.Errorf("no native classifier for OS %s", osName)
	}

	nativeClassifier = strings.ReplaceAll(nativeClassifier, "${arch}", runtime.GOARCH)

	classifierKey := "natives-" + nativeClassifier
	native, ok := lib.Downloads[classifierKey]
	if !ok {
		return fmt.Errorf("no native download found for classifier %s", classifierKey)
	}

	libPath := filepath.Join(ld.librariesDir, native.Path)
	if _, err := os.Stat(libPath); err == nil {
		if ld.downloader.verifyHash(libPath, native.SHA1) {
			return ld.extractNativeLibrary(libPath, nativesDir)
		}
	}

	if err := ld.downloader.DownloadFile(native.URL, libPath, native.SHA1); err != nil {
		return err
	}

	return ld.extractNativeLibrary(libPath, nativesDir)
}

func (ld *LibrariesDownloader) extractNativeLibrary(libPath, nativesDir string) error {
	if err := os.MkdirAll(nativesDir, 0755); err != nil {
		return fmt.Errorf("failed to create natives directory: %w", err)
	}

	logger.Info("Extracting native library: %s", filepath.Base(libPath))

	reader, err := zip.OpenReader(libPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	extracted := 0
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name))
		if ext != ".dll" && ext != ".so" && ext != ".dylib" {
			continue
		}

		destPath := filepath.Join(nativesDir, filepath.Base(file.Name))

		srcFile, err := file.Open()
		if err != nil {
			logger.Warn("Failed to open file in zip: %v", err)
			continue
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			srcFile.Close()
			logger.Warn("Failed to create native file: %v", err)
			continue
		}

		if _, err := io.Copy(destFile, srcFile); err != nil {
			srcFile.Close()
			destFile.Close()
			logger.Warn("Failed to extract native file: %v", err)
			continue
		}

		srcFile.Close()
		destFile.Close()
		extracted++
	}

	logger.Info("Extracted %d native libraries", extracted)
	return nil
}
