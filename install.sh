#!/bin/sh
#
# sprintctl installer script
#
# This script is designed to be run via curl:
#   curl -fsSL https://raw.githubusercontent.com/morethancertified/sprintctl/main/install.sh | sh
#
# It automatically detects the OS and architecture, then downloads the
# appropriate binary from the latest GitHub release.

set -e

# --- Configuration ---
REPO="morethancertified/sprintctl"
CLI_NAME="sprintctl"
INSTALL_DIR="/usr/local/bin"

# --- Helper Functions ---
info() {
    echo "INFO: $1"
}

error() {
    echo "ERROR: $1" >&2
    exit 1
}

# --- Main Logic ---

# 1. Detect OS and Architecture
os_type=$(uname -s | tr '[:upper:]' '[:lower:]')
arch_type=$(uname -m)

case "$os_type" in
    darwin)
        os="darwin"
        ;;
    linux)
        os="linux"
        ;;
    *)
        error "Unsupported OS: $os_type. Only macOS and Linux are supported."
        ;;
esac

case "$arch_type" in
    x86_64 | amd64)
        arch="amd64"
        ;;
    arm64 | aarch64)
        arch="arm64"
        ;;
    *)
        error "Unsupported architecture: $arch_type."
        ;;
esac

info "Detected OS: $os, Architecture: $arch"

# 2. Get the latest release version
info "Fetching the latest version of $CLI_NAME..."
latest_release_url="https://api.github.com/repos/$REPO/releases/latest"
# Use sed for portable JSON parsing, avoiding non-standard grep flags.
latest_version=$(curl -s "$latest_release_url" | sed -n 's/.*\"tag_name\": *\"\([^\"]*\)\".*/\1/p')

if [ -z "$latest_version" ]; then
    error "Could not fetch the latest release version. Check the repository and your connection."
fi

info "Latest version is $latest_version"

# 3. Construct the download URL
# The asset name format is based on your .goreleaser.yaml config (format: binary).
# We strip the 'v' from the version tag for the filename.
version_without_v=${latest_version#v}
asset_name="${CLI_NAME}_${version_without_v}_${os}_${arch}"
download_url="https://github.com/$REPO/releases/download/$latest_version/$asset_name"

info "Downloading from: $download_url"

# 4. Download to a temporary file
temp_file=$(mktemp)
trap 'rm -f "$temp_file"' EXIT # Ensure temp file is cleaned up on exit.

curl -fsSL "$download_url" -o "$temp_file"

info "Download complete."

# 5. Install the binary
install_path="$INSTALL_DIR/$CLI_NAME"
info "Installing $CLI_NAME to $install_path..."

# Use sudo if the install directory is not writable by the current user.
if [ -w "$INSTALL_DIR" ]; then
    mv "$temp_file" "$install_path"
    chmod +x "$install_path"
else
    echo "This script needs to install the '$CLI_NAME' binary to '$INSTALL_DIR'."
    echo "This may require administrative privileges."
    sudo mv "$temp_file" "$install_path"
    sudo chmod +x "$install_path"
fi

# 6. Verify installation
if [ -x "$install_path" ]; then
    info "$CLI_NAME installed successfully!"
    echo "You can now run '$CLI_NAME --version' to get started."
else
    error "Installation failed. The file was not found or is not executable at $install_path."
fi
