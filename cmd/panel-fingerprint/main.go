package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"universald/internal/db"
)

func main() {
	var dbPath string
	var uploadsDir string

	flag.StringVar(&dbPath, "db", "", "caminho do sqlite snapshot")
	flag.StringVar(&uploadsDir, "uploads", "", "diretorio de uploads restaurado")
	flag.Parse()

	dbPath = strings.TrimSpace(dbPath)
	if dbPath == "" {
		fatalf("db obrigatorio")
	}

	store, err := db.Open(dbPath)
	if err != nil {
		fatalf("abrir db: %v", err)
	}
	defer store.Close()

	dbFingerprint, err := store.PanelContentFingerprint()
	if err != nil {
		fatalf("fingerprint db: %v", err)
	}
	uploadDigest, err := dirFingerprint(strings.TrimSpace(uploadsDir))
	if err != nil {
		fatalf("fingerprint uploads: %v", err)
	}
	fmt.Print(hashStrings(dbFingerprint, uploadDigest))
}

func dirFingerprint(root string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return hashStrings("uploads-empty"), nil
	}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return hashStrings("uploads-missing"), nil
		}
		return "", fmt.Errorf("stat uploads dir: %w", err)
	}

	files := make([]string, 0, 16)
	if err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info == nil || info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return "", fmt.Errorf("walk uploads dir: %w", err)
	}
	sort.Strings(files)

	hasher := sha256.New()
	for _, path := range files {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return "", fmt.Errorf("relative path: %w", err)
		}
		info, err := os.Stat(path)
		if err != nil {
			return "", fmt.Errorf("stat file: %w", err)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read file: %w", err)
		}
		_, _ = hasher.Write([]byte(filepath.ToSlash(relative)))
		_, _ = hasher.Write([]byte{0})
		_, _ = hasher.Write([]byte(strconv.FormatInt(info.Size(), 10)))
		_, _ = hasher.Write([]byte{0})
		sum := sha256.Sum256(raw)
		_, _ = hasher.Write([]byte(hex.EncodeToString(sum[:])))
		_, _ = hasher.Write([]byte{0xff})
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func hashStrings(parts ...string) string {
	hasher := sha256.New()
	for _, part := range parts {
		_, _ = hasher.Write([]byte(strings.TrimSpace(part)))
		_, _ = hasher.Write([]byte{0xff})
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
