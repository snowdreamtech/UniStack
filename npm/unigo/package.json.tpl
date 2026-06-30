{
  "name": "@snowdreamtech/unigo",
  "version": "{{VERSION}}",
  "description": "UniGo - Universal Runtime Manager: enterprise-grade foundational toolchain for multi-AI IDE collaboration",
  "license": "MIT",
  "homepage": "https://github.com/snowdreamtech/unigo",
  "repository": {
    "type": "git",
    "url": "git+https://github.com/snowdreamtech/unigo.git"
  },
  "bugs": {
    "url": "https://github.com/snowdreamtech/unigo/issues"
  },
  "keywords": [
    "unigo",
    "runtime",
    "manager",
    "toolchain",
    "ai",
    "ide"
  ],
  "bin": {
    "unigo": "install.js"
  },
  "scripts": {
    "postinstall": "node install.js"
  },
  "files": [
    "install.js",
    "LICENSE",
    "README.md",
    "README_zh-CN.md"
  ],
  "optionalDependencies": {
    "@snowdreamtech/unigo-darwin-arm64": "{{VERSION}}",
    "@snowdreamtech/unigo-darwin-x64": "{{VERSION}}",
    "@snowdreamtech/unigo-linux-x64": "{{VERSION}}",
    "@snowdreamtech/unigo-linux-arm64": "{{VERSION}}",
    "@snowdreamtech/unigo-linux-ia32": "{{VERSION}}",
    "@snowdreamtech/unigo-linux-arm": "{{VERSION}}",
    "@snowdreamtech/unigo-linux-loong64": "{{VERSION}}",
    "@snowdreamtech/unigo-linux-ppc64le": "{{VERSION}}",
    "@snowdreamtech/unigo-linux-riscv64": "{{VERSION}}",
    "@snowdreamtech/unigo-linux-s390x": "{{VERSION}}",
    "@snowdreamtech/unigo-windows-x64": "{{VERSION}}",
    "@snowdreamtech/unigo-windows-arm64": "{{VERSION}}",
    "@snowdreamtech/unigo-windows-ia32": "{{VERSION}}"
  },
  "engines": {
    "node": ">=18"
  }
}
