#!/usr/bin/env bash
# vet-tool.sh - Check a GitHub repo or npm package before installing
#
# Usage:
#   vet-tool.sh https://github.com/junegunn/fzf
#   vet-tool.sh npm:prettier
#   vet-tool.sh go:github.com/some/module
#
# Requirements:
#   - scorecard: go install github.com/ossf/scorecard/v5/cmd/scorecard@latest
#   - guarddog: pip install guarddog (for npm packages)
#   - jq: for JSON parsing
#
# Optional:
#   - Clone malicious packages DB:
#     git clone --depth 1 https://github.com/ossf/malicious-packages ~/.local/share/malicious-packages

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

MALICIOUS_DB="${MALICIOUS_PACKAGES_DB:-$HOME/.local/share/malicious-packages}"
SCORECARD_THRESHOLD="${SCORECARD_THRESHOLD:-5}"
DEEP_SCAN="${DEEP_SCAN:-false}"
KEEP_CLONE="${KEEP_CLONE:-false}"

usage() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS] <target>

Targets:
  https://github.com/owner/repo    GitHub repository URL
  github.com/owner/repo            GitHub repository (without https)
  npm:<package-name>               npm package
  pypi:<package-name>              PyPI package
  go:<module-path>                 Go module (github.com/...)

Options:
  --deep          Clone repo and scan dependencies (for Go/Rust binaries)
  --keep          Keep cloned repo after scan (with --deep)
  -h, --help      Show this help

Environment Variables:
  SCORECARD_THRESHOLD   Minimum acceptable score (default: 5)
  MALICIOUS_PACKAGES_DB Path to local malicious-packages clone
  DEEP_SCAN             Set to 'true' for deep scanning
  KEEP_CLONE            Set to 'true' to keep cloned repos

Examples:
  $(basename "$0") https://github.com/junegunn/fzf
  $(basename "$0") --deep https://github.com/charmbracelet/gum
  $(basename "$0") npm:prettier
  $(basename "$0") pypi:requests
  $(basename "$0") go:github.com/charmbracelet/gum
EOF
  exit 1
}

log_info() { echo -e "${BLUE}ℹ${NC} $*"; }
log_ok() { echo -e "${GREEN}✓${NC} $*"; }
log_warn() { echo -e "${YELLOW}⚠${NC} $*"; }
log_error() { echo -e "${RED}✗${NC} $*"; }
log_section() { echo -e "\n${BLUE}===${NC} $* ${BLUE}===${NC}"; }
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

run_scorecard() {
  local repo_url="$1"
  local checks="${2:-}"
  
  if ! check_command scorecard; then
    return 0
  fi
  
  log_section "OpenSSF Scorecard"
  log_info "Analyzing: $repo_url"
  
  local scorecard_args=(--repo="$repo_url" --format=json)
  if [[ -n "$checks" ]]; then
    scorecard_args+=(--checks="$checks")
  fi
  
  local result
  if ! result=$(scorecard "${scorecard_args[@]}" 2>/dev/null); then
    log_warn "Scorecard analysis failed (repo may not exist or be private)"
    return 0
  fi
  
  local score
  score=$(echo "$result" | jq -r '.score // 0')
  
  echo -e "\nOverall Score: ${score}/10"
  
  if (( $(echo "$score < $SCORECARD_THRESHOLD" | bc -l) )); then
    log_warn "Score below threshold ($SCORECARD_THRESHOLD)"
  else
    log_ok "Score meets threshold ($SCORECARD_THRESHOLD)"
  fi
  
  echo -e "\nCheck Results:"
  echo "$result" | jq -r '.checks[] | "  \(.name): \(.score)/10\(if .score < 5 then " ⚠" else "" end)"'
  
  # Show reasons for low-scoring checks
  local low_scores
  low_scores=$(echo "$result" | jq -r '.checks[] | select(.score < 5) | "  → \(.name): \(.reason)"')
  if [[ -n "$low_scores" ]]; then
    echo -e "\nLow Score Details:"
    echo "$low_scores"
  fi
}

