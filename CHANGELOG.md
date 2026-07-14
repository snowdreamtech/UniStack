# Changelog

## [0.0.1](https://github.com/snowdreamtech/UniStack/compare/v0.0.1...v0.0.1) (2026-07-14)


### 🚀 Features

* **ansible/base:** add openSUSE/SLE package configuration ([570cdcc](https://github.com/snowdreamtech/UniStack/commit/570cdccc8c7d3a039e03d2e21e1a21506b7f652c))
* **base:** add Gentoo package configuration with fully-qualified names ([a69c915](https://github.com/snowdreamtech/UniStack/commit/a69c9158ce3c46b52ddc6f6d8aaec4cd3092a1c2))
* **base:** add Mageia and NixOS package configurations ([f70ba5b](https://github.com/snowdreamtech/UniStack/commit/f70ba5bcc4b41504d302b5444186c20bcb243bac))
* **base:** add NixOS package configuration ([d1c080d](https://github.com/snowdreamtech/UniStack/commit/d1c080d7a7e74134bb736fde148c3d87f102edfd))
* **base:** add OpenWrt package configuration ([9b92e0c](https://github.com/snowdreamtech/UniStack/commit/9b92e0cae878b1ea6c584e9167c4d776428af0ca))
* **base:** add Void Linux package configuration ([b6203de](https://github.com/snowdreamtech/UniStack/commit/b6203de9fdf5cdec9c22e49525ae68fd382ac6c5))
* port base role from Uniloader with zero external dependencies ([16f383e](https://github.com/snowdreamtech/UniStack/commit/16f383e2e5ec7fe96b15ba982b92b98d6f20f0d9))


### 🐛 Bug Fixes

* **ansible:** bypass urpmi list concatenation bug in variables ([614cbed](https://github.com/snowdreamtech/UniStack/commit/614cbed3f98ef67b354fdea86a0a229ac574c9ef))
* **ansible:** comment out shellcheck from openSUSE base packages ([9eee1f3](https://github.com/snowdreamtech/UniStack/commit/9eee1f38fc2cc37d9bafe8c9700f829d83e543da))
* **base:** add vars/ prefix to include_vars paths to avoid loading tasks files as vars ([4527f47](https://github.com/snowdreamtech/UniStack/commit/4527f470e2efe517e72a6ba298a656035e2ae26b))
* **base:** install python3-dnf on Photon before using package module ([1287d28](https://github.com/snowdreamtech/UniStack/commit/1287d2850b4e841e2a86b2746a7b4462e8d120d5))
* **base:** install python3-dnf to both system and venv for Photon compatibility ([54ad81f](https://github.com/snowdreamtech/UniStack/commit/54ad81f2a6ad9d935b73daf4ed9b5fb4756e4434))
* **base:** reduce Gentoo package list to guaranteed available packages only (remove non-existent packages like sys-apps/psmisc) ([781512b](https://github.com/snowdreamtech/UniStack/commit/781512b34dc99ba64eb9f742c2aa3493556d47f2))
* **base:** reduce openSUSE packages to guaranteed available set (no parallel, libcap, pinentry-curses) ([b3d6548](https://github.com/snowdreamtech/UniStack/commit/b3d6548bc4fa151b78510f88393aa656a5d94001))
* **base:** reduce OpenWrt package list to guaranteed available packages only ([0b01ed2](https://github.com/snowdreamtech/UniStack/commit/0b01ed284f2060011060d7c07f08094e487943f8))
* **base:** reduce Void Linux package list to essential packages only (many packages unavailable in xbps) ([317ebde](https://github.com/snowdreamtech/UniStack/commit/317ebde86d7e20e432eb5ead498728bce91ec3b2))
* **base:** remove pinentry-tty and pinentry-curses on macOS (Homebrew only has pinentry-mac) ([9a639f8](https://github.com/snowdreamtech/UniStack/commit/9a639f87bd7820651208f7a6bd99ea423bfc5181))
* **base:** use gpatch for both Homebrew and MacPorts on macOS ([446f46d](https://github.com/snowdreamtech/UniStack/commit/446f46dd9d9b9d37b6edab9bce7c82e6e7155fa9))
* **base:** use lowercase shellcheck for Void Linux package name ([a3cb6e3](https://github.com/snowdreamtech/UniStack/commit/a3cb6e32a91f251b37614ee20d7b64c565cf08f5))
* **context:** check /etc/os-release existence before reading (handles NixOS minimal containers) ([b81427c](https://github.com/snowdreamtech/UniStack/commit/b81427c02209d8f7af21e58a8ef6f2dd9f9cfe96))
* **init:** install python3-dnf on VMware Photon for Ansible dnf module ([3757665](https://github.com/snowdreamtech/UniStack/commit/3757665a96b2984ba652e2fa6c1358d6b7b00953))
* **init:** install python3-dnf to both system and venv for Photon support ([04be69e](https://github.com/snowdreamtech/UniStack/commit/04be69ecb05f1ad30d4e49feedfbea4df6e6888b))


### 🛠 Refactoring

* **core:** purge support for unsupported edge operating systems ([52998a5](https://github.com/snowdreamtech/UniStack/commit/52998a523e2f7052586a06f4aeba434746943ed3))
* remove urpmi support, add tdnf shell workarounds for base packages, and update OpenWrt default packages ([fbe0249](https://github.com/snowdreamtech/UniStack/commit/fbe0249c09570c115d7e5707c8664a3c9ef5b651))

