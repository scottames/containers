#!/usr/bin/env bash
# vet-deps.sh - Scan project dependencies for vulnerabilities and malicious packages
#
# Usage:
#   vet-deps.sh [project-directory]
#   vet-deps.sh .
#   vet-deps.sh ~/projects/my-app
#
# Requirements:
#   For Go projects:
#     - govulncheck: go install golang.org/x/vuln/cmd/govulncheck@latest
#     - osv-scanner: go install github.com/google/osv-scanner/cmd/osv-scanner@latest
#     - scorecard: go install github.com/ossf/scorecard/v5/cmd/scorecard@latest
#
#   For Rust projects:
#     - cargo-audit: cargo install cargo-audit
#     - cargo-deny: cargo install cargo-deny (optional)
#     - osv-scanner: go install github.com/google/osv-scanner/cmd/osv-scanner@latest
#
# Optional:
#   - jq: for JSON parsing (highly recommended)

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCORECARD_DEP_LIMIT="${SCORECARD_DEP_LIMIT:-5}"       # Max deps to run scorecard on
SCORECARD_THRESHOLD="${SCORECARD_THRESHOLD:-5}"       # Min acceptable score
SKIP_SCORECARD="${SKIP_SCORECARD:-false}"             # Skip slow scorecard checks
VERBOSE="${VERBOSE:-false}"

log_info() { echo -e "${BLUE}ℹ${NC} $*"; }
log_ok() { echo -e "${GREEN}✓${NC} $*"; }
log_warn() { echo -e "${YELLOW}⚠${NC} $*"; }
log_error() { echo -e "${RED}✗${NC} $*"; }
log_section() { echo -e "\n${CYAN}━━━${NC} $* ${CYAN}━━━${NC}"; }
log_subsection() { echo -e "\n${BLUE}→${NC} $*"; }

# Tools that can be installed via mise
MISE_TOOLS=(scorecard osv-scanner trivy gitleaks gosec govulncheck)
MISSING_MISE_TOOLS=()

check_command() {
  if ! command -v "$1" &>/dev/null; then
    log_warn "Command '$1' not found - skipping related checks"
    # Track if this is a mise-installable tool
    for mise_tool in "${MISE_TOOLS[@]}"; do
      if [[ "$1" == "$mise_tool" ]]; then
        MISSING_MISE_TOOLS+=("$1")
        break
      fi
    done
    return 1
  fi
  return 0
}

