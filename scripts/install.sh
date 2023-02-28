#!/usr/bin/env sh

set -e

# Error codes:
# 1 - general
# 2 - curl or wget not found
# 3 - signature verification issue

CLEAN_EXIT=0

log() {
  # print to stderr
  >&2 echo "$1"
}

# Clean up temporary files
cleanup() {
  exit_code=$?
  if [ "$exit_code" -ne 0 ] && [ "$CLEAN_EXIT" -ne 1 ]; then
    log "ERROR: script failed during execution process"

    if [ "$DEBUG" -eq 0 ]; then
      log "For more verbose output, execute this script using the debug flag (./install.sh --debug)"
    fi
  fi

  if [ -f "lstn_${RELEASE_VERSION}_checksums.txt" ]; then
    rm -rf lstn_${RELEASE_VERSION}_checksums.txt
  fi

  if [ -f "lstn_${RELEASE_VERSION}_${HOST_OS}_${HOST_ARCH}.tar.gz" ]; then
    rm -rf lstn_${RELEASE_VERSION}_${HOST_OS}_${HOST_ARCH}.tar.gz
  fi

  if [ -f "${RELEASE_VERSION}" ]; then
    rm -rf "${RELEASE_VERSION}"
  fi

  clean_exit "$exit_code"
}

trap cleanup EXIT

clean_exit() {
  CLEAN_EXIT=1
  exit "$1"
}

# Check latest release from GitHub, if RELEASE_VERSION is not set explicitly
check_release_version() {
  if [ -z "${RELEASE_VERSION}" ]; then
    if command -v curl > /dev/null; then
      RELEASE_VERSION=$(curl --silent "https://api.github.com/repos/listendev/lstn/releases/latest" | sed -n 's/.*"tag_name": "v\([^"]*\)".*/\1/p')
    elif command -v wget > /dev/null; then
      RELEASE_VERSION=$(wget -qO- "https://api.github.com/repos/listendev/lstn/releases/latest" | sed -n 's/.*"tag_name": "v\([^"]*\)".*/\1/p')
    else
      echo "Neither curl nor wget is available. Cannot download release version."
      exit 1
    fi
  fi
}

# Identify OS
HOST_OS="unknown"
uname_os=$(uname -s)

detect_host_os() {
  case "$uname_os" in
    Darwin)    HOST_OS="macos"   ;;
    Linux)     HOST_OS="linux"   ;;
    *MINGW64*) HOST_OS="windows" ;;
    *)
      log "ERROR: Unsupported OS '$uname_os'"
      log "Please report this issue:"
      log "https://github.com/listendev/lstn/issues/new?template=bug_report.md&title=[BUG]%20Unsupported%20OS"
      clean_exit 1
      ;;
  esac
  echo "Detected OS '$HOST_OS'"
}

# Identify arch
HOST_ARCH="unknown"
uname_machine=$(uname -m)

detect_host_arch() {
  if [ "$uname_machine" = "i386" ] || [ "$uname_machine" = "i686" ]; then
    HOST_ARCH="386"
  elif [ "$uname_machine" = "amd64" ] || [ "$uname_machine" = "x86_64" ]; then
    HOST_ARCH="amd64"
  elif [ "$uname_machine" = "armv6" ] || [ "$uname_machine" = "armv6l" ]; then
    HOST_ARCH="armv6"
  elif [ "$uname_machine" = "arm64" ] || [ "$uname_machine" = "aarch64" ]; then
    HOST_ARCH="arm64"
  else
    log "ERROR: Unsupported architecture '$uname_machine'"
    log "Please report this issue:"
    log "https://github.com/listendev/lstn/issues/new?template=bug_report.md&title=[BUG]%20Unsupported%20arch"
    clean_exit 1
  fi
  echo "Detected arch '$HOST_ARCH'"
}

# Check if curl is available, otherwise try to use wget
DOWNLOAD_CMD=""
check_download_command() {
  if command -v curl > /dev/null; then
    DOWNLOAD_CMD="curl -L -O"
  elif command -v wget > /dev/null; then
    DOWNLOAD_CMD="wget"
  else
    log "ERROR: Neither curl nor wget are installed"
    clean_exit 2
  fi
}



# Construct the download URL based on the host OS and arch
construct_download_url() {
  check_release_version
  detect_host_os
  detect_host_arch
  case "${HOST_OS}_${HOST_ARCH}" in
    "linux_armv6")
      DOWNLOAD_URL="https://github.com/listendev/lstn/releases/download/v${RELEASE_VERSION}/lstn_${RELEASE_VERSION}_linux_armv6.tar.gz"
      ;;
    "linux_amd64")
      DOWNLOAD_URL="https://github.com/listendev/lstn/releases/download/v${RELEASE_VERSION}/lstn_${RELEASE_VERSION}_linux_amd64.tar.gz"
      ;;
    "linux_386")
      DOWNLOAD_URL="https://github.com/listendev/lstn/releases/download/v${RELEASE_VERSION}/lstn_${RELEASE_VERSION}_linux_386.tar.gz"
      ;;
    "linux_arm64")
      DOWNLOAD_URL="https://github.com/listendev/lstn/releases/download/v${RELEASE_VERSION}/lstn_${RELEASE_VERSION}_linux_arm64.tar.gz"
      ;;
    "macos_amd64")
      DOWNLOAD_URL="https://github.com/listendev/lstn/releases/download/v${RELEASE_VERSION}/lstn_${RELEASE_VERSION}_macos_amd64.tar.gz"
      ;;
    "macos_arm64")
      DOWNLOAD_URL="https://github.com/listendev/lstn/releases/download/v${RELEASE_VERSION}/lstn_${RELEASE_VERSION}_macos_arm64.tar.gz"
      ;;
    *)
      echo "Unsupported host OS/Arch: ${HOST_OS}/${HOST_ARCH}"
      exit 1
      ;;
  esac
}



