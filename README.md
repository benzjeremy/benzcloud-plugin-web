# 🌐 BenzCloud Web-Hosting Plugin

> **Notice:** BenzCloud Web-Hosting Plugin is in **active development / pre-release status** (`v1.0`).

BenzCloud Web Plugin (`benzcloud-plugin-web`) is the multi-subdomain virtual website hosting engine for the BenzCloud enterprise suite. It allows users to create and serve multiple independent websites across arbitrary custom subdomains (e.g. `blog.<domain>`, `shop.<domain>`, `portfolio.<domain>`) with support for HTML5 (pre-installed by default), PHP scripting, and static Astro web applications.

## ✨ Features

- **Arbitrary Custom Subdomains:** Users select custom subdomains without restrictions to system subdomains.
- **Multi-Engine Support:**
  - **HTML5 & CSS3:** Ultra-fast static file serving out-of-the-box.
  - **PHP Execution:** Integrated dynamic runner with CGI execution.
  - **Astro & Modern Web:** Static bundle hosting for JAMstack web apps.
- **Zero-Config DNS Hook:** Automatically announces registered subdomains to the BenzCloud custom DNS server and reverse proxy.
- **Built-in Management Dashboard:** Responsive, 100% bilingual (DE/EN) web UI for site deployment and file management.

---

## 📦 Installation & Usage

```bash
# Launch plugin standalone
./benzcloud-plugin-web -port 8091 -domain benzjeremy.de

# Or install via Go
go install github.com/benzjeremy/benzcloud-plugin-web@latest
```

---

## 👥 Authors & Credits
- **Jeremy Benz** ([@benzjeremy](https://github.com/benzjeremy)) – Lead Engineer & Project Creator
- Pair-programmed with AI Assistant (Google Antigravity)
- © 2026 Jeremy Benz

## 📄 License
Released under the [GNU General Public License v3.0 (GPL-3.0)](LICENSE).
