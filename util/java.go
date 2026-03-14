package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type JavaInfo struct {
	Path    string
	Version string
	Major   int
	Arch    string
}

func FindJava() ([]JavaInfo, error) {
	javaInfos := make([]JavaInfo, 0)

	paths := getJavaSearchPaths()
	for _, path := range paths {
		javaPath := filepath.Join(path, "bin", "java"+getExecutableExt())
		if info, err := getJavaInfo(javaPath); err == nil {
			javaInfos = append(javaInfos, info)
		}
	}

	javaPath, err := exec.LookPath("java")
	if err == nil {
		if info, err := getJavaInfo(javaPath); err == nil {
			javaInfos = append(javaInfos, info)
		}
	}

	return javaInfos, nil
}

func getJavaInfo(javaPath string) (JavaInfo, error) {
	if _, err := os.Stat(javaPath); os.IsNotExist(err) {
		return JavaInfo{}, fmt.Errorf("java not found at %s", javaPath)
	}

	cmd := exec.Command(javaPath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return JavaInfo{}, fmt.Errorf("failed to get java version: %w", err)
	}

	versionStr := parseJavaVersion(string(output))
	majorVersion := parseMajorVersion(versionStr)

	return JavaInfo{
		Path:    javaPath,
		Version: versionStr,
		Major:   majorVersion,
		Arch:    runtime.GOARCH,
	}, nil
}

func parseJavaVersion(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "version") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if strings.Contains(part, "version") && i+1 < len(parts) {
					version := strings.Trim(parts[i+1], `"`)
					return version
				}
			}
		}
	}
	return "unknown"
}

func parseMajorVersion(version string) int {
	version = strings.TrimPrefix(version, "1.")
	version = strings.TrimPrefix(version, "2.")

	parts := strings.Split(version, ".")
	if len(parts) > 0 {
		var major int
		if _, err := fmt.Sscanf(parts[0], "%d", &major); err == nil {
			return major
		}
	}

	if strings.HasPrefix(version, "1.") {
		if len(parts) > 1 {
			var major int
			if _, err := fmt.Sscanf(parts[1], "%d", &major); err == nil {
				return major
			}
		}
	}

	return 8
}

func getJavaSearchPaths() []string {
	paths := make([]string, 0)

	if runtime.GOOS == "windows" {
		programFiles := os.Getenv("ProgramFiles")
		programFilesX86 := os.Getenv("ProgramFiles(x86)")

		if programFiles != "" {
			javaDir := filepath.Join(programFiles, "Java")
			if entries, err := os.ReadDir(javaDir); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						paths = append(paths, filepath.Join(javaDir, entry.Name()))
					}
				}
			}
		}

		if programFilesX86 != "" {
			javaDir := filepath.Join(programFilesX86, "Java")
			if entries, err := os.ReadDir(javaDir); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						paths = append(paths, filepath.Join(javaDir, entry.Name()))
					}
				}
			}
		}

		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			packagesDir := filepath.Join(localAppData, "Packages", "Microsoft.4297127D64EC6_8wekyb3d8bbwe", "LocalCache", "Local", "runtime")
			if entries, err := os.ReadDir(packagesDir); err == nil {
				for _, entry := range entries {
					if strings.Contains(entry.Name(), "java") {
						paths = append(paths, filepath.Join(packagesDir, entry.Name()))
					}
				}
			}
		}
	} else if runtime.GOOS == "darwin" {
		paths = append(paths, "/Library/Java/JavaVirtualMachines")
		homeDir, _ := os.UserHomeDir()
		if homeDir != "" {
			paths = append(paths, filepath.Join(homeDir, "Library", "Java", "JavaVirtualMachines"))
		}
	} else {
		paths = append(paths, "/usr/lib/jvm", "/usr/java", "/opt/java")
	}

	return paths
}

func getExecutableExt() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func ValidateJavaVersion(javaPath string, minVersion int) error {
	info, err := getJavaInfo(javaPath)
	if err != nil {
		return err
	}

	if info.Major < minVersion {
		return fmt.Errorf("java version %d is too low, minimum required: %d", info.Major, minVersion)
	}

	return nil
}

func FindBestJava(minVersion int) (string, error) {
	javaInfos, err := FindJava()
	if err != nil {
		return "", err
	}

	for _, info := range javaInfos {
		if info.Major >= minVersion {
			return info.Path, nil
		}
	}

	return "", fmt.Errorf("no suitable java version found (minimum: %d)", minVersion)
}