check_malicious_db() {
  local search_term="$1"
  
  log_section "Malicious Packages Database"
  
  if [[ ! -d "$MALICIOUS_DB" ]]; then
    log_info "Local malicious packages DB not found at: $MALICIOUS_DB"
    log_info "Clone with: git clone --depth 1 https://github.com/ossf/malicious-packages $MALICIOUS_DB"
    return 0
  fi
  
  log_info "Searching for: $search_term"
  
  if grep -ri "$search_term" "$MALICIOUS_DB" --include="*.json" --include="*.yaml" &>/dev/null; then
    log_error "FOUND IN MALICIOUS PACKAGES DATABASE!"
    grep -ri "$search_term" "$MALICIOUS_DB" --include="*.json" --include="*.yaml" -l | head -5
    return 1
  else
    log_ok "Not found in malicious packages database"
  fi
}

deep_scan_repo() {
  local repo_url="$1"
  local repo_path="$2"
  
  log_section "Deep Dependency Scan"
  
  # Create temp directory for clone
  local clone_dir
  clone_dir=$(mktemp -d -t vet-tool-XXXXXX)
  
  log_info "Cloning repository to $clone_dir..."
  if ! git clone --depth 1 "$repo_url" "$clone_dir" 2>/dev/null; then
    log_error "Failed to clone repository"
    rm -rf "$clone_dir"
    return 1
  fi
  
  cd "$clone_dir" || return 1
  
  # ==========================================================================
  # Source Code Security Scanning (language-agnostic)
  # ==========================================================================
  log_section "Source Code Analysis"
  
  # Semgrep - static analysis with security rules
  if check_command semgrep; then
    log_subsection "Semgrep Security Scan"
    log_info "Running security-focused static analysis..."
    
    # Run with security-audit and supply-chain rules
    semgrep --config "p/security-audit" --config "p/secrets" --config "p/supply-chain" \
      --metrics=off --quiet --no-git-ignore \
      --exclude="vendor" --exclude="node_modules" --exclude="*.min.js" \
      . 2>&1 || log_warn "Semgrep found issues (review above)"
  else
    log_info "Semgrep not installed - skipping static analysis"
    log_info "Install: pip install semgrep (or brew install semgrep)"
  fi
  
  # Gitleaks - secrets detection
  if check_command gitleaks; then
    log_subsection "Secrets Detection (Gitleaks)"
    if gitleaks detect --source . --no-git --quiet 2>/dev/null; then
      log_ok "No secrets detected"
    else
      log_warn "Potential secrets found:"
      gitleaks detect --source . --no-git --verbose 2>&1 | head -20
    fi
  fi
  
  # Trivy filesystem scan
  if check_command trivy; then
    log_subsection "Trivy Filesystem Scan"
    trivy fs --security-checks vuln,secret,misconfig --severity HIGH,CRITICAL \
      --quiet . 2>&1 || log_warn "Trivy found issues"
  fi
  
  # ==========================================================================
  # Suspicious Pattern Detection
  # ==========================================================================
  log_subsection "Suspicious Pattern Check"
  
  local suspicious_found=false
  
  # Check for obfuscated/encoded strings (common in malware)
  log_info "Checking for suspicious patterns..."
  
  # Base64 encoded blocks (potential hidden payloads)
  local base64_hits
  base64_hits=$(grep -rE '[A-Za-z0-9+/]{50,}={0,2}' --include="*.go" --include="*.rs" --include="*.js" --include="*.py" . 2>/dev/null | wc -l)
  if [[ $base64_hits -gt 5 ]]; then
    log_warn "Found $base64_hits potential base64 encoded strings"
    suspicious_found=true
  fi
  
  # Eval/exec patterns
  if grep -rE '\b(eval|exec|os\.system|subprocess\.call|child_process)\s*\(' \
    --include="*.py" --include="*.js" --include="*.ts" . 2>/dev/null | grep -v "test" | head -5 | grep -q .; then
    log_warn "Found eval/exec patterns:"
    grep -rE '\b(eval|exec|os\.system|subprocess\.call|child_process)\s*\(' \
      --include="*.py" --include="*.js" --include="*.ts" . 2>/dev/null | grep -v "test" | head -5
    suspicious_found=true
  fi
  
  # Network calls in unexpected places (install scripts, build scripts)
  if grep -rE '(curl|wget|http\.Get|reqwest::get|fetch\()' \
    setup.py Makefile build.rs build.go install.sh 2>/dev/null | head -5 | grep -q .; then
    log_warn "Found network calls in build/install scripts:"
    grep -rE '(curl|wget|http\.Get|reqwest::get|fetch\()' \
      setup.py Makefile build.rs build.go install.sh 2>/dev/null | head -5
    suspicious_found=true
  fi
  
  # Environment variable access (credential harvesting)
  if grep -rE '(os\.environ|std::env::var|process\.env)\[.*(KEY|SECRET|TOKEN|PASS|CRED|AUTH)' \
    --include="*.py" --include="*.rs" --include="*.go" --include="*.js" . 2>/dev/null | head -5 | grep -q .; then
    log_warn "Found sensitive env var access:"
    grep -rE '(os\.environ|std::env::var|process\.env)\[.*(KEY|SECRET|TOKEN|PASS|CRED|AUTH)' \
      --include="*.py" --include="*.rs" --include="*.go" --include="*.js" . 2>/dev/null | head -5
    suspicious_found=true
  fi
  
  # Suspicious domains/IPs
  if grep -rE '(pastebin\.com|ngrok\.io|burpcollaborator|[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})' \
    --include="*.go" --include="*.rs" --include="*.py" --include="*.js" . 2>/dev/null | grep -v "test" | grep -v "127.0.0.1" | grep -v "0.0.0.0" | head -5 | grep -q .; then
    log_warn "Found suspicious domains/IPs:"
    grep -rE '(pastebin\.com|ngrok\.io|burpcollaborator|[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})' \
      --include="*.go" --include="*.rs" --include="*.py" --include="*.js" . 2>/dev/null | grep -v "test" | grep -v "127.0.0.1" | grep -v "0.0.0.0" | head -5
    suspicious_found=true
  fi
  
  if [[ "$suspicious_found" == "false" ]]; then
    log_ok "No obvious suspicious patterns detected"
  fi
  
  # ==========================================================================
  # Language-Specific Security Analysis
  # ==========================================================================
  local found_project=false
  
  # Check for Go project
  if [[ -f "go.mod" ]]; then
    found_project=true
    log_section "Go Project Analysis"
    
    # gosec - Go security checker
    if check_command gosec; then
      log_subsection "gosec Security Scan"
      gosec -quiet -fmt text ./... 2>&1 || log_warn "gosec found issues"
    fi
    
    # govulncheck
    if check_command govulncheck; then
      log_subsection "govulncheck"
      govulncheck ./... 2>&1 || log_warn "Vulnerabilities found"
    fi
    
    # OSV Scanner
    if check_command osv-scanner && [[ -f "go.sum" ]]; then
      log_subsection "OSV Scanner"
      osv-scanner --lockfile=go.sum 2>&1 || log_warn "Vulnerabilities found in OSV"
    fi
    
    # Check for suspicious replace directives
    if grep -q "^replace" go.mod 2>/dev/null; then
      log_warn "Found 'replace' directives in go.mod:"
      grep "^replace" go.mod | head -5
    fi
    
    # List direct dependencies
    log_info "Direct dependencies:"
    go list -m -f '{{if not .Indirect}}  {{.Path}}{{end}}' all 2>/dev/null | grep -v "^$" | head -10
  fi
  
  # Check for Rust project
  if [[ -f "Cargo.toml" ]]; then
    found_project=true
    log_section "Rust Project Analysis"
    
    # Generate Cargo.lock if missing
    if [[ ! -f "Cargo.lock" ]]; then
      log_info "Generating Cargo.lock..."
      cargo generate-lockfile 2>/dev/null || true
    fi
    
    # cargo-geiger - detect unsafe code
    if check_command cargo-geiger || cargo geiger --version &>/dev/null 2>&1; then
      log_subsection "cargo-geiger (unsafe code detection)"
      cargo geiger --all-features --quiet 2>&1 | head -30 || log_info "cargo-geiger scan complete"
    fi
    
    # cargo-audit
    if check_command cargo-audit || cargo audit --version &>/dev/null 2>&1; then
      log_subsection "cargo-audit"
      cargo audit 2>&1 || log_warn "Vulnerabilities found"
    fi
    
    # cargo-deny (if deny.toml exists or just check advisories)
    if command -v cargo-deny &>/dev/null || cargo deny --version &>/dev/null 2>&1; then
      log_subsection "cargo-deny"
      if [[ -f "deny.toml" ]]; then
        cargo deny check 2>&1 || log_warn "Policy violations found"
      else
        cargo deny check advisories 2>&1 || log_warn "Advisories found"
      fi
    fi
    
    # OSV Scanner
    if check_command osv-scanner && [[ -f "Cargo.lock" ]]; then
      log_subsection "OSV Scanner"
      osv-scanner --lockfile=Cargo.lock 2>&1 || log_warn "Vulnerabilities found in OSV"
    fi
    
    # Check for git dependencies
    if grep -q 'git\s*=' Cargo.toml 2>/dev/null; then
      log_warn "Found git dependencies in Cargo.toml:"
      grep -B1 "git\s*=" Cargo.toml | head -10
    fi
    
    # Check for build.rs (can run arbitrary code at compile time)
    if [[ -f "build.rs" ]]; then
      log_warn "Project has build.rs - review for suspicious build-time code"
      log_info "Checking build.rs for network/process calls..."
      if grep -E '(Command::new|std::process|reqwest|curl|wget)' build.rs 2>/dev/null; then
        log_warn "build.rs contains process/network calls - review carefully"
      fi
    fi
    
    # List direct dependencies
    log_info "Direct dependencies:"
    if check_command cargo; then
      cargo metadata --format-version 1 --no-deps 2>/dev/null | \
        jq -r '.packages[0].dependencies[] | "  \(.name) \(.req // "")"' 2>/dev/null | head -10 || true
    fi
  fi
  
  # Check for Node project
  if [[ -f "package.json" ]]; then
    found_project=true
    log_section "Node.js Project Analysis"
    
    # Check for suspicious npm scripts
    log_subsection "npm Scripts Check"
    if check_command jq; then
      local scripts
      scripts=$(jq -r '.scripts // {} | to_entries[] | "\(.key): \(.value)"' package.json 2>/dev/null)
      
      if echo "$scripts" | grep -iE '(curl|wget|nc |netcat|bash -c|sh -c|eval|node -e)' | head -5 | grep -q .; then
        log_warn "Suspicious npm scripts found:"
        echo "$scripts" | grep -iE '(curl|wget|nc |netcat|bash -c|sh -c|eval|node -e)' | head -5
      else
        log_ok "No suspicious npm scripts detected"
      fi
      
      # Check for preinstall/postinstall hooks
      if echo "$scripts" | grep -E '^(preinstall|postinstall|preuninstall|postuninstall):' | grep -q .; then
        log_warn "Install hooks found (review carefully):"
        echo "$scripts" | grep -E '^(preinstall|postinstall|preuninstall|postuninstall):'
      fi
    fi
    
    # Socket scan
    if check_command socket; then
      log_subsection "Socket.dev Scan"
      socket scan create . --dry-run 2>&1 || true
    elif check_command npm; then
      log_subsection "npm audit"
      npm audit 2>&1 || log_warn "Vulnerabilities found"
    fi
  fi
  
  # Check for Python project
  if [[ -f "setup.py" ]] || [[ -f "pyproject.toml" ]]; then
    found_project=true
    log_section "Python Project Analysis"
    
    # Check setup.py for suspicious code
    if [[ -f "setup.py" ]]; then
      log_subsection "setup.py Analysis"
      if grep -E '(os\.system|subprocess|eval|exec|__import__|urllib|requests\.get|curl)' setup.py 2>/dev/null | head -5 | grep -q .; then
        log_warn "Suspicious patterns in setup.py:"
        grep -E '(os\.system|subprocess|eval|exec|__import__|urllib|requests\.get|curl)' setup.py | head -5
      else
        log_ok "No suspicious patterns in setup.py"
      fi
    fi
    
    # Bandit - Python security linter
    if check_command bandit; then
      log_subsection "Bandit Security Scan"
      bandit -r . -ll -q 2>&1 || log_warn "Bandit found issues"
    fi
    
    # pip-audit
    if check_command pip-audit && [[ -f "requirements.txt" ]]; then
      log_subsection "pip-audit"
      pip-audit -r requirements.txt 2>&1 || log_warn "Vulnerabilities found"
    fi
  fi
  
  if [[ "$found_project" == "false" ]]; then
    log_warn "No Go, Rust, Node.js, or Python project files found"
    log_info "This might be a different language or a binary-only distribution"
  fi
  
  # Return to original directory
  cd - > /dev/null
  
  # Cleanup or keep
  if [[ "$KEEP_CLONE" == "true" ]]; then
    log_info "Clone kept at: $clone_dir"
  else
    rm -rf "$clone_dir"
    log_info "Cleaned up temporary clone"
  fi
}