print_mise_reminder() {
  if [[ ${#MISSING_MISE_TOOLS[@]} -gt 0 ]]; then
    echo ""
    log_warn "Some security tools are not installed: ${MISSING_MISE_TOOLS[*]}"
    log_info "Install them with mise:"
    echo -e "  ${BLUE}mise install --config /etc/security-tools/mise-security-tools.toml${NC}"
    echo ""
  fi
}

# Track overall status
VULNS_FOUND=0
WARNINGS_FOUND=0

run_scorecard_on_deps() {
  local deps=("$@")
  
  if [[ "$SKIP_SCORECARD" == "true" ]]; then
    log_info "Skipping Scorecard checks (SKIP_SCORECARD=true)"
    return
  fi
  
  if ! check_command scorecard; then
    return
  fi
  
  log_subsection "Scorecard Analysis (top $SCORECARD_DEP_LIMIT dependencies)"
  
  local count=0
  for dep in "${deps[@]}"; do
    [[ $count -ge $SCORECARD_DEP_LIMIT ]] && break
    
    # Only process GitHub URLs
    if [[ "$dep" != *github.com* ]]; then
      continue
    fi
    
    # Normalize to HTTPS URL
    local repo_url="https://github.com/$(echo "$dep" | sed 's|.*github.com/||' | cut -d'/' -f1-2)"
    
    echo -e "\n  ${BLUE}$repo_url${NC}"
    
    local result
    if result=$(scorecard --repo="$repo_url" --checks=Maintained,Vulnerabilities,Code-Review,Branch-Protection --format=json 2>/dev/null); then
      local score
      score=$(echo "$result" | jq -r '.score // 0')
      
      if (( $(echo "$score < $SCORECARD_THRESHOLD" | bc -l) )); then
        log_warn "  Score: $score/10 (below threshold)"
        ((WARNINGS_FOUND++)) || true
      else
        echo "    Score: $score/10"
      fi
      
      # Show individual check scores
      echo "$result" | jq -r '.checks[] | "    \(.name): \(.score)/10\(if .score < 5 then " ⚠" else "" end)"'
    else
      echo "    (could not fetch scorecard)"
    fi
    
    ((count++)) || true
  done
  
  if [[ ${#deps[@]} -gt $SCORECARD_DEP_LIMIT ]]; then
    log_info "Showing $SCORECARD_DEP_LIMIT of ${#deps[@]} dependencies. Set SCORECARD_DEP_LIMIT to check more."
  fi
}

scan_go_project() {
  log_section "Go Project Analysis"
  
  local deps=()
  
  # govulncheck - official Go vulnerability scanner
  if check_command govulncheck; then
    log_subsection "govulncheck (official Go vulnerability database)"
    if ! govulncheck ./... 2>&1; then
      ((VULNS_FOUND++)) || true
    else
      log_ok "No vulnerabilities found"
    fi
  fi
  
  # OSV Scanner - broader vulnerability database
  if check_command osv-scanner; then
    log_subsection "OSV Scanner"
    if [[ -f "go.sum" ]]; then
      if ! osv-scanner --lockfile=go.sum 2>&1; then
        ((VULNS_FOUND++)) || true
      else
        log_ok "No vulnerabilities found in OSV database"
      fi
    else
      log_warn "go.sum not found - run 'go mod tidy' first"
    fi
  fi
  
  # Collect dependencies for Scorecard
  if check_command go && check_command jq; then
    mapfile -t deps < <(go list -m -json all 2>/dev/null | jq -r 'select(.Main != true) | .Path' | grep github.com || true)
  fi
  
  # Run Scorecard on top dependencies
  if [[ ${#deps[@]} -gt 0 ]]; then
    run_scorecard_on_deps "${deps[@]}"
  fi
  
  # Check for replace directives (potential supply chain risk)
  log_subsection "Supply Chain Checks"
  if grep -q "^replace" go.mod 2>/dev/null; then
    log_warn "Found 'replace' directives in go.mod - verify these are intentional:"
    grep "^replace" go.mod | head -5
    ((WARNINGS_FOUND++)) || true
  else
    log_ok "No replace directives found"
  fi
  
  # Check for retracted versions
  if go list -m -retracted all 2>/dev/null | grep -v "^$" | head -5 | grep -q .; then
    log_warn "Some dependencies have retracted versions available:"
    go list -m -retracted all 2>/dev/null | grep -v "^$" | head -5
    ((WARNINGS_FOUND++)) || true
  fi
}

scan_rust_project() {
  log_section "Rust Project Analysis"
  
  local deps=()
  
  # cargo-audit - RustSec advisory database
  if check_command cargo-audit || check_command cargo; then
    log_subsection "cargo-audit (RustSec advisory database)"
    if ! cargo audit 2>&1; then
      ((VULNS_FOUND++)) || true
    else
      log_ok "No vulnerabilities found"
    fi
  fi
  
  # cargo-deny - comprehensive policy checks
  if command -v cargo-deny &>/dev/null || cargo deny --version &>/dev/null 2>&1; then
    log_subsection "cargo-deny (license, ban, and source checks)"
    
    if [[ -f "deny.toml" ]]; then
      if ! cargo deny check 2>&1; then
        ((VULNS_FOUND++)) || true
      else
        log_ok "All cargo-deny checks passed"
      fi
    else
      log_info "No deny.toml found - running with defaults"
      log_info "Consider creating deny.toml for policy enforcement"
      cargo deny check advisories 2>&1 || ((VULNS_FOUND++)) || true
    fi
  fi
  
  # OSV Scanner
  if check_command osv-scanner; then
    log_subsection "OSV Scanner"
    if [[ -f "Cargo.lock" ]]; then
      if ! osv-scanner --lockfile=Cargo.lock 2>&1; then
        ((VULNS_FOUND++)) || true
      else
        log_ok "No vulnerabilities found in OSV database"
      fi
    else
      log_warn "Cargo.lock not found - run 'cargo generate-lockfile' first"
    fi
  fi
  
  # Collect dependencies for Scorecard
  if check_command cargo && check_command jq; then
    mapfile -t deps < <(cargo metadata --format-version 1 2>/dev/null | \
      jq -r '.packages[] | select(.source != null) | .repository // empty' | \
      grep github.com | sort -u || true)
  fi
  
  # Run Scorecard on top dependencies
  if [[ ${#deps[@]} -gt 0 ]]; then
    run_scorecard_on_deps "${deps[@]}"
  fi
  
  # Check for git dependencies (potential supply chain risk)
  log_subsection "Supply Chain Checks"
  if grep -q '\[dependencies\.' Cargo.toml 2>/dev/null && grep -A5 '\[dependencies' Cargo.toml | grep -q "git\s*="; then
    log_warn "Found git dependencies in Cargo.toml - verify these are intentional:"
    grep -B1 "git\s*=" Cargo.toml | head -10
    ((WARNINGS_FOUND++)) || true
  else
    log_ok "No git dependencies found"
  fi
  
  # Check for path dependencies
  if grep -q 'path\s*=' Cargo.toml 2>/dev/null; then
    log_info "Found path dependencies (normal for workspaces):"
    grep 'path\s*=' Cargo.toml | head -5
  fi
}

scan_node_project() {
  log_section "Node.js Project Analysis"
  
  # Socket CLI scan (preferred - comprehensive behavioral analysis)
  if check_command socket; then
    log_subsection "Socket.dev Security Scan"
    
    # socket scan create provides full analysis
    if socket scan create . --dry-run 2>&1; then
      log_ok "Socket scan completed"
    else
      # Fallback to wrapper check if scan fails
      log_info "Running Socket npm audit..."
      socket npm audit 2>&1 || ((VULNS_FOUND++)) || true
    fi
  # Fallback to standard npm audit
  elif check_command npm; then
    log_subsection "npm audit"
    if ! npm audit 2>&1; then
      ((VULNS_FOUND++)) || true
    else
      log_ok "No vulnerabilities found"
    fi
  fi
  
  # OSV Scanner
  if check_command osv-scanner; then
    log_subsection "OSV Scanner"
    if [[ -f "package-lock.json" ]]; then
      if ! osv-scanner --lockfile=package-lock.json 2>&1; then
        ((VULNS_FOUND++)) || true
      fi
    elif [[ -f "yarn.lock" ]]; then
      if ! osv-scanner --lockfile=yarn.lock 2>&1; then
        ((VULNS_FOUND++)) || true
      fi
    elif [[ -f "pnpm-lock.yaml" ]]; then
      if ! osv-scanner --lockfile=pnpm-lock.yaml 2>&1; then
        ((VULNS_FOUND++)) || true
      fi
    fi
  fi
  
  # GuardDog fallback if Socket not available
  if ! check_command socket && check_command guarddog && [[ -f "package.json" ]]; then
    log_subsection "GuardDog (malicious package detection)"
    guarddog npm verify package.json 2>&1 || ((VULNS_FOUND++)) || true
  fi
}

print_summary() {
  log_section "Summary"
  
  if [[ $VULNS_FOUND -gt 0 ]]; then
    log_error "Vulnerabilities found: $VULNS_FOUND"
  else
    log_ok "No vulnerabilities found"
  fi
  
  if [[ $WARNINGS_FOUND -gt 0 ]]; then
    log_warn "Warnings: $WARNINGS_FOUND"
  fi
  
  echo ""
  if [[ $VULNS_FOUND -gt 0 || $WARNINGS_FOUND -gt 0 ]]; then
    log_info "Review the issues above before deploying"
    return 1
  else
    log_ok "All checks passed"
    return 0
  fi
}

print_tool_status() {
  log_section "Tool Status"
  
  local tools=(
    "scorecard:OpenSSF Scorecard:go install github.com/ossf/scorecard/v5/cmd/scorecard@latest"
    "osv-scanner:OSV Scanner:go install github.com/google/osv-scanner/cmd/osv-scanner@latest"
    "govulncheck:Go Vulncheck:go install golang.org/x/vuln/cmd/govulncheck@latest"
    "cargo-audit:Cargo Audit:cargo install cargo-audit"
    "cargo-deny:Cargo Deny:cargo install cargo-deny"
    "guarddog:GuardDog:pip install guarddog"
  )
  
  for tool_info in "${tools[@]}"; do
    IFS=':' read -r cmd name install <<< "$tool_info"
    if command -v "$cmd" &>/dev/null; then
      echo -e "  ${GREEN}✓${NC} $name"
    else
      echo -e "  ${YELLOW}○${NC} $name (install: $install)"
    fi
  done
}

usage() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS] [project-directory]

Scan project dependencies for vulnerabilities and security issues.

Arguments:
  project-directory    Path to project (default: current directory)

Options:
  -h, --help          Show this help message
  -v, --verbose       Verbose output
  --status            Show tool installation status
  --skip-scorecard    Skip Scorecard checks (faster)

Environment Variables:
  SCORECARD_DEP_LIMIT   Max dependencies to run Scorecard on (default: 5)
  SCORECARD_THRESHOLD   Minimum acceptable Scorecard score (default: 5)
  SKIP_SCORECARD        Skip Scorecard checks if 'true'
  VERBOSE               Enable verbose output if 'true'

Examples:
  $(basename "$0")                    # Scan current directory
  $(basename "$0") ~/projects/myapp   # Scan specific project
  $(basename "$0") --skip-scorecard   # Fast scan without Scorecard
  SCORECARD_DEP_LIMIT=10 $(basename "$0")  # Check more dependencies
EOF
  exit 0
}

main() {
  local project_dir="."
  
  # Parse arguments
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h|--help)
        usage
        ;;
      -v|--verbose)
        VERBOSE=true
        shift
        ;;
      --status)
        print_tool_status
        exit 0
        ;;
      --skip-scorecard)
        SKIP_SCORECARD=true
        shift
        ;;
      -*)
        log_error "Unknown option: $1"
        usage
        ;;
      *)
        project_dir="$1"
        shift
        ;;
    esac
  done
  
  # Change to project directory
  if [[ ! -d "$project_dir" ]]; then
    log_error "Directory not found: $project_dir"
    exit 1
  fi
  
  cd "$project_dir"
  
  echo -e "${CYAN}╔════════════════════════════════════════════════════════════════╗${NC}"
  echo -e "${CYAN}║${NC}          Dependency Security Scan: $(basename "$(pwd)")          ${CYAN}║${NC}"
  echo -e "${CYAN}╚════════════════════════════════════════════════════════════════╝${NC}"
  echo ""
  log_info "Scanning: $(pwd)"
  
  # Detect and scan project types
  local project_found=false
  
  if [[ -f "go.mod" ]]; then
    project_found=true
    scan_go_project
  fi
  
  if [[ -f "Cargo.toml" ]]; then
    project_found=true
    scan_rust_project
  fi
  
  if [[ -f "package.json" ]]; then
    project_found=true
    scan_node_project
  fi
  
  if [[ "$project_found" == "false" ]]; then
    log_error "No supported project files found (go.mod, Cargo.toml, package.json)"
    log_info "Supported project types: Go, Rust, Node.js"
    exit 1
  fi

  print_mise_reminder
  print_summary
}

main "$@"
