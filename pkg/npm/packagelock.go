package npm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/listendev/lstn/pkg/validate"
)

// GeneratePackageLock generates a package-lock.json by executing npm
// against the package.json file in the dir directory.
//
// It returns the package-lock.json file as a byte array.
//
// It assumes that the input directory exists and it already contains
// a package.json file.
func GeneratePackageLock(dir string) ([]byte, error) {
	// Get the npm executable
	exe, err := getNPM()
	if err != nil {
		return []byte{}, err
	}

	// Create temporary directory
	tmp, err := os.MkdirTemp("", "lstn-*")
	if err != nil {
		return []byte{}, fmt.Errorf("couldn't create a temporary directory where to do the dirty work")
	}
	defer os.RemoveAll(tmp)

	// Copy the package.json in the temporary directory
	packageJSONPath := filepath.Join(dir, "package.json")
	packageJSON, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return []byte{}, fmt.Errorf("couldn't read the package.json file")
	}
	if err := os.WriteFile(filepath.Join(tmp, "package.json"), packageJSON, 0644); err != nil {
		return []byte{}, fmt.Errorf("couldn't copy the package.json file")
	}

	// Generate the package-lock.json file
	// TODO(leodido) > Show progress?
	npmPackageLockOnly := exec.Command(exe, "install", "--package-lock-only", "--no-audit")
	npmPackageLockOnly.Dir = tmp
	if err := npmPackageLockOnly.Run(); err != nil {
		return []byte{}, fmt.Errorf("couldn't generate the package-lock.json file")
	}
	packageLockJSON, _ := os.ReadFile(filepath.Join(tmp, "package-lock.json"))

	return packageLockJSON, nil
}

// getNPM returns the absolute path of the npm executable.
//
// It checks that the npm version is >= 6.x too.
func getNPM() (string, error) {
	// Check the system has the npm executable
	exe, err := exec.LookPath("npm")
	if err != nil {
		return "", fmt.Errorf("couldn't find the npm executable")
	}

	// Obtain the npm version
	npmVersionCmd := exec.Command(exe, "--version")
	npmVersionOut, err := npmVersionCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("couldn't get the npm version")
	}
	npmVersionString := string(bytes.Trim(npmVersionOut, "\n"))

	// Check the npm version is valid
	npmVersionErrors := validate.Singleton.Var(npmVersionString, "semver")
	if npmVersionErrors != nil {
		return "", fmt.Errorf("couldn't validate the npm version string")
	}
	npmVersion, err := semver.NewVersion(npmVersionString)
	if err != nil {
		return "", fmt.Errorf("couldn't get a valid npm version")
	}

	// Check the npm is at least version 6.x
	npmVersionConstraint, err := semver.NewConstraint(">= 6.x")
	if err != nil {
		return "", fmt.Errorf("couldn't compare the npm version")
	}
	npmVersionValid, _ := npmVersionConstraint.Validate(npmVersion)
	if !npmVersionValid {
		return "", fmt.Errorf("we do not support npm version < 6.x")
	}

	return exe, nil
}