scan_npm_package() {
  local package="$1"
  
  log_section "npm Package Analysis: $package"
  
  # Socket.dev analysis (preferred - more comprehensive)
  if check_command socket; then
    log_info "Running Socket.dev security analysis..."
    echo ""
    # Socket package score gives detailed security info
    socket package score npm "$package" --markdown 2>/dev/null || true
  # Fallback to GuardDog if Socket not available
  elif check_command guarddog; then
    log_info "Running GuardDog behavioral analysis..."
    echo ""
    guarddog npm scan "$package" 2>/dev/null || true
  else
    log_warn "Neither Socket CLI nor GuardDog installed - skipping behavioral analysis"
    log_info "Install Socket: npm install -g socket"
  fi
  
  # Check malicious DB
  check_malicious_db "$package" || true
  
  # Try to get GitHub repo for Scorecard
  if check_command npm; then
    local repo_url
    repo_url=$(npm view "$package" repository.url 2>/dev/null | sed 's/git+//;s/\.git$//' | sed 's|^git://|https://|' || echo "")
    
    if [[ -n "$repo_url" && "$repo_url" == *github.com* ]]; then
      # Normalize URL
      repo_url=$(echo "$repo_url" | sed 's|.*github.com/|https://github.com/|' | sed 's|\.git$||')
      run_scorecard "$repo_url"
    else
      log_warn "Could not find GitHub repository for Scorecard analysis"
    fi
  fi
}

