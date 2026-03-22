package commands

import "testing"

// TestMaxManifestSize_Exists verifies the constant is defined and reasonable.
func TestMaxManifestSize_Exists(t *testing.T) {
	if maxManifestSize <= 0 {
		t.Error("maxManifestSize must be positive")
	}
	if maxManifestSize > 20*1024*1024 {
		t.Errorf("maxManifestSize = %d, suspiciously large (> 20MB)", maxManifestSize)
	}
}

// TestMaxConvertFileSize_Exists verifies the convert size constant.
func TestMaxConvertFileSize_Exists(t *testing.T) {
	if maxConvertFileSize <= 0 {
		t.Error("maxConvertFileSize must be positive")
	}
}
