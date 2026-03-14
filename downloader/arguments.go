package downloader

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ParseArguments(args []interface{}) []string {
	result := make([]string, 0)

	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			result = append(result, v)
		case []interface{}:
			for _, item := range v {
				switch itemVal := item.(type) {
				case string:
					result = append(result, itemVal)
				case map[string]interface{}:
					if parsed := parseArgumentRule(itemVal); parsed != nil {
						result = append(result, parsed...)
					}
				}
			}
		case map[string]interface{}:
			if parsed := parseArgumentRule(v); parsed != nil {
				result = append(result, parsed...)
			}
		}
	}

	return result
}

func parseArgumentRule(rule map[string]interface{}) []string {
	action, ok := rule["action"].(string)
	if !ok {
		return nil
	}

	if action != "allow" {
		return nil
	}

	value, ok := rule["value"]
	if !ok {
		return nil
	}

	return parseValue(value)
}

func parseValue(value interface{}) []string {
	switch v := value.(type) {
	case string:
		return []string{v}
	case []interface{}:
		result := make([]string, 0)
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		data, _ := json.Marshal(value)
		var str string
		if err := json.Unmarshal(data, &str); err == nil {
			return []string{str}
		}
		var strs []string
		if err := json.Unmarshal(data, &strs); err == nil {
			return strs
		}
		return nil
	}
}

func ParseMinecraftArguments(argsStr string) []string {
	if argsStr == "" {
		return []string{}
	}

	result := make([]string, 0)
	parts := strings.Fields(argsStr)

	for _, part := range parts {
		if strings.HasPrefix(part, "${") && strings.HasSuffix(part, "}") {
			result = append(result, part)
		} else {
			result = append(result, part)
		}
	}

	return result
}

func GetGameArguments(versionInfo *VersionInfo) []string {
	if versionInfo.Arguments != nil && len(versionInfo.Arguments.Game) > 0 {
		return ParseArguments(versionInfo.Arguments.Game)
	}

	if versionInfo.MinecraftArguments != "" {
		return ParseMinecraftArguments(versionInfo.MinecraftArguments)
	}

	return []string{}
}

func GetJVMArguments(versionInfo *VersionInfo) []string {
	if versionInfo.Arguments != nil && len(versionInfo.Arguments.JVM) > 0 {
		return ParseArguments(versionInfo.Arguments.JVM)
	}

	return []string{}
}

func GetMainClass(versionInfo *VersionInfo) string {
	if versionInfo.MainClass != "" {
		return versionInfo.MainClass
	}

	return "net.minecraft.client.main.Main"
}

func GetAssetIndex(versionInfo *VersionInfo) string {
	if versionInfo.AssetIndex.ID != "" {
		return versionInfo.AssetIndex.ID
	}

	if versionInfo.Assets != "" {
		return versionInfo.Assets
	}

	return "legacy"
}

func GetJavaVersion(versionInfo *VersionInfo) int {
	if versionInfo.JavaVersion != nil {
		return versionInfo.JavaVersion.MajorVersion
	}

	return 8
}

func GetLibraries(versionInfo *VersionInfo) []Library {
	if versionInfo.Libraries != nil {
		return versionInfo.Libraries
	}

	return []Library{}
}

func GetClientDownload(versionInfo *VersionInfo) (*FileDownload, error) {
	if versionInfo.Downloads != nil {
		if client, ok := versionInfo.Downloads["client"]; ok {
			return &client, nil
		}
	}

	return nil, fmt.Errorf("client download not found")
}
