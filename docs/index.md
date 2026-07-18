---
layout: home

hero:
  name: "UniStack"
  text: "The Universal Package Engine"
  tagline: "A unified, cross-platform package manager powered by embedded Ansible."
  actions:
    - theme: brand
      text: Read Specifications
      link: /reference/package-format-specification
    - theme: alt
      text: View on GitHub
      link: https://github.com/snowdreamtech/unistack

features:
  - icon: 📦
    title: Universal Format
    details: Create one package that seamlessly installs across Linux, macOS, and Windows.
  - icon: ⚡
    title: Embedded Execution
    details: Ansible runs embedded natively within Go, zero external dependencies required.
  - icon: 🛡️
    title: Uncompromising Security
    details: Pure stateless extraction with strict SHA-256 verification and atomic swap installations.
---
<script>
if (typeof window !== 'undefined') {
  const lang = navigator.language || navigator.userLanguage;
  if (lang.startsWith('zh') && window.location.pathname === '/') {
    window.location.replace('/zh/');
  }
}
</script>
