package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// CacheFileName is used for storing downloaded emojis
	CacheFileName = "emoji.json"
	// IndexFileName is used for storing the emoji index
	IndexFileName = "emos.index"
	// ImageCacheDir is used to cache emoji images
	ImageCacheDir = "imgs"
)

// Loc returns the expected location of the config file
func Loc(name string) string {
	return filepath.Join(emosDir(), name)
}

// Dir provides the base config location
func Dir() string {
	return emosDir()
}

func emosDir() string {
	cfg, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("failed to get config dir")
		panic(err)
	}

	return filepath.Join(cfg, "emos")
}
