package modloader

import (
	"fmt"
)

type ModLoaderType string

const (
	ModLoaderNone    ModLoaderType = "none"
	ModLoaderForge   ModLoaderType = "forge"
	ModLoaderFabric  ModLoaderType = "fabric"
	ModLoaderQuilt   ModLoaderType = "quilt"
	ModLoaderNeoForge ModLoaderType = "neoforge"
)

type ModLoader interface {
	GetType() ModLoaderType
	GetVersion() string
	GetID() string
	Install(version string) error
	Uninstall(version string) error
	IsInstalled(version string) bool
	GetLibraries(version string) ([]string, error)
	GetLaunchArguments(version string) ([]string, error)
}

type ModLoaderConfig struct {
	Type    ModLoaderType
	Version string
}

type ModLoaderManager struct {
	loaders map[ModLoaderType]ModLoader
}

func NewModLoaderManager() *ModLoaderManager {
	return &ModLoaderManager{
		loaders: make(map[ModLoaderType]ModLoader),
	}
}

func (m *ModLoaderManager) RegisterLoader(loader ModLoader) {
	m.loaders[loader.GetType()] = loader
}

func (m *ModLoaderManager) GetLoader(loaderType ModLoaderType) (ModLoader, error) {
	loader, ok := m.loaders[loaderType]
	if !ok {
		return nil, fmt.Errorf("mod loader type %s is not supported", loaderType)
	}
	return loader, nil
}

func (m *ModLoaderManager) GetSupportedLoaders() []ModLoaderType {
	types := make([]ModLoaderType, 0, len(m.loaders))
	for t := range m.loaders {
		types = append(types, t)
	}
	return types
}

func (m *ModLoaderManager) IsLoaderSupported(loaderType ModLoaderType) bool {
	_, ok := m.loaders[loaderType]
	return ok
}

type BaseModLoader struct {
	loaderType ModLoaderType
	version    string
}

func NewBaseModLoader(loaderType ModLoaderType, version string) *BaseModLoader {
	return &BaseModLoader{
		loaderType: loaderType,
		version:    version,
	}
}

func (b *BaseModLoader) GetType() ModLoaderType {
	return b.loaderType
}

func (b *BaseModLoader) GetVersion() string {
	return b.version
}

func (b *BaseModLoader) GetID() string {
	return fmt.Sprintf("%s-%s", b.loaderType, b.version)
}

func (b *BaseModLoader) Install(version string) error {
	return fmt.Errorf("mod loader installation not implemented")
}

func (b *BaseModLoader) Uninstall(version string) error {
	return fmt.Errorf("mod loader uninstallation not implemented")
}

func (b *BaseModLoader) IsInstalled(version string) bool {
	return false
}

func (b *BaseModLoader) GetLibraries(version string) ([]string, error) {
	return []string{}, nil
}

func (b *BaseModLoader) GetLaunchArguments(version string) ([]string, error) {
	return []string{}, nil
}
