package template

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func Render(srcDir, dstDir string, vars map[string]string) error {
	return walk(srcDir, dstDir, vars, true)
}

// Copy walks the template tree without running text/template rendering,
// preserving file contents byte-for-byte.
func Copy(srcDir, dstDir string) error {
	return walk(srcDir, dstDir, nil, false)
}

func walk(srcDir, dstDir string, vars map[string]string, render bool) error {
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == srcDir {
			return nil
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if rel == ManifestFilename {
			return nil
		}

		outRel := rel
		if render {
			outRel, err = renderPath(rel, vars)
			if err != nil {
				return fmt.Errorf("render path %q: %w", rel, err)
			}
			if outRel == "" {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		outPath := filepath.Join(dstDir, outRel)

		if d.IsDir() {
			return os.MkdirAll(outPath, 0o755)
		}

		if render {
			return renderFile(path, outPath, vars)
		}
		return copyFile(path, outPath)
	})
}

func copyFile(srcPath, dstPath string) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("create parent dir for %s: %w", dstPath, err)
	}
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", srcPath, err)
	}
	defer func() { _ = src.Close() }()
	info, err := src.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", srcPath, err)
	}
	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return fmt.Errorf("create %s: %w", dstPath, err)
	}
	defer func() { _ = dst.Close() }()
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copy %s: %w", srcPath, err)
	}
	return nil
}

func renderPath(rel string, vars map[string]string) (string, error) {
	parts := strings.Split(rel, string(os.PathSeparator))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		rendered, err := renderString(p, vars)
		if err != nil {
			return "", err
		}
		if rendered == "" {
			return "", nil
		}
		out = append(out, rendered)
	}
	return filepath.Join(out...), nil
}

func renderString(s string, vars map[string]string) (string, error) {
	if !strings.Contains(s, "{{") {
		return s, nil
	}
	tmpl, err := template.New("s").Option("missingkey=zero").Parse(s)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderFile(srcPath, dstPath string, vars map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("create parent dir for %s: %w", dstPath, err)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", srcPath, err)
	}
	defer func() { _ = src.Close() }()

	info, err := src.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", srcPath, err)
	}

	const sniffSize = 8000
	head := make([]byte, sniffSize)
	n, err := io.ReadFull(src, head)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return fmt.Errorf("read %s: %w", srcPath, err)
	}
	head = head[:n]

	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return fmt.Errorf("create %s: %w", dstPath, err)
	}
	defer func() { _ = dst.Close() }()

	if isBinary(head) {
		if _, err := dst.Write(head); err != nil {
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("copy binary %s: %w", srcPath, err)
		}
		return nil
	}

	rest, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("read rest of %s: %w", srcPath, err)
	}
	full := append(head, rest...)

	rendered, err := renderString(string(full), vars)
	if err != nil {
		return fmt.Errorf("render %s: %w", srcPath, err)
	}
	if _, err := dst.Write([]byte(rendered)); err != nil {
		return fmt.Errorf("write %s: %w", dstPath, err)
	}
	return nil
}

func isBinary(b []byte) bool {
	for _, c := range b {
		if c == 0 {
			return true
		}
	}
	return false
}
