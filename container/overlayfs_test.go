package container

import (
	"os"
	"path"
	"testing"
)

func testOverlayfsConfig() *Container {
	return &Container{
		ID:        "overlayfstest",
		WorkDir:   "/home/foo/writingadocker/hind/overlayfstest",
		ImagePath: "/home/foo/writingadocker/hind/images/hello-world.tar",
		Overlay:   true,
	}
}

// Test_makeOverlayFS
//   - makeOverlayFS 曾用名 initOverlayFS, here we keep the old test name for compatibility.
//
// REQUIREMENT:
//   - ./images/hello-world.tar should exist
//   - run as root
//
// To make sure the test won't mess up the host system,
// Use hind to run this test in a container:
//
//	$ sudo go run . run -ti / sh -c "cd /home/foo/writingadocker/hind; go test -v -run Test_initOverlayFS ./container && go test -v -run Test_destroyOverlayFS ./container"
//
// Yes, hind is developed with the assistance of hind itself.
//
// Use Test_destroyOverlayFS to clean up this test.
func Test_initOverlayFS(t *testing.T) {
	config := testOverlayfsConfig()

	err := makeOverlayFS(config)
	if err != nil {
		t.Errorf("initOverlayFS() error = %v", err)
	}

	// Try to write to the overlayfs (merged directory)
	// and check if the file is created in the upper directory.

	fileName := "test.txt"
	testContent := []byte("hello")

	if err := os.WriteFile(path.Join(config.overlayMergedDir(), fileName), testContent, 0644); err != nil {
		t.Errorf("WriteFile in merged overlay fs error: %v", err)
	}

	if got, err := os.ReadFile(path.Join(config.overlayUpperDir(), fileName)); err != nil {
		t.Errorf("ReadFile in writable layer error: %v", err)
	} else if string(got) != string(testContent) {
		t.Errorf("ReadFile in writable layer got = %v, want %v", got, testContent)
	}
}

func Test_destroyOverlayFS(t *testing.T) {
	config := testOverlayfsConfig()

	err := destroyOverlayFS(config)
	if err != nil {
		t.Errorf("cleanupOverlayFS() error = %v", err)
	}
}

func TestOverlayFS4ImageIsDir(t *testing.T) {
	config := &Container{
		ID:        "overlayfstest",
		WorkDir:   "/home/foo/writingadocker/hind/overlayfstest",
		ImagePath: "/home/foo/writingadocker/hind/images/hello-world",
		Overlay:   true,
	}

	t.Run("makeOverlayFS", func(t *testing.T) {
		err := makeOverlayFS(config)
		if err != nil {
			t.Errorf("initOverlayFS() error = %v", err)
		}

		// there should not a new created "image" directory:
		if _, err := os.Stat(path.Join(config.overlayRootDir(), "/image")); err == nil {
			t.Errorf("image directory should not exist")
		}
	})

	t.Run("writeFile", func(t *testing.T) {
		fileName := "test.txt"
		testContent := []byte("hello")

		if err := os.WriteFile(path.Join(config.overlayMergedDir(), fileName), testContent, 0644); err != nil {
			t.Errorf("WriteFile in merged overlay fs error: %v", err)
		}

		// Check if the file is created in the upper directory.
		if got, err := os.ReadFile(path.Join(config.overlayUpperDir(), fileName)); err != nil {
			t.Errorf("ReadFile in writable layer error: %v", err)
		} else if string(got) != string(testContent) {
			t.Errorf("ReadFile in writable layer got = %v, want %v", got, testContent)
		}

		// Check if the file is NOT touch in the lower directory.
		if _, err := os.ReadFile(path.Join(config.overlayLowerDir(), fileName)); err == nil {
			t.Errorf("ReadFile in lower layer should error, but got nil")
		}
	})

	t.Run("destroyOverlayFS", func(t *testing.T) {
		err := destroyOverlayFS(config)
		if err != nil {
			t.Errorf("cleanupOverlayFS() error = %v", err)
		}
	})
}
