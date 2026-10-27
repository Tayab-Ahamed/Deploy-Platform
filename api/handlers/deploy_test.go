package handlers

import (
	"archive/zip"
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testArchive(t *testing.T, name string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "source.zip")
	f, err := os.Create(file)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	entry, err := w.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.WriteString(entry, "source"); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestUnzipRejectsTraversalInsteadOfSilentlyDroppingFiles(t *testing.T) {
	for _, name := range []string{"../escape", "/absolute", `..\escape`} {
		t.Run(name, func(t *testing.T) {
			if err := unzip(testArchive(t, name), t.TempDir()); err == nil {
				t.Fatal("expected unsafe archive path to be rejected")
			}
		})
	}
}

func TestUnzipPreservesLocalDotPrefixedNames(t *testing.T) {
	dest := t.TempDir()
	if err := unzip(testArchive(t, "..config"), dest); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(dest, "..config"))
	if err != nil || string(content) != "source" {
		t.Fatalf("local source file lost: %q, %v", content, err)
	}
}

func TestDeployRejectsBodyAboveUploadLimit(t *testing.T) {
	var prefix bytes.Buffer
	w := multipart.NewWriter(&prefix)
	if _, err := w.CreateFormFile("source", "source.zip"); err != nil {
		t.Fatal(err)
	}
	body := io.MultiReader(&prefix, io.LimitReader(zeroReader{}, (64<<20)+1), strings.NewReader("\r\n--"+w.Boundary()+"--\r\n"))
	req := httptest.NewRequest(http.MethodPost, "/deploy", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	recorder := httptest.NewRecorder()
	DeployHandler(recorder, req)
	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	clear(p)
	return len(p), nil
}
