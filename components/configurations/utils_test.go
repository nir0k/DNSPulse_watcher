package config

import (
	"crypto/md5"
	"fmt"
	"os"
	"testing"
)

func TestCalculateHash(t *testing.T) {
	// Create a temporary file
	content := []byte("test content")
	tmpfile, err := os.CreateTemp("", "example.*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Read the file content back for debugging
	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("File content: %s\n", string(data))

	// Test CalculateHash
	got, err := CalculateHash(tmpfile.Name(), md5.New)
	if err != nil {
		t.Errorf("CalculateHash() error = %v", err)
		return
	}

	want := "9473fdd0d880a43c21b7778d34872157"
	if got != want {
		t.Errorf("CalculateHash() = %v, want %v", got, want)
	}
}

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"Valid Path", "/usr/local/bin", true},
		{"Invalid Path", "c:\\windows\\system32", false},
		{"Empty Path", "", false},
		// Add more test cases if necessary
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateFilePath(tt.path); got != tt.want {
				t.Errorf("ValidateFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLocalIP(t *testing.T) {
	got := GetLocalIP()
	if got == "" {
		t.Error("GetLocalIP() returned an empty string")
	}

	// Add more robust tests if necessary, such as checking for valid IP format
}