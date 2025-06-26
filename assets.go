package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); errors.Is(err, fs.ErrNotExist) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func mediaTypeToExt(contentType string) string {
	splitContentType := strings.Split(contentType, "/")
	if len(splitContentType) != 2 {
		return ".bin"
	}
	return "." + splitContentType[1]
}

func getAssetPath(videoID uuid.UUID, mediaType string) string {
	fileExt := mediaTypeToExt(mediaType)
	return fmt.Sprintf("%s%s", videoID, fileExt)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetDiskPath string) string {
	return fmt.Sprintf("http://localhost:%s/%s", cfg.port, assetDiskPath)
}
