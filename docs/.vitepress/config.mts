import { defineConfig } from "vitepress";

export default defineConfig({
  lang: "en-US",
  title: "UniStack",
  description: "A unified, cross-platform package manager powered by embedded Ansible.",

  base: "/",

  head: [
    ['script', {}, `
      if (typeof window !== 'undefined') {
        const lang = navigator.language || navigator.userLanguage;
        if (lang.startsWith('zh') && window.location.pathname === '/') {
          window.location.replace('/zh/');
        }
      }
    `]
  ],

  themeConfig: {
    logo: "/logo.png",
    search: { provider: "local" },
    socialLinks: [
      { icon: "github", link: "https://github.com/snowdreamtech/unistack" },
    ],
  },

  locales: {
    root: {
      label: 'English',
      lang: 'en-US',
      themeConfig: {
        nav: [
          { text: "Reference", link: "/reference/package-format-specification" }
        ],
        sidebar: [
          {
            text: "Reference",
            items: [
              { text: "Package Format", link: "/reference/package-format-specification" },
              { text: "Registry Research", link: "/reference/package-registry-research" },
            ]
          }
        ]
      }
    },
    zh: {
      label: '简体中文',
      lang: 'zh-Hans',
      link: '/zh/',
      themeConfig: {
        nav: [
          { text: "规范参考", link: "/zh/reference/package-format-specification" }
        ],
        sidebar: [
          {
            text: "参考",
            items: [
              { text: "包格式规范", link: "/zh/reference/package-format-specification" },
              { text: "软件源生态研究", link: "/zh/reference/package-registry-research" },
            ]
          }
        ]
      }
    }
  }
});
