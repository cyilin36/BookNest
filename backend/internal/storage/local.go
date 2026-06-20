package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"booknest/backend/internal/config"
	"github.com/google/uuid"
)

type Local struct {
	cfg *config.Config
}

type StoredFile struct {
	RelativePath string
	AbsolutePath string
	Hash         string
	Size         int64
}

func NewLocal(cfg *config.Config) *Local {
	return &Local{cfg: cfg}
}

func (s *Local) EnsureDirs() error {
	for _, dir := range []string{
		filepath.Join(s.cfg.BooksDir, "private"),
		filepath.Join(s.cfg.BooksDir, "public"),
		s.cfg.CoversDir,
		s.cfg.AssetsDir,
		s.cfg.TempDir,
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (s *Local) SaveUpload(file multipart.File, header *multipart.FileHeader, visibility, format string) (*StoredFile, error) {
	tmpName := uuid.NewString() + ".upload"
	tmpPath := filepath.Join(s.cfg.TempDir, tmpName)
	out, err := os.Create(tmpPath)
	if err != nil {
		return nil, err
	}
	hasher := sha256.New()
	size, copyErr := io.Copy(io.MultiWriter(out, hasher), file)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return nil, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return nil, closeErr
	}
	hash := hex.EncodeToString(hasher.Sum(nil))
	ext := "." + format
	if originalExt := strings.ToLower(filepath.Ext(header.Filename)); originalExt != "" {
		ext = originalExt
	}
	name := hash + ext
	if visibility == "private" {
		name = uuid.NewString() + "-" + hash + ext
	}
	rel := filepath.Join(visibility, hash[:2], name)
	abs := filepath.Join(s.cfg.BooksDir, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		_ = os.Remove(tmpPath)
		return nil, err
	}
	if err := moveFile(tmpPath, abs); err != nil {
		_ = os.Remove(tmpPath)
		return nil, err
	}
	return &StoredFile{RelativePath: rel, AbsolutePath: abs, Hash: hash, Size: size}, nil
}

func (s *Local) SaveCoverBytes(data []byte, bookID int64, ext string) (string, string, error) {
	if strings.TrimSpace(ext) == "" {
		ext = ".jpg"
	}
	rel := filepath.Join(fmt.Sprintf("%d", bookID), fmt.Sprintf("cover%s", ext))
	abs := filepath.Join(s.cfg.CoversDir, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(abs, data, 0o644); err != nil {
		return "", "", err
	}
	return rel, abs, nil
}

func (s *Local) SaveCoverFile(file multipart.File, ownerType string, ownerID int64, ext string) (string, string, error) {
	if strings.TrimSpace(ext) == "" {
		ext = ".jpg"
	}
	rel := filepath.Join(ownerType, fmt.Sprintf("%d", ownerID), "cover-"+uuid.NewString()+ext)
	abs := filepath.Join(s.cfg.CoversDir, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", "", err
	}
	tmpPath := filepath.Join(s.cfg.TempDir, uuid.NewString()+".cover")
	out, err := os.Create(tmpPath)
	if err != nil {
		return "", "", err
	}
	_, copyErr := io.Copy(out, file)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return "", "", copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return "", "", closeErr
	}
	if err := moveFile(tmpPath, abs); err != nil {
		_ = os.Remove(tmpPath)
		return "", "", err
	}
	return rel, abs, nil
}

func (s *Local) SaveSiteIcon(file multipart.File, ext string) (string, error) {
	if strings.TrimSpace(ext) == "" {
		ext = ".png"
	}
	rel := filepath.Join("site", "icon"+ext)
	abs := filepath.Join(s.cfg.AssetsDir, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", err
	}
	tmpPath := filepath.Join(s.cfg.TempDir, uuid.NewString()+".site-icon")
	out, err := os.Create(tmpPath)
	if err != nil {
		return "", err
	}
	_, copyErr := io.Copy(out, file)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return "", copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return "", closeErr
	}
	if err := moveFile(tmpPath, abs); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	return rel, nil
}

func (s *Local) SaveLoginBackground(file multipart.File, ext string) (string, error) {
	if strings.TrimSpace(ext) == "" {
		ext = ".jpg"
	}
	rel := filepath.Join("site", "login-background"+ext)
	abs := filepath.Join(s.cfg.AssetsDir, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", err
	}
	tmpPath := filepath.Join(s.cfg.TempDir, uuid.NewString()+".login-bg")
	out, err := os.Create(tmpPath)
	if err != nil {
		return "", err
	}
	_, copyErr := io.Copy(out, file)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return "", copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return "", closeErr
	}
	if err := moveFile(tmpPath, abs); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	return rel, nil
}

func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Remove(src)
}

func (s *Local) BookPath(relative string) (string, error) {
	return safeJoin(s.cfg.BooksDir, relative)
}

func (s *Local) CoverPath(relative string) (string, error) {
	return safeJoin(s.cfg.CoversDir, relative)
}

func (s *Local) AssetPath(relative string) (string, error) {
	return safeJoin(s.cfg.AssetsDir, relative)
}

func safeJoin(root, relative string) (string, error) {
	if strings.TrimSpace(relative) == "" {
		return "", fmt.Errorf("empty path")
	}
	clean := filepath.Clean(relative)
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("unsafe path")
	}
	path := filepath.Join(root, clean)
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if pathAbs != rootAbs && !strings.HasPrefix(pathAbs, rootAbs+string(os.PathSeparator)) {
		return "", fmt.Errorf("unsafe path")
	}
	return pathAbs, nil
}
