package install

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tars/internal/formula"
	"tars/internal/security"
)

const downloadTimeout = 30 * time.Minute

// Download fetches url to destPath and verifies SHA256 before returning.
func Download(url, destPath, expectedSHA256 string) error {
	client := &http.Client{Timeout: downloadTimeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "tars/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: HTTP %s", resp.Status)
	}
	tmp := destPath + ".part"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := security.VerifySHA256(tmp, expectedSHA256); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("security: %w", err)
	}
	return os.Rename(tmp, destPath)
}

// Install extracts or copies artifact into prefix and returns install root for the version.
func Install(f *formula.Formula, artifactPath, prefix string) (string, error) {
	verDir := filepath.Join(prefix, f.Name, f.Version)
	if err := os.MkdirAll(verDir, 0o755); err != nil {
		return "", err
	}

	lower := strings.ToLower(artifactPath)
	switch {
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		if err := extractTarGz(artifactPath, verDir); err != nil {
			return "", err
		}
	case strings.HasSuffix(lower, ".zip"):
		if err := extractZip(artifactPath, verDir); err != nil {
			return "", err
		}
	default:
		// Single binary
		base := filepath.Base(f.URL)
		if base == "" || base == "/" {
			base = f.Name
		}
		dst := filepath.Join(verDir, base)
		if err := copyFile(artifactPath, dst); err != nil {
			return "", err
		}
		if err := os.Chmod(dst, 0o755); err != nil {
			return "", err
		}
	}

	return verDir, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func extractTarGz(path, dest string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag == tar.TypeDir {
			continue
		}
		name := hdr.Name
		// strip single top-level directory if present
		parts := strings.SplitN(filepath.Clean(name), string(filepath.Separator), 2)
		if len(parts) == 2 {
			name = parts[1]
		}
		if name == "" || name == "." {
			continue
		}
		target := filepath.Join(dest, name)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			continue
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		if err := out.Close(); err != nil {
			return err
		}
	}
	return nil
}

func extractZip(path, dest string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, zf := range r.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		name := zf.Name
		parts := strings.SplitN(filepath.Clean(name), string(filepath.Separator), 2)
		if len(parts) == 2 {
			name = parts[1]
		}
		if name == "" {
			continue
		}
		target := filepath.Join(dest, name)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, zf.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		_, copyErr := io.Copy(out, rc)
		closeErr := out.Close()
		rc.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
