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

// generatePackageLock generates a package-lock.json by executing npm
// against the package.json file in the dir directory.
//
// It returns the package-lock.json file as a byte array.
//
// It assumes that the input directory exists and it already contains
// a package.json file.
func generatePackageLock(dir string) ([]byte, error) {
	// Get the npm command
	npmPackageLockOnly, err := getNPMPackageLockOnly()
	if err != nil {
		// Fallback to npm via nvm
		npmPackageLockOnlyFromNVM, nvmErr := getNPMPackageLockOnlyFromNVM()
		if nvmErr != nil {
			// FIXME > return more errors or a generic one
			return []byte{}, err
		}
		npmPackageLockOnly = npmPackageLockOnlyFromNVM
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
	npmPackageLockOnly.Dir = tmp
	if err := npmPackageLockOnly.Run(); err != nil {
		return []byte{}, fmt.Errorf("couldn't generate the package-lock.json file")
	}
	packageLockJSON, _ := os.ReadFile(filepath.Join(tmp, "package-lock.json"))

	return packageLockJSON, nil
}

func checkNPMVersion(c *exec.Cmd, constraint string) error {
	// Obtain the npm version
	npmVersionCmd := c
	npmVersionOut, err := npmVersionCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("couldn't get the npm version")
	}
	npmVersionString := string(bytes.Trim(npmVersionOut, "\n"))

	// Check the npm version is valid
	npmVersionErrors := validate.Singleton.Var(npmVersionString, "semver")
	if npmVersionErrors != nil {
		return fmt.Errorf("couldn't validate the npm version string")
	}
	npmVersion, err := semver.NewVersion(npmVersionString)
	if err != nil {
		return fmt.Errorf("couldn't get a valid npm version")
	}

	// Check the npm is at least version 6.x
	npmVersionConstraint, err := semver.NewConstraint(constraint)
	if err != nil {
		return fmt.Errorf("couldn't compare the npm version")
	}
	npmVersionValid, _ := npmVersionConstraint.Validate(npmVersion)
	if !npmVersionValid {
		return fmt.Errorf("the npm version is not %s", constraint)
	}
	return nil
}

// getNPMPackageLockOnly returns the command to generate the package-lock.json file.
//
// It also checks that:
// - the npm executable is available in the PATH
// - its version is greater or equal than version "6.x".
func getNPMPackageLockOnly() (*exec.Cmd, error) {
	// Check the system has the npm executable
	exe, err := exec.LookPath("npm")
	if err != nil {
		return nil, fmt.Errorf("couldn't find the npm executable in the PATH")
	}

	npmVersionCmd := exec.Command(exe, "--version")
	if err := checkNPMVersion(npmVersionCmd, ">= 6.x"); err != nil {
		return nil, err
	}

	return exec.Command(exe, "install", "--package-lock-only", "--no-audit"), nil
}

// getNPMPackageLockOnlyFromNVM return the command to generate the package-lock.json file
// when the npm executable is behind nvm.
//
// In fact, it is likely that npm is not in the PATH because nvm is lazy-loading it.
//
// It also checks that:
// - the npm version is greater or equal than version "6.x".
func getNPMPackageLockOnlyFromNVM() (*exec.Cmd, error) {
	nvmDir := os.Getenv("NVM_DIR")
	if nvmDir == "" {
		return nil, fmt.Errorf("couldn't detect the nvm directory")
	}

	cmdline := fmt.Sprintf("source %s/nvm.sh", nvmDir)

	nvmNoUse := os.Getenv("NVM_NO_USE")
	if nvmNoUse == "true" {
		cmdline += " --no-use"
	}

	// Obtain the npm version
	npmVersionCmd := exec.Command("bash", "-c", fmt.Sprintf("%s && npm --version", cmdline))
	if err := checkNPMVersion(npmVersionCmd, ">= 6.x"); err != nil {
		return nil, err
	}

	return exec.Command("bash", "-c", fmt.Sprintf("%s && npm install --package-lock-only --no-audit", cmdline)), nil
}
