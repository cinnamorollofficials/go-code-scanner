package release

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ArchiveBinary(binaryPath, archivePath string, timestamp time.Time) error {
	if timestamp.IsZero() {
		return fmt.Errorf("archive timestamp is required")
	}
	info, err := os.Lstat(binaryPath)
	if err != nil {
		return fmt.Errorf("inspect binary: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("binary must be a regular file")
	}
	input, err := os.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("open binary: %w", err)
	}
	defer input.Close()

	temporary, err := os.CreateTemp(filepath.Dir(archivePath), ".release-archive-*")
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	name := filepath.Base(binaryPath)
	stamp := timestamp.UTC()
	if strings.HasSuffix(archivePath, ".zip") {
		err = writeZIP(temporary, input, name, stamp, info.Mode().Perm())
	} else if strings.HasSuffix(archivePath, ".tar.gz") {
		err = writeTarGzip(temporary, input, name, stamp, info.Mode().Perm())
	} else {
		err = fmt.Errorf("archive path must end in .zip or .tar.gz")
	}
	if err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close archive: %w", err)
	}
	if err := os.Chmod(temporaryPath, 0o644); err != nil {
		return fmt.Errorf("set archive permissions: %w", err)
	}
	return os.Rename(temporaryPath, archivePath)
}

func writeZIP(output io.Writer, input io.Reader, name string, stamp time.Time, mode os.FileMode) error {
	archive := zip.NewWriter(output)
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	header.SetMode(mode)
	header.SetModTime(stamp)
	entry, err := archive.CreateHeader(header)
	if err == nil {
		_, err = io.Copy(entry, input)
	}
	if closeErr := archive.Close(); err == nil {
		err = closeErr
	}
	return err
}

func writeTarGzip(output io.Writer, input io.Reader, name string, stamp time.Time, mode os.FileMode) error {
	gzipWriter, err := gzip.NewWriterLevel(output, gzip.BestCompression)
	if err != nil {
		return err
	}
	gzipWriter.Header.ModTime = stamp
	gzipWriter.Header.OS = 255
	tarWriter := tar.NewWriter(gzipWriter)
	data, err := io.ReadAll(input)
	if err == nil {
		err = tarWriter.WriteHeader(&tar.Header{Name: name, Mode: int64(mode.Perm()), Size: int64(len(data)), ModTime: stamp, Format: tar.FormatPAX})
	}
	if err == nil {
		_, err = tarWriter.Write(data)
	}
	if closeErr := tarWriter.Close(); err == nil {
		err = closeErr
	}
	if closeErr := gzipWriter.Close(); err == nil {
		err = closeErr
	}
	return err
}
