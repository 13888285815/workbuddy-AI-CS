import Prism from "prismjs";
import "prismjs/components/prism-bash";
import "prismjs/components/prism-jsx";
import "prismjs/components/prism-typescript";
import "prismjs/components/prism-tsx";
import "prismjs/components/prism-markup";
import "./style.css";

Prism.manual = true;

type SiteTheme = "dark" | "light";

const themeToggle = requiredElement<HTMLButtonElement>("#themeToggle");
let siteTheme: SiteTheme = readStorage("ofv-site-theme") === "dark" ? "dark" : "light";

function requiredElement<T extends HTMLElement>(selector: string): T {
  const element = document.querySelector<T>(selector);
  if (!element) {
    throw new Error(`Missing required docs element: ${selector}`);
  }
  return element;
}

function readStorage(key: string): string | null {
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}

function writeStorage(key: string, value: string): void {
  try {
    localStorage.setItem(key, value);
  } catch {
    // Local preferences are optional.
  }
}

function iconSvg(id: string): string {
  return `<svg aria-hidden="true" focusable="false"><use href="#${id}"></use></svg>`;
}

function applySiteTheme(nextTheme: SiteTheme) {
  siteTheme = nextTheme;
  document.documentElement.dataset.siteTheme = siteTheme;
  themeToggle.innerHTML = iconSvg(siteTheme === "dark" ? "icon-sun" : "icon-moon");
  themeToggle.setAttribute("aria-label", siteTheme === "dark" ? "Switch to light mode" : "Switch to dark mode");
  writeStorage("ofv-site-theme", siteTheme);
}

function syncNavigationState(): void {
  document.documentElement.dataset.navState = window.scrollY > 36 ? "scrolled" : "top";
}

function setHighlightedCode(element: HTMLElement, source: string, languageName: string) {
  const grammar = Prism.languages[languageName] || Prism.languages.markup || Prism.languages.plain;
  element.className = `language-${languageName}`;
  element.parentElement?.classList.add(`language-${languageName}`);
  const highlighted = Prism.highlight(source, grammar, languageName);
  element.innerHTML = highlighted
    .split("\n")
    .map((line: string) => `<span class="code-line">${line || "&nbsp;"}</span>`)
    .join("");
}

function highlightCodeBlocks() {
  for (const code of document.querySelectorAll<HTMLElement>("pre code")) {
    const languageName = code.dataset.language || "typescript";
    setHighlightedCode(code, code.textContent || "", languageName);
  }
}

themeToggle.addEventListener("click", () => {
  applySiteTheme(siteTheme === "dark" ? "light" : "dark");
});

window.addEventListener("scroll", syncNavigationState, { passive: true });

applySiteTheme(siteTheme);
highlightCodeBlocks();
syncNavigationState();
requestAnimationFrame(() => {
  document.documentElement.dataset.siteReady = "true";
  document.documentElement.dataset.siteBoot = "ready";
});
