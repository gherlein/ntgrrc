#!/bin/bash

# CI Test Suite for ntgrrc
# Comprehensive testing script for continuous integration

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly BOLD='\033[1m'
readonly NC='\033[0m' # No Color

# Configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-80}
readonly GO_VERSION_MIN="1.23"
readonly TIMEOUT_TESTS="10m"
readonly TIMEOUT_BUILD="5m"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

log_step() {
    echo -e "\n${BOLD}${BLUE}==>${NC} ${BOLD}$*${NC}"
}

# Error handling
cleanup() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        log_error "CI pipeline failed with exit code $exit_code"
    fi
    
    # Clean up any background processes
    jobs -p | xargs -r kill 2>/dev/null || true
    
    # Return to original directory
    cd "${PROJECT_ROOT}"
    
    exit $exit_code
}

trap cleanup EXIT

# Utility functions
check_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "Required command '$1' not found"
        return 1
    fi
}

check_go_version() {
    local go_version
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    
    if ! printf '%s\n%s\n' "$GO_VERSION_MIN" "$go_version" | sort -V -C; then
        log_error "Go version $go_version is below minimum required version $GO_VERSION_MIN"
        return 1
    fi
    
    log_info "Go version: $go_version ✓"
}

# Main CI steps
step_environment_check() {
    log_step "Environment Check"
    
    # Check required commands
    check_command "go"
    check_command "git"
    check_command "make"
    check_command "bc"
    
    # Check Go version
    check_go_version
    
    # Check project structure
    if [ ! -f "${PROJECT_ROOT}/go.mod" ]; then
        log_error "go.mod not found in project root"
        return 1
    fi
    
    if [ ! -f "${PROJECT_ROOT}/Makefile" ]; then
        log_error "Makefile not found in project root"
        return 1
    fi
    
    # Environment info
    log_info "Environment:"
    log_info "  OS: $(uname -s) $(uname -r)"
    log_info "  Architecture: $(uname -m)"
    log_info "  Shell: $SHELL"
    log_info "  Working Directory: $(pwd)"
    log_info "  Git Commit: $(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
    
    log_success "Environment check passed"
}

step_dependency_management() {
    log_step "Dependency Management"
    
    # Clean any previous state
    log_info "Cleaning previous build artifacts..."
    timeout "${TIMEOUT_BUILD}" make clean
    
    # Download dependencies
    log_info "Downloading dependencies..."
    timeout "${TIMEOUT_BUILD}" make deps
    
    # Verify dependencies
    log_info "Verifying dependencies..."
    go mod verify
    
    # Check for vulnerabilities (if govulncheck is available)
    if command -v govulncheck &> /dev/null; then
        log_info "Checking for known vulnerabilities..."
        govulncheck ./... || log_warning "Vulnerability check failed or found issues"
    else
        log_warning "govulncheck not available, skipping vulnerability check"
    fi
    
    log_success "Dependencies ready"
}

step_code_quality() {
    log_step "Code Quality Checks"
    
    # Format check
    log_info "Checking code formatting..."
    if ! go fmt ./... | grep -q .; then
        log_success "Code is properly formatted"
    else
        log_error "Code formatting issues found. Run 'make fmt' to fix."
        return 1
    fi
    
    # Vet check
    log_info "Running go vet..."
    timeout "${TIMEOUT_BUILD}" go vet ./...
    
    # Check for common issues with staticcheck (if available)
    if command -v staticcheck &> /dev/null; then
        log_info "Running staticcheck..."
        staticcheck ./...
    else
        log_warning "staticcheck not available, skipping static analysis"
    fi
    
    # Check for ineffectual assignments (if available)
    if command -v ineffassign &> /dev/null; then
        log_info "Checking for ineffectual assignments..."
        ineffassign ./...
    fi
    
    log_success "Code quality checks passed"
}

step_build_tests() {
    log_step "Build Tests"
    
    # Test build for current platform
    log_info "Building for current platform..."
    timeout "${TIMEOUT_BUILD}" make build
    
    # Verify binary
    if [ -f "build/ntgrrc" ]; then
        log_info "Binary size: $(ls -lh build/ntgrrc | awk '{print $5}')"
        
        # Test basic functionality
        log_info "Testing basic functionality..."
        timeout 10s ./build/ntgrrc version || log_warning "Version command failed"
        timeout 10s ./build/ntgrrc --help > /dev/null
    else
        log_error "Binary not found after build"
        return 1
    fi
    
    # Development build with race detection
    log_info "Building development version with race detection..."
    timeout "${TIMEOUT_BUILD}" make build-dev
    
    log_success "Build tests passed"
}

step_unit_tests() {
    log_step "Unit Tests"
    
    log_info "Running unit tests..."
    timeout "${TIMEOUT_TESTS}" make test-unit
    
    log_success "Unit tests passed"
}

