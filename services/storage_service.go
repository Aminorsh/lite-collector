package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"lite-collector/utils"
)

// StorageService persists uploaded files to a local directory and returns
// URL paths that the app serves via a static route. A MinIO/OSS-backed
// implementation can replace this by satisfying the same method signature.
type StorageService struct {
	baseDir    string // on-disk root, e.g. "./data"
	publicBase string // URL prefix clients should GET, e.g. "/static"
}

// NewStorageService returns a storage service rooted at baseDir; files
// are served to clients under publicBase + the stored relative path.
func NewStorageService(baseDir, publicBase string) *StorageService {
	return &StorageService{baseDir: baseDir, publicBase: publicBase}
}

// allowed image content types for avatar upload.
var allowedAvatarTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// UploadAvatar writes the reader's bytes to avatars/{userID}-{ts}{ext}
// and returns the public URL. contentType must be a whitelisted MIME type.
func (s *StorageService) UploadAvatar(userID uint64, r io.Reader, contentType string) (string, error) {
	ext, ok := allowedAvatarTypes[contentType]
	if !ok {
		return "", utils.ErrAvatarBadType
	}

	dir := filepath.Join(s.baseDir, "avatars")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", utils.ErrInternal
	}

	name := fmt.Sprintf("%d-%d%s", userID, time.Now().UnixNano(), ext)
	fullPath := filepath.Join(dir, name)
	f, err := os.Create(fullPath)
	if err != nil {
		return "", utils.ErrInternal
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", utils.ErrInternal
	}

	return s.publicBase + "/avatars/" + name, nil
}
