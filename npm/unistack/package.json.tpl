{
  "name": "@snowdreamtech/unistack",
  "version": "{{VERSION}}",
  "description": "UniStack - Universal Runtime Manager: enterprise-grade foundational toolchain for multi-AI IDE collaboration",
  "license": "MIT",
  "homepage": "https://github.com/snowdreamtech/unistack",
  "repository": {
    "type": "git",
    "url": "git+https://github.com/snowdreamtech/unistack.git"
  },
  "bugs": {
    "url": "https://github.com/snowdreamtech/unistack/issues"
  },
  "keywords": [
    "unistack",
    "runtime",
    "manager",
    "toolchain",
    "ai",
    "ide"
  ],
  "bin": {
    "unistack": "install.js"
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
    "@snowdreamtech/unistack-darwin-arm64": "{{VERSION}}",
    "@snowdreamtech/unistack-darwin-x64": "{{VERSION}}",
    "@snowdreamtech/unistack-linux-x64": "{{VERSION}}",
    "@snowdreamtech/unistack-linux-arm64": "{{VERSION}}",
    "@snowdreamtech/unistack-linux-ia32": "{{VERSION}}",
    "@snowdreamtech/unistack-linux-arm": "{{VERSION}}",
    "@snowdreamtech/unistack-linux-loong64": "{{VERSION}}",
    "@snowdreamtech/unistack-linux-ppc64le": "{{VERSION}}",
    "@snowdreamtech/unistack-linux-riscv64": "{{VERSION}}",
    "@snowdreamtech/unistack-linux-s390x": "{{VERSION}}",
    "@snowdreamtech/unistack-windows-x64": "{{VERSION}}",
    "@snowdreamtech/unistack-windows-arm64": "{{VERSION}}",
    "@snowdreamtech/unistack-windows-ia32": "{{VERSION}}"
  },
  "engines": {
    "node": ">=18"
  }
}
