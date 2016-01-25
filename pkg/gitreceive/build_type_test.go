package gitreceive

import (
	"os"
	"testing"
)

func TestGetBuildTypeForDir(t *testing.T) {
	tmpDir := os.TempDir()
	bType := getBuildTypeForDir(tmpDir)
	if bType != buildTypeProcfile {
		t.Fatalf("expected procfile build, got %s", bType)
	}
	if _, err := os.Create(tmpDir + "/Dockerfile"); err != nil {
		t.Fatalf("error creating %s/Dockerfile (%s)", tmpDir, err)
	}
	defer func() {
		if err := os.Remove(tmpDir + "/Dockerfile"); err != nil {
			t.Fatalf("failed to remove Dockerfile from %s (%s)", tmpDir, err)
		}
	}()

	bType = getBuildTypeForDir(tmpDir)
	if bType != buildTypeDockerfile {
		t.Fatalf("expected dockerfile build, got %s", bType)
	}
}
