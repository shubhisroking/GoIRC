#!/bin/bash

# 🚀 GoIRC Enhanced Installation Script
# A modern IRC client built with Go and Bubble Tea

set -e

# Colors for enhanced output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m' # No Color

# Enhanced Unicode symbols
CHECKMARK="✅"
CROSS="❌"
INFO="ℹ️"
ROCKET="🚀"
GEAR="⚙️"
PACKAGE="📦"
SPARKLES="✨"
WARNING="⚠️"
ARROW="➜"
BULLET="•"

# Installation configuration
REPO_URL="https://github.com/yourusername/goirc"
INSTALL_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.config/goirc"
DATA_DIR="$HOME/.local/share/goirc"
LOG_DIR="$HOME/.local/share/goirc/logs"
VERSION="latest"

# Feature flags
ENABLE_AUTO_START=false
ENABLE_DESKTOP_ENTRY=false
SKIP_DEPS_CHECK=false

print_banner() {
    echo -e "${PURPLE}${BOLD}"
    echo "╔══════════════════════════════════════════════════════════════════════╗"
    echo "║                                                                      ║"
    echo "║                    ${ROCKET} GoIRC Enhanced Installer                       ║"
    echo "║                                                                      ║"
    echo "║           ${SPARKLES} A modern IRC client built with Go & Bubble Tea          ║"
    echo "║                                                                      ║"
    echo "║                     ${CYAN}https://github.com/goirc/goirc${PURPLE}                    ║"
    echo "║                                                                      ║"
    echo "╚══════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo
}

# Enhanced logging functions
log() {
    echo -e "${BLUE}${INFO} ${BOLD}$1${NC}"
}

success() {
    echo -e "${GREEN}${CHECKMARK} ${BOLD}$1${NC}"
}

error() {
    echo -e "${RED}${CROSS} ${BOLD}$1${NC}"
}

warn() {
    echo -e "${YELLOW}${WARNING} ${BOLD}$1${NC}"
}

highlight() {
    echo -e "${CYAN}${ARROW} ${BOLD}$1${NC}"
}

step() {
    echo -e "${PURPLE}${GEAR} ${BOLD}Step $1: $2${NC}"
    echo
}

check_requirements() {
    step "1" "Checking System Requirements"
    
    # Check operating system
    OS="$(uname -s)"
    case "${OS}" in
        Linux*)     MACHINE=Linux;;
        Darwin*)    MACHINE=Mac;;
        CYGWIN*)    MACHINE=Cygwin;;
        MINGW*)     MACHINE=MinGw;;
        *)          MACHINE="UNKNOWN:${OS}"
    esac
    
    highlight "Detected OS: $MACHINE"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        error "Go is not installed!"
        echo
        echo -e "${CYAN}${BULLET} Please install Go 1.21 or higher from: ${BOLD}https://golang.org/doc/install${NC}"
        echo -e "${CYAN}${BULLET} Or use your package manager:${NC}"
        echo -e "  ${DIM}Ubuntu/Debian: ${WHITE}sudo apt install golang-go${NC}"
        echo -e "  ${DIM}CentOS/RHEL:   ${WHITE}sudo yum install golang${NC}"
        echo -e "  ${DIM}macOS:         ${WHITE}brew install go${NC}"
        echo -e "  ${DIM}Arch Linux:    ${WHITE}sudo pacman -S go${NC}"
        echo
        exit 1
    fi
    
    # Check Go version with enhanced parsing
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.21"
    
    log "Checking Go version: $GO_VERSION"
    
    if ! version_compare "$GO_VERSION" "$REQUIRED_VERSION"; then
        error "Go version $GO_VERSION is too old!"
        echo -e "${CYAN}${BULLET} Required: Go $REQUIRED_VERSION or higher${NC}"
        echo -e "${CYAN}${BULLET} Please upgrade Go from: ${BOLD}https://golang.org/doc/install${NC}"
        exit 1
    fi
    
    success "Go $GO_VERSION is installed and compatible"
    
    # Check if Git is installed
    if ! command -v git &> /dev/null; then
        error "Git is not installed!"
        echo -e "${CYAN}${BULLET} Please install Git from your package manager${NC}"
        exit 1
    fi
    
    success "Git is installed"
    
    # Check available tools
    local tools_found=()
    local tools_missing=()
    
    if command -v make &> /dev/null; then
        tools_found+=("make")
    else
        tools_missing+=("make")
    fi
    
    if command -v curl &> /dev/null; then
        tools_found+=("curl")
    else
        tools_missing+=("curl")
    fi
    
    if command -v wget &> /dev/null; then
        tools_found+=("wget")
    else
        tools_missing+=("wget")
    fi
    
    if [ ${#tools_found[@]} -gt 0 ]; then
        highlight "Available tools: ${tools_found[*]}"
    fi
    
    if [ ${#tools_missing[@]} -gt 0 ]; then
        warn "Optional tools not found: ${tools_missing[*]}"
        echo -e "${DIM}  These tools can enhance the installation experience${NC}"
    fi
    
    # Check directory permissions
    log "Checking directory permissions..."
    
    if [ ! -w "$(dirname "$INSTALL_DIR")" ]; then
        warn "Cannot write to $(dirname "$INSTALL_DIR")"
        echo -e "${CYAN}${BULLET} Will attempt to create directory with mkdir -p${NC}"
    fi
    
    success "System requirements check completed"
    echo
}

enhanced_version_compare() {
    local version1=$1
    local version2=$2
    
    # Convert versions to comparable format
    version1_numeric=$(echo "$version1" | sed 's/[^0-9.]//g')
    version2_numeric=$(echo "$version2" | sed 's/[^0-9.]//g')
    
    # Use sort -V if available, fallback to basic comparison
    if command -v sort &> /dev/null && sort --version-sort /dev/null &> /dev/null 2>&1; then
        if [ "$(printf '%s\n' "$version2_numeric" "$version1_numeric" | sort -V | head -n1)" = "$version2_numeric" ]; then
            return 0
        else
            return 1
        fi
    else
        # Fallback comparison
        if [ "$(printf '%s\n' "$version2_numeric" "$version1_numeric" | sort -n | head -n1)" = "$version2_numeric" ]; then
            return 0
        else
            return 1
        fi
    fi
}

version_compare() {
    local version1=$1
    local version2=$2
    
    # Simple version comparison (works for Go versions like 1.21.0)
    if [ "$(printf '%s\n' "$version2" "$version1" | sort -V | head -n1)" = "$version2" ]; then
        return 0
    else
        return 1
    fi
}

create_directories() {
    log "Creating directories..."
    
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$DATA_DIR"
    
    success "Directories created"
}

build_goirc() {
    log "Building GoIRC..."
    
    # Build the application
    if command -v make &> /dev/null; then
        make build
    else
        go build -ldflags="-s -w" -o goirc .
    fi
    
    success "GoIRC built successfully"
}

install_binary() {
    log "Installing GoIRC binary..."
    
    # Copy binary to install directory
    cp goirc "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/goirc"
    
    success "GoIRC installed to $INSTALL_DIR/goirc"
}

create_config() {
    log "Creating default configuration..."
    
    CONFIG_FILE="$CONFIG_DIR/config.json"
    
    if [ ! -f "$CONFIG_FILE" ]; then
        cat > "$CONFIG_FILE" << 'EOF'
{
  "irc": {
    "server": "irc.libera.chat:6697",
    "nick": "goirc_user",
    "username": "goirc_user",
    "realname": "GoIRC User",
    "channels": ["#test"],
    "use_ssl": true,
    "quit_message": "Goodbye from GoIRC!"
  },
  "ui": {
    "show_sidebar": true,
    "sidebar_width": 20,
    "theme": {
      "primary": "#7C3AED",
      "secondary": "#06B6D4",
      "accent": "#F59E0B"
    }
  },
  "logging": {
    "enabled": true,
    "max_size_kb": 1024,
    "log_path": "~/.local/share/goirc/",
    "debug_mode": false
  }
}
EOF
        success "Default configuration created at $CONFIG_FILE"
    else
        warn "Configuration file already exists at $CONFIG_FILE"
    fi
}

create_desktop_entry() {
    if [ -n "$XDG_DATA_HOME" ]; then
        DESKTOP_DIR="$XDG_DATA_HOME/applications"
    else
        DESKTOP_DIR="$HOME/.local/share/applications"
    fi
    
    if command -v desktop-file-validate &> /dev/null; then
        log "Creating desktop entry..."
        
        mkdir -p "$DESKTOP_DIR"
        
        cat > "$DESKTOP_DIR/goirc.desktop" << EOF
[Desktop Entry]
Version=1.0
Type=Application
Name=GoIRC
Comment=Modern IRC Client
Exec=$INSTALL_DIR/goirc
Icon=utilities-terminal
Terminal=true
Categories=Network;Chat;
Keywords=irc;chat;network;terminal;
EOF
        
        if desktop-file-validate "$DESKTOP_DIR/goirc.desktop"; then
            success "Desktop entry created"
        else
            warn "Desktop entry validation failed"
        fi
    fi
}

setup_path() {
    log "Checking PATH configuration..."
    
    # Check if install directory is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        warn "$INSTALL_DIR is not in your PATH"
        echo
        echo -e "${YELLOW}To use GoIRC from anywhere, add this to your shell profile:${NC}"
        echo -e "${CYAN}export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
        echo
        echo -e "${YELLOW}Shell profile locations:${NC}"
        echo -e "${CYAN}  Bash: ~/.bashrc or ~/.bash_profile${NC}"
        echo -e "${CYAN}  Zsh:  ~/.zshrc${NC}"
        echo -e "${CYAN}  Fish: ~/.config/fish/config.fish${NC}"
        echo
    else
        success "$INSTALL_DIR is already in your PATH"
    fi
}

print_completion_message() {
    echo
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                               ║${NC}"
    echo -e "${GREEN}║                 ${CHECKMARK} Installation Complete! ${CHECKMARK}                 ║${NC}"
    echo -e "${GREEN}║                                                               ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo
    echo -e "${WHITE}${ROCKET} GoIRC has been successfully installed!${NC}"
    echo
    echo -e "${CYAN}Installation details:${NC}"
    echo -e "  Binary:        $INSTALL_DIR/goirc"
    echo -e "  Configuration: $CONFIG_DIR/"
    echo -e "  Data:          $DATA_DIR/"
    echo
    echo -e "${CYAN}Getting started:${NC}"
    echo -e "  ${WHITE}goirc${NC}                    # Start with interactive setup"
    echo -e "  ${WHITE}goirc --help${NC}             # Show help message"
    echo
    echo -e "${CYAN}Configuration:${NC}"
    echo -e "  Edit: ${WHITE}$CONFIG_FILE${NC}"
    echo
    echo -e "${CYAN}Useful commands:${NC}"
    echo -e "  ${WHITE}make help${NC}                # Show all available commands"
    echo -e "  ${WHITE}make dev${NC}                 # Run development workflow"
    echo -e "  ${WHITE}make build-all${NC}           # Cross-compile for all platforms"
    echo
    echo -e "${PURPLE}Enjoy using GoIRC! ${ROCKET}${NC}"
    echo
}

main() {
    print_banner
    
    # Check if we're in the goirc directory
    if [ ! -f "go.mod" ] || ! grep -q "goirc" go.mod; then
        error "This script must be run from the GoIRC source directory"
        exit 1
    fi
    
    check_requirements
    create_directories
    build_goirc
    install_binary
    create_config
    create_desktop_entry
    setup_path
    
    print_completion_message
}

# Handle Ctrl+C
trap 'echo -e "\n${RED}${CROSS} Installation cancelled by user${NC}"; exit 1' INT

# Run main function
main "$@"
