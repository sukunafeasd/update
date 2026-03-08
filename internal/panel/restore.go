package panel

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"universald/internal/model"
)

func (s *Service) RestoreBackupArchive(r io.Reader, source string) error {
	if s == nil || s.store == nil {
		return fmt.Errorf("panel service indisponivel")
	}
	if r == nil {
		return fmt.Errorf("backup de entrada vazio")
	}

	tempDir, err := os.MkdirTemp("", "painel-dief-import-*")
	if err != nil {
		return fmt.Errorf("create import temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, "import.zip")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create import archive: %w", err)
	}
	if _, err := io.Copy(archiveFile, r); err != nil {
		_ = archiveFile.Close()
		return fmt.Errorf("store import archive: %w", err)
	}
	if err := archiveFile.Close(); err != nil {
		return fmt.Errorf("close import archive: %w", err)
	}

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open import zip: %w", err)
	}
	defer reader.Close()

	snapshotPath := filepath.Join(tempDir, "universald.snapshot.db")
	extractedUploads := filepath.Join(tempDir, "panel_uploads")
	for _, file := range reader.File {
		name := filepath.ToSlash(strings.TrimSpace(file.Name))
		switch {
		case name == "universald.snapshot.db":
			if err := extractZipFile(file, snapshotPath); err != nil {
				return err
			}
		case strings.HasPrefix(name, "panel_uploads/"):
			dest := filepath.Join(tempDir, filepath.FromSlash(name))
			if err := extractZipFile(file, dest); err != nil {
				return err
			}
		}
	}

	if _, err := os.Stat(snapshotPath); err != nil {
		return fmt.Errorf("snapshot do backup nao encontrado: %w", err)
	}
	if err := s.store.RestoreFromSnapshot(snapshotPath); err != nil {
		return err
	}
	if err := syncUploadsDir(extractedUploads, s.uploadsDir); err != nil {
		return err
	}
	s.resetRuntimeCaches()
	if err := s.EnsureBootstrapped(); err != nil {
		return err
	}
	s.logGuestAction("ops_import", strings.TrimSpace(source), 0, "", "restaurou backup remoto na instancia atual")
	s.bumpVersion()
	return nil
}

func extractZipFile(file *zip.File, dest string) error {
	handle, err := file.Open()
	if err != nil {
		return fmt.Errorf("open zip file %s: %w", file.Name, err)
	}
	defer handle.Close()

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create extract dir: %w", err)
	}
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create extracted file: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, handle); err != nil {
		return fmt.Errorf("extract file %s: %w", file.Name, err)
	}
	return nil
}

func syncUploadsDir(source, target string) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return fmt.Errorf("uploads target vazio")
	}
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("reset current uploads: %w", err)
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return fmt.Errorf("create uploads dir: %w", err)
	}
	if strings.TrimSpace(source) == "" {
		return nil
	}
	if _, err := os.Stat(source); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat extracted uploads: %w", err)
	}
	return filepath.Walk(source, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info == nil || info.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(target, relative)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, raw, 0o644)
	})
}

func (s *Service) resetRuntimeCaches() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unlocked = map[string]map[int64]time.Time{}
	s.typing = map[int64]map[int64]model.PanelTyping{}
	s.flood = map[int64][]time.Time{}
	s.loginAttempts = map[string]loginAttemptState{}
	s.roomVer = map[int64]int64{}
	s.version++
}
