#!/bin/bash

# Commit Validation Script
# Provides manual validation and quality checking before commits
# Can be used independently of pre-commit hooks for testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Function to print section headers
print_section() {
    echo
    echo -e "${BLUE}======================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}======================================${NC}"
}

# Function to run Go formatting check
check_go_format() {
    print_section "Go Format Check"
    
    cd "$PROJECT_ROOT"
    
    unformatted=$(gofmt -l .)
    if [ -n "$unformatted" ]; then
        echo -e "${RED}❌ The following files are not properly formatted:${NC}"
        echo "$unformatted"
        echo -e "${YELLOW}Run 'gofmt -w .' to fix${NC}"
        return 1
    else
        echo -e "${GREEN}✅ All Go files are properly formatted${NC}"
        return 0
    fi
}

# Function to run Go imports check
check_go_imports() {
    print_section "Go Imports Check"
    
    cd "$PROJECT_ROOT"
    
    if command -v goimports >/dev/null 2>&1; then
        unorganized=$(goimports -l .)
        if [ -n "$unorganized" ]; then
            echo -e "${RED}❌ The following files have unorganized imports:${NC}"
            echo "$unorganized"
            echo -e "${YELLOW}Run 'goimports -w .' to fix${NC}"
            return 1
        else
            echo -e "${GREEN}✅ All Go imports are properly organized${NC}"
            return 0
        fi
    else
        echo -e "${YELLOW}⚠️  goimports not found, skipping import check${NC}"
        return 0
    fi
}

# Function to run Go vet
check_go_vet() {
    print_section "Go Vet Analysis"
    
    cd "$PROJECT_ROOT"
    
    if go vet ./...; then
        echo -e "${GREEN}✅ Go vet passed${NC}"
        return 0
    else
        echo -e "${RED}❌ Go vet found issues${NC}"
        return 1
    fi
}

# Function to run Go tests
run_go_tests() {
    print_section "Go Tests"
    
    cd "$PROJECT_ROOT"
    
    echo "Running Go tests with race detection..."
    if go test -race -timeout=30s ./...; then
        echo -e "${GREEN}✅ All Go tests passed${NC}"
        return 0
    else
        echo -e "${RED}❌ Go tests failed${NC}"
        return 1
    fi
}

# Function to check Go cyclomatic complexity
check_go_complexity() {
    print_section "Go Complexity Check"
    
    cd "$PROJECT_ROOT"
    
    if command -v gocyclo >/dev/null 2>&1; then
        complex_funcs=$(gocyclo -over 15 .)
        if [ -n "$complex_funcs" ]; then
            echo -e "${RED}❌ Functions with high complexity (>15):${NC}"
            echo "$complex_funcs"
            echo -e "${YELLOW}Consider refactoring these functions${NC}"
            return 1
        else
            echo -e "${GREEN}✅ All functions have acceptable complexity${NC}"
            return 0
        fi
    else
        echo -e "${YELLOW}⚠️  gocyclo not found, skipping complexity check${NC}"
        return 0
    fi
}

# Function to run security scanning
run_security_scan() {
    print_section "Security Scan"
    
    cd "$PROJECT_ROOT"
    
    if command -v gosec >/dev/null 2>&1; then
        if gosec -quiet ./...; then
            echo -e "${GREEN}✅ Security scan passed${NC}"
            return 0
        else
            echo -e "${RED}❌ Security issues found${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠️  gosec not found, skipping security scan${NC}"
        return 0
    fi
}

# Function to check Python widget syntax
check_python_syntax() {
    print_section "Python Widget Syntax"
    
    cd "$PROJECT_ROOT"
    
    if [ -d "widgets" ]; then
        python_files=$(find widgets -name "*.py" -type f)
        if [ -n "$python_files" ]; then
            echo "Checking Python widget syntax..."
            for file in $python_files; do
                if python3 -m py_compile "$file"; then
                    echo -e "${GREEN}✅ $file syntax OK${NC}"
                else
                    echo -e "${RED}❌ $file has syntax errors${NC}"
                    return 1
                fi
            done
            return 0
        else
            echo -e "${YELLOW}⚠️  No Python widget files found${NC}"
            return 0
        fi
    else
        echo -e "${YELLOW}⚠️  widgets directory not found${NC}"
        return 0
    fi
}

# Function to check for secrets
check_secrets() {
    print_section "Secret Detection"
    
    cd "$PROJECT_ROOT"
    
    if command -v detect-secrets >/dev/null 2>&1; then
        if detect-secrets scan --baseline .secrets.baseline --exclude-files 'go\.sum$|go\.mod$|coverage\.out$'; then
            echo -e "${GREEN}✅ No new secrets detected${NC}"
            return 0
        else
            echo -e "${RED}❌ Potential secrets detected${NC}"
            echo -e "${YELLOW}Review the findings and update .secrets.baseline if needed${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠️  detect-secrets not found, skipping secret detection${NC}"
        return 0
    fi
}

# Function to validate file sizes
check_file_sizes() {
    print_section "File Size Check"
    
    cd "$PROJECT_ROOT"
    
    large_files=$(find . -type f -size +1000k -not -path "./.git/*" -not -path "./bin/*" -not -path "./*.tar.gz")
    if [ -n "$large_files" ]; then
        echo -e "${RED}❌ Large files detected (>1MB):${NC}"
        echo "$large_files"
        echo -e "${YELLOW}Consider using Git LFS for large files${NC}"
        return 1
    else
        echo -e "${GREEN}✅ No large files detected${NC}"
        return 0
    fi
}