scan_pypi_package() {
  local package="$1"
  
  log_section "PyPI Package Analysis: $package"
  
  # Socket.dev analysis (preferred)
  if check_command socket; then
    log_info "Running Socket.dev security analysis..."
    echo ""
    socket package score pypi "$package" --markdown 2>/dev/null || true
  # Fallback to GuardDog
  elif check_command guarddog; then
    log_info "Running GuardDog behavioral analysis..."
    echo ""
    guarddog pypi scan "$package" 2>/dev/null || true
  else
    log_warn "Neither Socket CLI nor GuardDog installed - skipping behavioral analysis"
    log_info "Install Socket: npm install -g socket"
  fi
  
  # Check malicious DB
  check_malicious_db "$package" || true
  
  # Try to get GitHub repo from PyPI metadata
  if check_command curl && check_command jq; then
    local repo_url
    repo_url=$(curl -s "https://pypi.org/pypi/$package/json" 2>/dev/null | \
      jq -r '.info.project_urls // {} | to_entries[] | select(.value | contains("github.com")) | .value' | \
      head -1 || echo "")
    
    if [[ -n "$repo_url" && "$repo_url" == *github.com* ]]; then
      repo_url=$(echo "$repo_url" | sed 's|.*github.com/|https://github.com/|' | cut -d'/' -f1-5)
      run_scorecard "$repo_url"
    else
      log_warn "Could not find GitHub repository for Scorecard analysis"
    fi
  fi
}

