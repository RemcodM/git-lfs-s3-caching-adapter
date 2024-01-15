package stats

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/lfs"
)

func ByteCountIEC(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func Percentage(fraction uint64, total uint64) string {
	if total == 0 {
		return fmt.Sprintf("%5.1f%%", 0.0)
	}
	return fmt.Sprintf("%5.1f%%", float64(fraction*100)/float64(total))
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func allStatsFiles() ([]string, error) {
	cacheStoreDir, err := requireCacheStoreDirectory()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(cacheStoreDir)
	if err != nil {
		return nil, err
	}

	var filenames []string

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			filenames = append(filenames, file.Name())
		}
	}

	return filenames, nil
}

func requireCacheStoreDirectory() (string, error) {
	cacheStoreDir, err := cacheStoreDirectory()
	if err != nil {
		return cacheStoreDir, err
	}

	err = os.MkdirAll(cacheStoreDir, 0755)
	if err != nil {
		return cacheStoreDir, err
	}
	return cacheStoreDir, nil
}

func cacheStoreDirectory() (string, error) {
	config := lfs.GetPassthroughConfiguration()
	if !config.InRepo() {
		return "", fmt.Errorf("not in a git repository")
	}
	return fmt.Sprintf("%s/%s", config.LFSStorageDir(), "cache_stats"), nil
}