# Function to check documentation
check_documentation() {
    print_section "Documentation Check"
    
    cd "$PROJECT_ROOT"
    
    # Check for basic documentation files
    docs_exist=true
    
    if [ ! -f "README.md" ]; then
        echo -e "${YELLOW}⚠️  README.md not found${NC}"
        docs_exist=false
    fi
    
    if [ ! -f "go.mod" ]; then
        echo -e "${RED}❌ go.mod not found${NC}"
        docs_exist=false
    fi
    
    if [ "$docs_exist" = true ]; then
        echo -e "${GREEN}✅ Basic documentation files present${NC}"
        return 0
    else
        return 1
    fi
}

# Function to run widget tests
test_widgets() {
    print_section "Widget Functional Tests"
    
    cd "$PROJECT_ROOT"
    
    if [ -d "widgets" ]; then
        widget_files=$(find widgets -name "*.py" -type f)
        if [ -n "$widget_files" ]; then
            echo "Testing widget execution..."
            failed_widgets=""
            
            for widget in $widget_files; do
                echo "Testing $widget..."
                if timeout 10s python3 "$widget" '{}' >/dev/null 2>&1; then
                    echo -e "${GREEN}✅ $widget executed successfully${NC}"
                else
                    echo -e "${RED}❌ $widget failed to execute${NC}"
                    failed_widgets="$failed_widgets $widget"
                fi
            done
            
            if [ -n "$failed_widgets" ]; then
                echo -e "${RED}❌ Failed widgets:$failed_widgets${NC}"
                return 1
            else
                echo -e "${GREEN}✅ All widgets executed successfully${NC}"
                return 0
            fi
        else
            echo -e "${YELLOW}⚠️  No Python widget files found${NC}"
            return 0
        fi
    else
        echo -e "${YELLOW}⚠️  widgets directory not found${NC}"
        return 0
    fi
}

# Function to display summary
show_summary() {
    echo
    echo -e "${BLUE}======================================${NC}"
    echo -e "${BLUE}VALIDATION SUMMARY${NC}"
    echo -e "${BLUE}======================================${NC}"
    
    if [ $total_errors -eq 0 ]; then
        echo -e "${GREEN}🎉 ALL CHECKS PASSED!${NC}"
        echo -e "${GREEN}Your code is ready for commit.${NC}"
    else
        echo -e "${RED}❌ $total_errors CHECK(S) FAILED${NC}"
        echo -e "${RED}Please fix the issues before committing.${NC}"
    fi
    
    echo
    echo "To run individual checks:"
    echo "  Go format:     gofmt -l ."
    echo "  Go vet:        go vet ./..."
    echo "  Go tests:      go test -race ./..."
    echo "  Security:      gosec ./..."
    echo "  Secrets:       detect-secrets scan"
    echo
    echo "To run all pre-commit hooks:"
    echo "  pre-commit run --all-files"
    echo
}

# Main validation function
main() {
    echo -e "${GREEN}🔍 Running Comprehensive Code Validation${NC}"
    echo "Project: E-Paper Dashboard"
    echo "Path: $PROJECT_ROOT"
    echo
    
    total_errors=0
    
    # Run all checks
    check_go_format || ((total_errors++))
    check_go_imports || ((total_errors++))
    check_go_vet || ((total_errors++))
    run_go_tests || ((total_errors++))
    check_go_complexity || ((total_errors++))
    run_security_scan || ((total_errors++))
    check_python_syntax || ((total_errors++))
    check_secrets || ((total_errors++))
    check_file_sizes || ((total_errors++))
    check_documentation || ((total_errors++))
    test_widgets || ((total_errors++))
    
    # Show summary
    show_summary
    
    # Exit with error if any checks failed
    exit $total_errors
}

# Show help if requested
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    cat << EOF
E-Paper Dashboard Commit Validation Script

Usage: $0 [options]

Options:
  -h, --help     Show this help message
  --no-tests     Skip running tests (faster validation)
  --format-only  Only check formatting and basic issues

This script runs comprehensive validation checks that mirror
the pre-commit hooks. Use it to validate your changes before
committing or to debug hook failures.

All checks must pass for the script to exit with status 0.
EOF
    exit 0
fi

# Handle options
if [ "$1" = "--no-tests" ]; then
    # Override test functions to skip
    run_go_tests() { echo -e "${YELLOW}⚠️  Skipping Go tests${NC}"; return 0; }
    test_widgets() { echo -e "${YELLOW}⚠️  Skipping widget tests${NC}"; return 0; }
elif [ "$1" = "--format-only" ]; then
    # Override complex checks to skip
    run_go_tests() { echo -e "${YELLOW}⚠️  Skipping Go tests${NC}"; return 0; }
    test_widgets() { echo -e "${YELLOW}⚠️  Skipping widget tests${NC}"; return 0; }
    check_go_complexity() { echo -e "${YELLOW}⚠️  Skipping complexity check${NC}"; return 0; }
    run_security_scan() { echo -e "${YELLOW}⚠️  Skipping security scan${NC}"; return 0; }
    check_secrets() { echo -e "${YELLOW}⚠️  Skipping secret detection${NC}"; return 0; }
fi

# Run main function
main