# Download the binary and its related files
download_binary() {
  check_download_command
  construct_download_url

  echo "Downloading ${DOWNLOAD_URL} ..."
  $DOWNLOAD_CMD "${DOWNLOAD_URL}"

  # Download the checksum file
  CHECKSUM_URL="https://github.com/listendev/lstn/releases/download/v${RELEASE_VERSION}/lstn_${RELEASE_VERSION}_checksums.txt"
  echo "Downloading ${CHECKSUM_URL} ..."
  $DOWNLOAD_CMD "${CHECKSUM_URL}"
}

# Verify the integrity of the downloaded binary
verify_binary_integrity() {
  echo "Verifying the SHA256 checksum of the downloaded binary..."
  if command -v shasum > /dev/null; then
    if ! grep "$(basename "${DOWNLOAD_URL}")" "lstn_${RELEASE_VERSION}_checksums.txt" | shasum -a 256 -c --ignore-missing; then
      echo "Failed to verify the integrity of the downloaded binary"
      clean_exit 3
    fi
  else
    if ! grep "$(basename "${DOWNLOAD_URL}")" "lstn_${RELEASE_VERSION}_checksums.txt" | sha256sum -c; then
      echo "Failed to verify the integrity of the downloaded binary"
      clean_exit 3
    fi
  fi
}

# Install the binary to the system
install_lstn() {

  verify_binary_integrity
  EXTRACT_DIR="lstn_${RELEASE_VERSION}_${HOST_OS}_${HOST_ARCH}"
  case "${HOST_OS}_${HOST_ARCH}" in
    "linux_armv6" | "linux_386" | "linux_amd64" | "linux_arm64")
      INSTALL_DIR="/usr/bin"
      tar -xzf "lstn_${RELEASE_VERSION}_${HOST_OS}_${HOST_ARCH}.tar.gz" -C "/tmp/"
      chown "$(id -u):$(id -g)" "/tmp/$EXTRACT_DIR/lstn"
      chmod 755 "/tmp/$EXTRACT_DIR/lstn"
      mv -f "/tmp/$EXTRACT_DIR/lstn" "$INSTALL_DIR"
      rm -rf "/tmp/$EXTRACT_DIR"
      ;;
    "macos_amd64" | "macos_arm64")
      INSTALL_DIR="/usr/local/bin"
      tar -xzf "lstn_${RELEASE_VERSION}_${HOST_OS}_${HOST_ARCH}.tar.gz" -C "/tmp/"
      chown "$(id -u):$(id -g)" "/tmp/$EXTRACT_DIR/lstn"
      chmod 755 "/tmp/$EXTRACT_DIR/lstn"
      mv -f "/tmp/$EXTRACT_DIR/lstn" "$INSTALL_DIR"
      rm -rf "/tmp/$EXTRACT_DIR"
      ;;
    *)
      echo "Unsupported host OS/Arch: ${HOST_OS}/${HOST_ARCH}"
      exit 1
      ;;
  esac
  echo "lstn version $(lstn version) has been installed successfully in $INSTALL_DIR"
}

# Check for existing versions
check_existing_versions() {
  if command -v lstn > /dev/null; then
    INSTALLED_VERSION=$(lstn version | grep -o -E 'lstn v[0-9]+\.[0-9]+\.[0-9]+' | head -n1 | sed 's/^lstn v//' | sed 's/^v//')
    if command -v semver > /dev/null; then
      # Check if installed version is greater than or equal to release version
      if semver -r "${RELEASE_VERSION#v}" -r ">=${INSTALLED_VERSION}"; then
        echo "The latest version of lstn ${INSTALLED_VERSION} is already installed."
      else
        echo "Updating lstn ${INSTALLED_VERSION} to the latest version: ${RELEASE_VERSION}"
        install_lstn
      fi
    else
      # semver command not found, do string comparison instead
      if [[ "${INSTALLED_VERSION}" == "${RELEASE_VERSION}" ]]; then
        echo "lstn is installed and up-to-date."
      elif [[ "${INSTALLED_VERSION}" > "${RELEASE_VERSION}" ]]; then
        echo "The latest version of lstn ${INSTALLED_VERSION} is already installed."
      else
        echo "Updating lstn ${INSTALLED_VERSION} to the latest version: ${RELEASE_VERSION}"
        install_lstn
      fi
    fi
  else
    echo "No existing installation found for lstn. Proceeding to install the latest version..."
    install_lstn
  fi
}

download_binary
check_existing_versions
cleanup