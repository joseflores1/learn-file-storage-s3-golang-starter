package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"strings"
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

func getAssetPath(mediaType string) string {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic("couldn't generate random bytes")
	}
	base64Key := base64.RawURLEncoding.EncodeToString(key)

	fileExt := mediaTypeToExt(mediaType)
	return fmt.Sprintf("%s%s", base64Key, fileExt)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetDiskPath string) string {
	return fmt.Sprintf("http://localhost:%s/%s", cfg.port, assetDiskPath)
}

func checkAssetMediaType(mediaType string, allowedTypes map[string]struct{}) error {
	mimeType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return err
	}

	if _, ok := allowedTypes[mimeType]; !ok {
		return errors.New("mime type not allowed")
	}

	return nil
}
