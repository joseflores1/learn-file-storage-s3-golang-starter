package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"mime"
	"os"
	"os/exec"
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

func (cfg apiConfig) generateS3Path(key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)
}

func getVideoAspectRatio(filePath string) (string, error) {
	const cmdName = "ffprobe"
	args := []string{"-v", "error", "-print_format", "json", "-show_streams", filePath}
	cmd := exec.Command(cmdName, args...)

	var b bytes.Buffer
	cmd.Stdout = &b
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ffprobe error: %v", err)
	}

	data := b.Bytes()

	var dimStruct struct {
		Streams []struct {
			Width  int `json:"width,omitempty"`
			Height int `json:"height,omitempty"`
		} `json:"streams"`
	}

	err = json.Unmarshal(data, &dimStruct)
	if err != nil {
		return "", fmt.Errorf("could not parse ffprobe output: %v", err)
	}

	if len(dimStruct.Streams) == 0 {
		return "", errors.New("no video streams found")
	}
	const epsilon = 1e-2
	const landscape = float64(16) / 9
	const portrait = float64(9) / 16

	ratio := float64(dimStruct.Streams[0].Width) / float64(dimStruct.Streams[0].Height)
	if withinTolerance(ratio, landscape, epsilon) {
		return "16:9", nil
	}
	if withinTolerance(ratio, portrait, epsilon) {
		return "9:16", nil
	}
	return "other", nil
}

func nameAspectRatio(ratio string) string {
	nameMap := map[string]string{
		"16:9":  "landscape/",
		"9:16":  "portrait/",
		"other": "other/",
	}
	return nameMap[ratio]
}

func withinTolerance(a, b, e float64) bool {
	if a == b {
		return true
	}

	d := math.Abs(a - b)

	if b == 0 {
		return d < e
	}

	return (d / math.Abs(b)) < e
}