scan_github_repo() {
  local repo_url="$1"
  
  # Normalize URL
  repo_url=$(echo "$repo_url" | sed 's|^git@github.com:|https://github.com/|')
  repo_url=$(echo "$repo_url" | sed 's|^github.com|https://github.com|')
  repo_url=$(echo "$repo_url" | sed 's|\.git$||')
  
  local repo_path
  repo_path=$(echo "$repo_url" | sed 's|.*github.com/||')
  
  log_section "GitHub Repository Analysis: $repo_path"
  
  # Check malicious DB
  check_malicious_db "$repo_path" || true
  
  # Scorecard analysis
  run_scorecard "$repo_url"
  
  # Additional repo checks via GitHub API (if gh CLI available)
  if check_command gh; then
    log_section "GitHub Metadata"
    
    local repo_info
    if repo_info=$(gh repo view "$repo_path" --json stargazerCount,forkCount,pushedAt,isArchived,licenseInfo 2>/dev/null); then
      echo "$repo_info" | jq -r '
        "Stars: \(.stargazerCount)",
        "Forks: \(.forkCount)",
        "Last Push: \(.pushedAt)",
        "Archived: \(.isArchived)",
        "License: \(.licenseInfo.name // "None")"
      '
      
      # Check if archived
      if echo "$repo_info" | jq -e '.isArchived == true' &>/dev/null; then
        log_warn "Repository is ARCHIVED - may be unmaintained"
      fi
      
      # Check last activity
      local last_push
      last_push=$(echo "$repo_info" | jq -r '.pushedAt')
      local days_ago
      days_ago=$(( ($(date +%s) - $(date -d "$last_push" +%s 2>/dev/null || echo 0)) / 86400 ))
      
      if [[ $days_ago -gt 365 ]]; then
        log_warn "No commits in over a year ($days_ago days)"
      elif [[ $days_ago -gt 180 ]]; then
        log_warn "No commits in over 6 months ($days_ago days)"
      fi
    fi
  fi
  
  # Deep scan if requested
  if [[ "$DEEP_SCAN" == "true" ]]; then
    deep_scan_repo "$repo_url" "$repo_path"
  else
    log_info "Tip: Use --deep to clone and scan dependencies"
  fi
}

