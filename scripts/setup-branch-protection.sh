#!/usr/bin/env bash
#
# Set up branch protection on main for the Opero repo, via the GitHub REST API
# using the gh CLI. Run this yourself — it needs GitHub API credentials that the
# assistant does not have.
#
# Usage:
#   bash scripts/setup-branch-protection.sh
#   OWNER=davidnguyen2205 REPO=opero BRANCH=main CHECK=backend bash scripts/setup-branch-protection.sh
#
# Prereqs:
#   - gh CLI installed:            brew install gh
#   - gh authenticated AS THE REPO OWNER (gh uses its OWN token, NOT your SSH
#     `github-per` alias):         gh auth login   # choose account davidnguyen2205
#
# CAVEATS (read these):
#   - Classic branch protection on PRIVATE repos has historically required a
#     paid plan (GitHub Pro). If you get a 403 about your plan, use Rulesets
#     instead (Settings -> Rules -> Rulesets) — this script can't do those.
#   - The exact API payload below is best-effort (assistant training knowledge,
#     not tested in your environment). The script reads the protection back and
#     prints it so you can confirm it applied as intended.
set -euo pipefail

OWNER="${OWNER:-davidnguyen2205}"
REPO="${REPO:-opero}"
BRANCH="${BRANCH:-main}"
CHECK="${CHECK:-backend}"   # CI status-check context (the ci.yml job name)

die() { echo "ERROR: $*" >&2; exit 1; }

# --- preflight ---
command -v gh >/dev/null 2>&1 || die "gh CLI not found. Install with: brew install gh"
gh auth status >/dev/null 2>&1 || die "gh is not authenticated. Run: gh auth login  (as $OWNER)"

echo "Target: $OWNER/$REPO branch '$BRANCH', required check '$CHECK'"

# Confirm the repo is reachable and report visibility (so you notice if it's public).
visibility="$(gh api "repos/$OWNER/$REPO" --jq '.visibility' 2>/dev/null)" \
  || die "Cannot access repos/$OWNER/$REPO. Is gh authed as the owner, and does the repo exist?"
echo "Repo visibility: $visibility"
[ "$visibility" = "public" ] && echo "WARNING: this repo is PUBLIC. Proprietary code is world-readable." >&2

# --- apply protection ---
# required_pull_request_reviews present (with 0 approvals) => require a PR before
# merging, but don't require an approval (solo-repo friendly; you can't approve
# your own PR). enforce_admins=false so the owner can still bypass in a pinch.
read -r -d '' BODY <<JSON || true
{
  "required_status_checks": { "strict": true, "contexts": ["$CHECK"] },
  "enforce_admins": false,
  "required_pull_request_reviews": { "required_approving_review_count": 0 },
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false
}
JSON

echo "Applying protection..."
if ! printf '%s' "$BODY" | gh api -X PUT \
      "repos/$OWNER/$REPO/branches/$BRANCH/protection" \
      -H "Accept: application/vnd.github+json" \
      --input - >/dev/null 2>/tmp/opero-bp-err; then
  echo "--- gh error ---" >&2; cat /tmp/opero-bp-err >&2
  if grep -qi "upgrade\|plan\|not available" /tmp/opero-bp-err; then
    die "Branch protection appears unavailable on this plan (likely private+free). Use Rulesets in the web UI instead."
  fi
  die "Failed to apply branch protection (see error above)."
fi

# --- verify (read back the key fields) ---
echo "Applied. Current protection on '$BRANCH':"
gh api "repos/$OWNER/$REPO/branches/$BRANCH/protection" --jq '{
  required_pull_request: (.required_pull_request_reviews != null),
  required_approvals: (.required_pull_request_reviews.required_approving_review_count // 0),
  required_checks: .required_status_checks.contexts,
  strict_up_to_date: .required_status_checks.strict,
  force_pushes_allowed: .allow_force_pushes.enabled,
  deletions_allowed: .allow_deletions.enabled,
  admins_enforced: .enforce_admins.enabled
}'

echo
echo "Done. Note: the '$CHECK' check only blocks merges once CI has run at least"
echo "once and reported that context. Verify in Settings -> Branches."
