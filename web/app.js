let currentLang = localStorage.getItem("benzcloud_web_lang") || "de";
let baseDomain = "intern";

const i18n = {
  de: {
    heading_sites: "Gehostete Webseiten",
    desc_sites: "Erstelle und verwalte Webseiten unter frei wählbaren Subdomains mit HTML, PHP oder Astro.",
    btn_new_site: "➕ Neue Webseite",
    modal_title: "Neue Webseite erstellen",
    lbl_subdomain: "Wunsch-Subdomain:",
    lbl_title: "Titel der Webseite:",
    lbl_engine: "Web-Engine / Technologie:",
    btn_cancel: "Abbrechen",
    btn_create: "Webseite anlegen",
    btn_open: "🌐 Öffnen",
    btn_delete: "🗑️ Löschen",
    confirm_delete: "Möchtest du diese Webseite wirklich löschen?"
  },
  en: {
    heading_sites: "Hosted Websites",
    desc_sites: "Create and manage websites under arbitrary subdomains using HTML, PHP, or Astro.",
    btn_new_site: "➕ New Website",
    modal_title: "Create New Website",
    lbl_subdomain: "Custom Subdomain:",
    lbl_title: "Website Title:",
    lbl_engine: "Web Engine / Technology:",
    btn_cancel: "Cancel",
    btn_create: "Create Website",
    btn_open: "🌐 Open",
    btn_delete: "🗑️ Delete",
    confirm_delete: "Are you sure you want to delete this website?"
  }
};

document.addEventListener("DOMContentLoaded", () => {
  applyLanguage(currentLang);
  fetchSites();
});

function setLanguage(lang) {
  currentLang = lang;
  localStorage.setItem("benzcloud_web_lang", lang);
  applyLanguage(lang);
}

function applyLanguage(lang) {
  document.querySelectorAll("[data-i18n]").forEach(el => {
    const key = el.getAttribute("data-i18n");
    if (i18n[lang] && i18n[lang][key]) {
      el.textContent = i18n[lang][key];
    }
  });
  document.getElementById("langDE").classList.toggle("active", lang === "de");
  document.getElementById("langEN").classList.toggle("active", lang === "en");
}

async function fetchSites() {
  try {
    const res = await fetch("/api/sites");
    const data = await res.json();
    baseDomain = data.base_domain || "intern";
    document.getElementById("dispDomainSuffix").textContent = `.${baseDomain}`;

    const grid = document.getElementById("sitesGrid");
    grid.innerHTML = "";

    data.sites.forEach(s => {
      const card = document.createElement("div");
      card.className = "site-card";
      const siteUrl = `http://${s.subdomain}.${baseDomain}`;
      const engineClass = s.engine === "php" ? "badge-php" : "badge-engine";

      card.innerHTML = `
        <div class="s-top">
          <div>
            <div class="s-title">${escapeHtml(s.title)}</div>
            <div class="s-url">${s.subdomain}.${baseDomain}</div>
          </div>
          <span class="badge-engine ${engineClass}">${s.engine.toUpperCase()}</span>
        </div>
        <p class="desc" style="font-size:0.8rem;">Erstellt am: ${new Date(s.created_at).toLocaleDateString()}</p>
        <div class="s-actions">
          <a class="btn-primary btn-sm" href="${siteUrl}" target="_blank">${i18n[currentLang].btn_open}</a>
          <button class="btn-outline btn-sm" onclick="deleteSite('${escapeHtml(s.subdomain)}')">${i18n[currentLang].btn_delete}</button>
        </div>
      `;
      grid.appendChild(card);
    });
  } catch (err) {
    console.error("Failed to load sites:", err);
  }
}

function openCreateModal() {
  document.getElementById("createSiteModal").style.display = "flex";
}

function closeCreateModal() {
  document.getElementById("createSiteModal").style.display = "none";
}

async function submitNewSite(e) {
  e.preventDefault();
  const subdomain = document.getElementById("newSubdomain").value.trim();
  const title = document.getElementById("newTitle").value.trim();
  const engine = document.getElementById("newEngine").value;

  try {
    const res = await fetch("/api/sites", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ subdomain, title, engine })
    });
    if (!res.ok) {
      const err = await res.json();
      alert("Fehler: " + (err.error || "Erstellung fehlgeschlagen"));
      return;
    }
    closeCreateModal();
    fetchSites();
  } catch (err) {
    alert("Netzwerkfehler: " + err);
  }
}

async function deleteSite(sub) {
  if (!confirm(i18n[currentLang].confirm_delete)) return;
  try {
    await fetch(`/api/sites?subdomain=${encodeURIComponent(sub)}`, { method: "DELETE" });
    fetchSites();
  } catch (err) {
    alert("Löschen fehlgeschlagen");
  }
}

function escapeHtml(str) {
  if (!str) return "";
  return String(str).replace(/[&<>"']/g, m => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#39;"
  }[m]));
}