scan_go_module() {
  local module="$1"
  
  log_section "Go Module Analysis: $module"
  
  # Most Go modules are on GitHub
  if [[ "$module" == github.com/* ]]; then
    scan_github_repo "https://$module"
  else
    log_warn "Non-GitHub Go module - limited analysis available"
    check_malicious_db "$module" || true
  fi
}

print_manual_checks() {
  log_section "Manual Verification Checklist"
  cat <<EOF
Before installing, consider checking:
  □ Recent release history (is it actively maintained?)
  □ Open issues/PRs (are security issues addressed?)
  □ Maintainer reputation and history
  □ Install scripts for suspicious commands
  □ Dependencies being pulled in
  □ Download counts / community adoption
EOF
}

main() {
  local target=""
  
  # Parse arguments
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --deep)
        DEEP_SCAN=true
        shift
        ;;
      --keep)
        KEEP_CLONE=true
        shift
        ;;
      -h|--help)
        usage
        ;;
      -*)
        log_error "Unknown option: $1"
        usage
        ;;
      *)
        target="$1"
        shift
        ;;
    esac
  done
  
  [[ -z "$target" ]] && usage
  
  case "$target" in
    npm:*)
      local package="${target#npm:}"
      scan_npm_package "$package"
      ;;
    pypi:*)
      local package="${target#pypi:}"
      scan_pypi_package "$package"
      ;;
    go:*)
      local module="${target#go:}"
      scan_go_module "$module"
      ;;
    *github.com*)
      scan_github_repo "$target"
      ;;
    *)
      log_error "Unknown target format: $target"
      usage
      ;;
  esac
  
  print_manual_checks
  print_mise_reminder

  echo ""
  log_info "Vetting complete. Review results above before proceeding."
}

main "$@"
