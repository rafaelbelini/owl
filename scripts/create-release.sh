#!/usr/bin/env bash
set -euo pipefail

# Script to create a release branch and PR
# Compares develop with main, creates release/<VERSION> branch, and opens a PR to main

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if we're in a git repository
cd "$PROJECT_ROOT"
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    log_error "Not a git repository: $PROJECT_ROOT"
    exit 1
fi

# Ensure develop and main branches exist
for branch in develop main; do
    if ! git rev-parse --verify "$branch" > /dev/null 2>&1; then
        log_error "Branch '$branch' does not exist"
        exit 1
    fi
done

# Fetch latest from remote
log_info "Fetching latest changes from remote..."
git fetch origin develop main 2>/dev/null || true

# Check for unpushed changes on develop
if ! git diff --quiet HEAD origin/develop 2>/dev/null; then
    log_warn "You have unpushed changes on develop. It's recommended to push before creating a release."
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Aborted."
        exit 0
    fi
fi

# Get list of commits in develop not in main
COMMITS_AHEAD=$(git log --oneline origin/main..origin/develop 2>/dev/null || git log --oneline main..develop)
if [ -z "$COMMITS_AHEAD" ]; then
    log_info "develop is already up to date with main. Nothing to release."
    exit 0
fi

COMMIT_COUNT=$(echo "$COMMITS_AHEAD" | wc -l)
log_info "Found $COMMIT_COUNT commit(s) in develop not in main:"
echo "$COMMITS_AHEAD" | head -10
if [ "$COMMIT_COUNT" -gt 10 ]; then
    echo "  ... and $((COMMIT_COUNT - 10)) more"
fi
echo

# Calculate next version
# Get the latest tag on main (if any)
LATEST_TAG=$(git tag --list --merged origin/main 2>/dev/null | grep -E '^v?[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -1)

if [ -z "$LATEST_TAG" ]; then
    # No tags, start from 0.1.0 or 1.0.0
    MAJOR=0
    MINOR=1
    PATCH=0
    log_info "No previous release found. Starting with version 0.1.0"
else
    # Parse version from tag (remove 'v' prefix if present)
    VERSION_STR="${LATEST_TAG#v}"
    MAJOR=$(echo "$VERSION_STR" | cut -d. -f1)
    MINOR=$(echo "$VERSION_STR" | cut -d. -f2)
    PATCH=$(echo "$VERSION_STR" | cut -d. -f3)

    # Increment minor version for release
    MINOR=$((MINOR + 1))
    PATCH=0
    log_info "Latest tag: $LATEST_TAG. Next version: $MAJOR.$MINOR.$PATCH"
fi

VERSION="$MAJOR.$MINOR.$PATCH"
RELEASE_BRANCH="release/v$VERSION"

log_info "Will create release branch: $RELEASE_BRANCH"

# Ask for confirmation
read -p "Create release branch and PR? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_info "Aborted."
    exit 0
fi

# Create release branch from develop
log_info "Creating release branch '$RELEASE_BRANCH' from develop..."
git checkout -b "$RELEASE_BRANCH" develop

# Push release branch to remote
log_info "Pushing release branch to origin..."
git push -u origin "$RELEASE_BRANCH"

# Create PR using gh CLI
if command -v gh &> /dev/null; then
    log_info "Creating pull request to main..."

    # Generate PR title and body
    PR_TITLE="Release v$VERSION"

    # Build PR body with commit list
    PR_BODY="## Release v$VERSION

This PR includes the following changes from develop:

"
    while IFS= read -r commit; do
        PR_BODY+="- $commit\n"
    done <<< "$COMMITS_AHEAD"

    PR_BODY+="
🤖 Generated with [Claude Code](https://claude.com/claude-code)"

    gh pr create \
        --base main \
        --head "$RELEASE_BRANCH" \
        --title "$PR_TITLE" \
        --body "$PR_BODY" \
        --label "release"

    log_info "Pull request created successfully!"
else
    log_warn "GitHub CLI (gh) is not installed. Please create the PR manually."
    log_info "Run the following command to create the PR:"
    echo
    echo "  gh pr create --base main --head $RELEASE_BRANCH --title 'Release v$VERSION'"
    echo
fi

# Return to develop branch
git checkout develop

log_info "Done! Release branch '$RELEASE_BRANCH' has been created and pushed."