step_integration_tests() {
    log_step "Integration Tests"
    
    log_info "Running integration tests..."
    timeout "${TIMEOUT_TESTS}" make test-integration
    
    log_success "Integration tests passed"
}

step_race_detection() {
    log_step "Race Condition Detection"
    
    log_info "Running tests with race detection..."
    timeout "${TIMEOUT_TESTS}" make test-race
    
    log_success "Race detection tests passed"
}

step_coverage_analysis() {
    log_step "Coverage Analysis"
    
    log_info "Generating coverage report..."
    timeout "${TIMEOUT_TESTS}" make test-coverage
    
    # Check coverage threshold
    log_info "Checking coverage threshold (${COVERAGE_THRESHOLD}%)..."
    
    if [ -f "coverage/coverage.out" ]; then
        local coverage
        coverage=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        
        log_info "Current coverage: ${coverage}%"
        
        if (( $(echo "$coverage < $COVERAGE_THRESHOLD" | bc -l) )); then
            log_error "Coverage ${coverage}% is below threshold ${COVERAGE_THRESHOLD}%"
            
            # Show uncovered functions
            log_info "Functions with low coverage:"
            go tool cover -func=coverage/coverage.out | awk '$3 < 80 {print "  " $1 ": " $3}' | head -10
            
            return 1
        else
            log_success "Coverage ${coverage}% meets threshold ${COVERAGE_THRESHOLD}%"
        fi
        
        # Generate HTML report for local development
        if [ "${CI:-false}" != "true" ]; then
            log_info "Generating HTML coverage report..."
            make coverage-html
            log_info "Coverage report: file://$(pwd)/coverage/coverage.html"
        fi
    else
        log_error "Coverage report not found"
        return 1
    fi
}

