package utils

import (
	"strings"
	"testing"
)

func TestSHA256FromDLJSON(t *testing.T) {
	payload := `[
	  {
	    "version": "go1.27.3",
	    "stable": true,
	    "files": [
	      {
	        "filename": "go1.27.3.linux-amd64.tar.gz",
	        "sha256": "65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40",
	        "kind": "archive"
	      },
	      {
	        "filename": "go1.27.3.src.tar.gz",
	        "sha256": "6e7749035cf360f06b2575d162ea79df9fda8ef671889e5c375f8e50777f26bb",
	        "kind": "source"
	      }
	    ]
	  }
	]`

	got, err := sha256FromDLJSON([]byte(payload), "go1.27.3.linux-amd64.tar.gz")
	if err != nil {
		t.Fatalf("sha256FromDLJSON: %v", err)
	}
	want := "65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	_, err = sha256FromDLJSON([]byte(payload), "missing.tar.gz")
	if err == nil || !strings.Contains(err.Error(), "no SHA256 found") {
		t.Fatalf("expected missing-file error, got %v", err)
	}
}
