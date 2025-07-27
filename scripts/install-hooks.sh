#!/bin/bash

# Pre-commit Hook Installation Script for E-Paper Dashboard
# Installs and configures comprehensive pre-commit hooks with no exceptions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${GREEN}🔧 Installing Pre-commit Hook System${NC}"
echo "Project root: $PROJECT_ROOT"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install pre-commit if not available
install_precommit() {
    echo -e "${YELLOW}📦 Installing pre-commit...${NC}"
    
    if command_exists pip3; then
        pip3 install pre-commit
    elif command_exists pip; then
        pip install pre-commit
    elif command_exists brew; then
        brew install pre-commit
    else
        echo -e "${RED}❌ Error: Could not install pre-commit. Please install it manually:${NC}"
        echo "  pip install pre-commit"
        echo "  or"
        echo "  brew install pre-commit"
        exit 1
    fi
}

# Function to install Go tools
install_go_tools() {
    echo -e "${YELLOW}🔧 Installing Go tools...${NC}"
    
    # Install gocyclo for complexity checking
    if ! command_exists gocyclo; then
        go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
    fi
    
    # Install gosec for security scanning
    if ! command_exists gosec; then
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    fi
    
    # Install golangci-lint for comprehensive linting
    if ! command_exists golangci-lint; then
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
    fi
}

# Function to install Python tools
install_python_tools() {
    echo -e "${YELLOW}🐍 Installing Python tools...${NC}"
    
    # Install Python linting and formatting tools
    pip3 install black flake8 pylint mypy
}

# Function to create secrets baseline
create_secrets_baseline() {
    echo -e "${YELLOW}🔒 Creating secrets baseline...${NC}"
    
    cd "$PROJECT_ROOT"
    
    # Install detect-secrets if not available
    if ! command_exists detect-secrets; then
        pip3 install detect-secrets
    fi
    
    # Create baseline file
    detect-secrets scan --baseline .secrets.baseline --exclude-files 'go\.sum$|go\.mod$|coverage\.out$|\.git/.*$'
    
    echo -e "${GREEN}✅ Created .secrets.baseline${NC}"
}

# Function to verify pre-commit configuration
verify_config() {
    echo -e "${YELLOW}📋 Verifying pre-commit configuration...${NC}"
    
    cd "$PROJECT_ROOT"
    
    if [ ! -f ".pre-commit-config.yaml" ]; then
        echo -e "${RED}❌ Error: .pre-commit-config.yaml not found${NC}"
        exit 1
    fi
    
    # Validate the configuration
    pre-commit validate-config
    echo -e "${GREEN}✅ Configuration is valid${NC}"
}

# Function to install the hooks
install_hooks() {
    echo -e "${YELLOW}🪝 Installing pre-commit hooks...${NC}"
    
    cd "$PROJECT_ROOT"
    
    # Install the pre-commit hooks
    pre-commit install --install-hooks
    
    # Install pre-push hooks as well
    pre-commit install --hook-type pre-push
    
    echo -e "${GREEN}✅ Pre-commit hooks installed${NC}"
}

# Function to create additional Git hooks
create_git_hooks() {
    echo -e "${YELLOW}⚙️ Creating additional Git hooks...${NC}"
    
    # Create commit-msg hook for conventional commits
    cat > "$PROJECT_ROOT/.git/hooks/commit-msg" << 'EOF'
#!/bin/bash
# Commit message validation hook
# Enforces conventional commit format

commit_regex='^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\(.+\))?: .{1,50}'

error_msg="Invalid commit message format!
Expected format: type(scope): description
Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert
Example: feat(admin): add widget validation system"

if ! grep -qE "$commit_regex" "$1"; then
    echo "$error_msg" >&2
    exit 1
fi
EOF

    # Make commit-msg hook executable
    chmod +x "$PROJECT_ROOT/.git/hooks/commit-msg"
    
    echo -e "${GREEN}✅ Additional Git hooks created${NC}"
}

# Function to test the installation
test_installation() {
    echo -e "${YELLOW}🧪 Testing hook installation...${NC}"
    
    cd "$PROJECT_ROOT"
    
    # Run pre-commit on all files to test
    echo "Running pre-commit on all files (this may take a while)..."
    
    if pre-commit run --all-files; then
        echo -e "${GREEN}✅ All hooks passed successfully${NC}"
    else
        echo -e "${YELLOW}⚠️  Some hooks failed or made changes${NC}"
        echo "This is normal for the first run - hooks may auto-fix issues"
        echo "Run 'git add .' and commit to apply the fixes"
    fi
}

# Function to display usage information
show_usage() {
    cat << EOF

${GREEN}🎉 Pre-commit Hook System Installation Complete!${NC}

${YELLOW}What was installed:${NC}
✅ Pre-commit framework
✅ Go linting and security tools
✅ Python formatting and linting tools
✅ Comprehensive hook configuration
✅ Conventional commit validation
✅ Secrets detection baseline

${YELLOW}How it works:${NC}
• Every commit automatically runs:
  - Go formatting, imports, vet, complexity checks
  - Go unit tests with race detection
  - Python syntax, formatting, linting (for widgets)
  - Security scanning and secret detection
  - File validation and documentation checks

${YELLOW}No exceptions policy:${NC}
🚨 ALL checks must pass before commits are accepted
🚨 Tests must pass with 100% success rate
🚨 No --no-verify bypassing allowed in CI/CD

${YELLOW}Common commands:${NC}
• pre-commit run --all-files    - Run all hooks on all files
• pre-commit run <hook-name>    - Run specific hook
• pre-commit autoupdate         - Update hook versions
• make test                     - Run project test suite

${YELLOW}If hooks fail:${NC}
1. Fix the reported issues
2. Add the fixed files: git add .
3. Commit again

${GREEN}Happy coding with enforced quality! 🚀${NC}

EOF
}

# Main installation process
main() {
    echo
    echo "Starting pre-commit hook installation..."
    echo "This will set up comprehensive code quality enforcement."
    echo
    
    # Check if pre-commit is installed
    if ! command_exists pre-commit; then
        install_precommit
    else
        echo -e "${GREEN}✅ pre-commit already installed${NC}"
    fi
    
    # Install Go tools
    if command_exists go; then
        install_go_tools
    else
        echo -e "${YELLOW}⚠️  Go not found, skipping Go tools installation${NC}"
    fi
    
    # Install Python tools
    if command_exists python3 || command_exists python; then
        install_python_tools
    else
        echo -e "${YELLOW}⚠️  Python not found, skipping Python tools installation${NC}"
    fi
    
    # Create secrets baseline
    create_secrets_baseline
    
    # Verify configuration
    verify_config
    
    # Install hooks
    install_hooks
    
    # Create additional Git hooks
    create_git_hooks
    
    # Test installation
    test_installation
    
    # Show usage information
    show_usage
}

# Run main function
main "$@"