step_cross_compilation() {
    log_step "Cross-Compilation Tests"
    
    log_info "Testing cross-compilation for all platforms..."
    timeout "${TIMEOUT_BUILD}" make cross-build
    
    # Verify all expected binaries were created
    local expected_binaries=(
        "ntgrrc-linux-amd64"
        "ntgrrc-linux-arm64"
        "ntgrrc-linux-386"
        "ntgrrc-darwin-amd64"
        "ntgrrc-darwin-arm64"
        "ntgrrc-windows-amd64.exe"
        "ntgrrc-windows-386.exe"
        "ntgrrc-freebsd-amd64"
    )
    
    local missing_binaries=()
    for binary in "${expected_binaries[@]}"; do
        if [ ! -f "dist/${binary}" ]; then
            missing_binaries+=("${binary}")
        fi
    done
    
    if [ ${#missing_binaries[@]} -gt 0 ]; then
        log_error "Missing binaries: ${missing_binaries[*]}"
        return 1
    fi
    
    # Show binary sizes
    log_info "Cross-compiled binary sizes:"
    ls -lh dist/ | grep -v "^total" | awk '{print "  " $9 ": " $5}'
    
    log_success "Cross-compilation tests passed"
}

step_security_checks() {
    log_step "Security Checks"
    
    # Check for hardcoded secrets (basic patterns)
    log_info "Checking for potential hardcoded secrets..."
    local secret_patterns=(
        "password.*=.*[\"'][^\"']{8,}[\"']"
        "token.*=.*[\"'][^\"']{20,}[\"']"
        "api[_-]?key.*=.*[\"'][^\"']{20,}[\"']"
        "secret.*=.*[\"'][^\"']{20,}[\"']"
    )
    
    local findings=0
    for pattern in "${secret_patterns[@]}"; do
        if grep -r -i -E "$pattern" --include="*.go" . 2>/dev/null; then
            log_warning "Potential hardcoded secret found (pattern: $pattern)"
            ((findings++))
        fi
    done
    
    if [ $findings -eq 0 ]; then
        log_success "No obvious hardcoded secrets found"
    else
        log_warning "Found $findings potential security issues (review manually)"
    fi
    
    # Check file permissions on sensitive files
    log_info "Checking file permissions..."
    local sensitive_files=("go.mod" "go.sum" "*.go")
    for pattern in "${sensitive_files[@]}"; do
        find . -name "$pattern" -type f -perm /o+w 2>/dev/null | while read -r file; do
            log_warning "World-writable file: $file"
        done
    done
}

step_performance_tests() {
    log_step "Performance Tests"
    
    if go test -list=. | grep -q "Benchmark"; then
        log_info "Running benchmark tests..."
        timeout "${TIMEOUT_TESTS}" make bench
    else
        log_info "No benchmark tests found, skipping performance tests"
    fi
    
    log_success "Performance tests completed"
}

step_documentation_check() {
    log_step "Documentation Check"
    
    # Check for required documentation files
    local required_docs=("README.md" "LICENSE")
    for doc in "${required_docs[@]}"; do
        if [ ! -f "$doc" ]; then
            log_error "Required documentation file missing: $doc"
            return 1
        fi
    done
    
    # Check if documentation is up to date with --help output
    if [ -f "build/ntgrrc" ]; then
        log_info "Checking if help documentation is current..."
        
        # Extract help from binary and README
        local help_output
        help_output=$(timeout 10s ./build/ntgrrc --help 2>&1 || echo "Help failed")
        
        # Basic check - ensure README mentions main commands
        local main_commands=("login" "poe" "port" "version")
        for cmd in "${main_commands[@]}"; do
            if ! grep -q "$cmd" README.md; then
                log_warning "Command '$cmd' not documented in README.md"
            fi
        done
    fi
    
    log_success "Documentation check completed"
}

step_final_validation() {
    log_step "Final Validation"
    
    # Run a comprehensive smoke test
    log_info "Running final smoke test..."
    
    if [ -f "build/ntgrrc" ]; then
        # Test version command
        local version_output
        version_output=$(timeout 10s ./build/ntgrrc version 2>&1)
        if [[ "$version_output" =~ ^(dev|v[0-9]+\.[0-9]+\.[0-9]+) ]]; then
            log_success "Version command working: $version_output"
        else
            log_error "Version command returned unexpected output: $version_output"
            return 1
        fi
        
        # Test help command
        if timeout 10s ./build/ntgrrc --help > /dev/null 2>&1; then
            log_success "Help command working"
        else
            log_error "Help command failed"
            return 1
        fi
    fi
    
    # Verify clean git state (if in git repo)
    if git rev-parse --git-dir > /dev/null 2>&1; then
        if [ -n "$(git status --porcelain)" ]; then
            log_warning "Git working directory is not clean"
            git status --short
        else
            log_success "Git working directory is clean"
        fi
    fi
    
    log_success "Final validation passed"
}

# Progress tracking
print_progress() {
    local current=$1
    local total=$2
    local step_name=$3
    
    local percent=$((current * 100 / total))
    printf "\n${BOLD}Progress: [%d/%d] %d%% - %s${NC}\n" "$current" "$total" "$percent" "$step_name"
}

# Main execution
main() {
    local start_time
    start_time=$(date +%s)
    
    echo -e "${BOLD}${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════════════════════╗"
    echo "║                           ntgrrc CI Test Suite                              ║"
    echo "╚══════════════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    log_info "Starting CI pipeline at $(date)"
    log_info "Project: ntgrrc - Netgear Remote Control CLI"
    log_info "Coverage threshold: ${COVERAGE_THRESHOLD}%"
    
    # Change to project root
    cd "${PROJECT_ROOT}"
    
    # Define all steps
    local steps=(
        "step_environment_check:Environment Check"
        "step_dependency_management:Dependency Management"
        "step_code_quality:Code Quality"
        "step_build_tests:Build Tests"
        "step_unit_tests:Unit Tests"
        "step_integration_tests:Integration Tests"
        "step_race_detection:Race Detection"
        "step_coverage_analysis:Coverage Analysis"
        "step_cross_compilation:Cross Compilation"
        "step_security_checks:Security Checks"
        "step_performance_tests:Performance Tests"
        "step_documentation_check:Documentation Check"
        "step_final_validation:Final Validation"
    )
    
    local total_steps=${#steps[@]}
    local current_step=0
    
    # Execute all steps
    for step_info in "${steps[@]}"; do
        local step_func="${step_info%%:*}"
        local step_name="${step_info##*:}"
        
        ((current_step++))
        print_progress $current_step $total_steps "$step_name"
        
        if ! $step_func; then
            log_error "Step failed: $step_name"
            return 1
        fi
    done
    
    # Success summary
    local end_time
    end_time=$(date +%s)
    local duration=$((end_time - start_time))
    local minutes=$((duration / 60))
    local seconds=$((duration % 60))
    
    echo -e "\n${BOLD}${GREEN}"
    echo "╔══════════════════════════════════════════════════════════════════════════════╗"
    echo "║                             CI PIPELINE PASSED                              ║"
    echo "╚══════════════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    log_success "All CI checks completed successfully!"
    log_success "Total execution time: ${minutes}m ${seconds}s"
    log_success "Ready for deployment/release"
    
    # Show final summary
    echo -e "\n${BOLD}Summary:${NC}"
    echo "  ✓ Environment validated"
    echo "  ✓ Dependencies verified"
    echo "  ✓ Code quality passed"
    echo "  ✓ Build successful"
    echo "  ✓ All tests passed"
    echo "  ✓ Race conditions checked"
    echo "  ✓ Coverage threshold met"
    echo "  ✓ Cross-compilation successful"
    echo "  ✓ Security checks completed"
    echo "  ✓ Performance validated"
    echo "  ✓ Documentation verified"
    echo "  ✓ Final validation passed"
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi