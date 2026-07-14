#!/bin/sh
# =====================================================================
# roles/base/files/git_config.sh
#
# Purpose:
#   Automatic configuration of global Git settings.
#   Sets user identity, preferred editor, commit signing, push/pull
#   behaviors, and helpful aliases for developer productivity.
#
# Simple Usage:
#   GIT_USER_NAME="John Doe" GIT_USER_EMAIL="john@example.com" ./git_config.sh
#
# Comprehensive Usage:
#   GIT_USER_NAME="John Doe" \
#   GIT_USER_EMAIL="john@example.com" \
#   GIT_SIGNING=true \
#   GIT_SIGNING_KEY="ABC123" \
#   ./git_config.sh
#
# =====================================================================

# ---------------------------------------------------------------------
# Debugging: Log the (non-sensitive) input environment variables.
# NOTE: GIT_SIGNING_KEY is intentionally omitted to avoid leaking
# sensitive key material in logs.
# ---------------------------------------------------------------------
echo "GIT_USER_NAME: ${GIT_USER_NAME}"
echo "GIT_USER_EMAIL: ${GIT_USER_EMAIL}"
echo "GIT_EDITOR: ${GIT_EDITOR}"
echo "GIT_SIGNING: ${GIT_SIGNING}"

# ---------------------------------------------------------------------
# Configuration: Apply user identity if provided.
# ---------------------------------------------------------------------
if [ -n "${GIT_USER_NAME}" ]; then
  git config --global user.name "${GIT_USER_NAME}"
fi

if [ -n "${GIT_USER_EMAIL}" ]; then
  git config --global user.email "${GIT_USER_EMAIL}"
fi

if [ -n "${GIT_EDITOR}" ]; then
  git config --global core.editor "${GIT_EDITOR}"
fi

# ---------------------------------------------------------------------
# Security: Configure GPG commit signing based on the
# GIT_SIGNING flag. Supports various truthy/falsy strings.
# ---------------------------------------------------------------------
case "${GIT_SIGNING}" in
1 | true | TRUE | True | yes | YES | Yes | on | ON | On)
  if [ -n "${GIT_SIGNING_KEY}" ]; then
    git config --global user.signingkey "${GIT_SIGNING_KEY}"

    # signing with OpenPGP (GPG or GPG2)
    git config --global gpg.program gpg
    git config --global user.signingformat openpgp
    git config --global commit.gpgsign true

    # git signing with ssh (Git 2.34+)
    # git config --global user.signingkey ~/.ssh/id_ed25519.pub
    # git config --global user.signingformat ssh
    # git config --global commit.gpgsign true
  fi
  ;;
0 | false | FALSE | False | no | NO | No | off | OFF | Off | "")
  git config --global commit.gpgsign false
  ;;
*)
  echo "WARN: invalid GIT_SIGNING='${GIT_SIGNING}', expected true/false; disabling signing"
  git config --global commit.gpgsign false
  ;;
esac

# ---------------------------------------------------------------------
# Optimization: Apply standardized Git behaviors for push, pull,
# rebase, and performance features like commitGraph.
# ---------------------------------------------------------------------
git config --global init.defaultBranch main
git config --global push.default simple
git config --global push.autoSetupRemote true
git config --global core.ignorecase false
git config --global pull.rebase true
git config --global rebase.autoStash true
git config --global rebase.updateRefs true
git config --global rerere.enabled true
git config --global color.ui auto
git config --global credential.helper 'cache --timeout=3600'
git config --global help.autocorrect 20
git config --global log.date iso
git config --global core.whitespace trailing-space,space-before-tab
git config --global diff.algorithm histogram
git config --global diff.colorMoved zebra
git config --global core.commitGraph true
git config --global fetch.writeCommitGraph true
git config --global pack.useBitmap true
git config --global fetch.prune true
git config --global fetch.pruneTags true

# ---------------------------------------------------------------------
# Usability: Set up a comprehensive list of shortcuts (aliases) for
# common commands (checkout, branch, commit, log, push, pull).
# ---------------------------------------------------------------------
git config --global alias.co checkout
git config --global alias.cl clone
git config --global alias.cob "checkout -b"
git config --global alias.br branch
git config --global alias.brd "branch -d"
git config --global alias.brD "branch -D"
git config --global alias.ci commit
git config --global alias.st status
git config --global alias.lg 'log --oneline --graph --decorate'
git config --global alias.ls 'log --pretty=format:"%C(yellow)%h %Creset%s" --decorate'
git config --global alias.lsu 'log --pretty=format:"%C(yellow)%h %Creset%s %C(cyan)<%cn>" --decorate'
git config --global alias.ll 'log --pretty=format:"%C(yellow)%h %Creset%s" --decorate --numstat'
git config --global alias.llu 'log --pretty=format:"%C(yellow)%h %Creset%s %C(cyan)<%cn>" --decorate --numstat'
git config --global alias.lnc 'log --pretty=format:"%h %s"'
git config --global alias.lncu 'log --pretty=format:"%h %s <%cn>"'
git config --global alias.lds 'log --pretty=format:"%C(yellow)%h %C(green)%ad %Creset%s" --decorate --date=short'
git config --global alias.ldsu 'log --pretty=format:"%C(yellow)%h %C(green)%ad %Creset%s %C(cyan)<%cn>" --decorate --date=short'
git config --global alias.ld 'log --pretty=format:"%C(yellow)%h %C(green)%ad %Creset%s" --decorate --date=relative'
git config --global alias.ldu 'log --pretty=format:"%C(yellow)%h %C(green)%ad %Creset%s %C(cyan)<%cn>" --decorate --date=relative'
git config --global alias.s "status -s"
git config --global alias.ri "rebase -i"
git config --global alias.mnf "merge --no-ff"
git config --global alias.mg "merge"
git config --global alias.a "add ."
git config --global alias.pf "push --force"
git config --global alias.ps "push"
git config --global alias.po "push origin"
git config --global alias.pom "push origin main:main"
git config --global alias.pomf "push origin main:main --force"
git config --global alias.pod "push origin dev:dev"
git config --global alias.podf "push origin dev:dev --force"
git config --global alias.pos "push origin staging:staging"
git config --global alias.posf "push origin staging:staging --force"
git config --global alias.poa "push origin --all"
git config --global alias.poaf "push origin --all --force"
git config --global alias.pot "push origin --tags"
git config --global alias.potf "push origin --tags --force"
git config --global alias.poat '!git push origin --all && git push origin --tags'
git config --global alias.poatf '!git push origin --all --force && git push origin --tags --force'
git config --global alias.pl "pull"
git config --global alias.plo "pull origin"
git config --global alias.plom "pull origin main"
git config --global alias.plod "pull origin dev"
git config --global alias.plos "pull origin staging"
