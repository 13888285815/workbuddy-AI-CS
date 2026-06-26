import {
  archivePlugin,
  assetPlugin,
  audioPlugin,
  cadPlugin,
  createViewer,
  drawingPlugin,
  emailPlugin,
  epubPlugin,
  gisPlugin,
  imagePlugin,
  model3dPlugin,
  officePlugin,
  ofdPlugin,
  pdfPlugin,
  textPlugin,
  videoPlugin,
  xpsPlugin
} from "@open-file-viewer/core";
import "@open-file-viewer/core/style.css";
import pdfWorkerSrc from "pdfjs-dist/build/pdf.worker.mjs?url";
import "./style.css";

interface FileNode {
  name: string;
  type: "file" | "folder";
  children?: FileNode[];
  file?: File;
  expanded?: boolean;
}

interface DemoFile {
  label: string;
  file: File;
  category: string;
}

const plugins = [
  imagePlugin(),
  pdfPlugin({ workerSrc: pdfWorkerSrc }),
  officePlugin(),
  textPlugin(),
  archivePlugin(),
  audioPlugin(),
  videoPlugin(),
  epubPlugin(),
  xpsPlugin(),
  ofdPlugin(),
  drawingPlugin(),
  gisPlugin(),
  model3dPlugin(),
  cadPlugin(),
  assetPlugin(),
  emailPlugin()
];

let currentViewer: any = null;
let currentFile: File | null = null;
let fileTree: FileNode[] = [];
let filteredTree: FileNode[] = [];
let searchQuery = "";
let selectedNode: FileNode | null = null;

const demoFiles: DemoFile[] = [
  { label: "欢迎 Markdown", file: createMarkdownFile(), category: "文档" },
  { label: "API JSON", file: createJsonFile(), category: "文档" },
  { label: "CSV 表格", file: createCsvFile(), category: "文档" },
  { label: "HTML 页面", file: createHtmlFile(), category: "文档" },
  { label: "Word 文档", file: createDocxSample(), category: "办公文档" },
  { label: "Excel 表格", file: createXlsxSample(), category: "办公文档" },
  { label: "PPT 演示", file: createPptxSample(), category: "办公文档" },
  { label: "PDF 文档", file: createPdfSample(), category: "文档" },
  { label: "矢量 SVG", file: createSvgFile(), category: "图片" },
  { label: "GeoJSON 地图", file: createGeoJsonFile(), category: "地理" },
  { label: "Excalidraw 白板", file: createExcalidrawFile(), category: "绘图" },
  { label: "DXF 图纸", file: createDxfFile(), category: "工程" },
  { label: "音频 WAV", file: createWavSample(), category: "媒体" },
  { label: "邮件 EML", file: createEmailSample(), category: "文档" },
  { label: "压缩包 ZIP", file: createArchiveSample(), category: "归档" },
  { label: "EPUB 电子书", file: createEpubSample(), category: "文档" },
  { label: "OFD 发票", file: createOfdSample(), category: "文档" },
  { label: "STEP 模型", file: createStepSample(), category: "三维模型" },
  { label: "OBJ 模型", file: createObjSample(), category: "三维模型" },
  { label: "DWG 图纸", file: createDwgSample(), category: "工程" }
];

function buildFileTree(): FileNode[] {
  const categories: Record<string, FileNode> = {};
  
  demoFiles.forEach(demo => {
    if (!categories[demo.category]) {
      categories[demo.category] = {
        name: demo.category,
        type: "folder",
        children: [],
        expanded: true
      };
    }
    categories[demo.category].children!.push({
      name: demo.label,
      type: "file",
      file: demo.file
    });
  });
  
  return Object.values(categories);
}

function filterTree(nodes: FileNode[], query: string): FileNode[] {
  if (!query) return nodes;
  const lowerQuery = query.toLowerCase();
  
  return nodes.reduce<FileNode[]>((acc, node) => {
    if (node.type === "folder") {
      const filteredChildren = filterTree(node.children || [], query);
      if (filteredChildren.length > 0 || node.name.toLowerCase().includes(lowerQuery)) {
        acc.push({ ...node, children: filteredChildren, expanded: true });
      }
    } else if (node.name.toLowerCase().includes(lowerQuery)) {
      acc.push(node);
    }
    return acc;
  }, []);
}

function createFileTreeHTML(nodes: FileNode[], level = 0): string {
  return nodes.map(node => {
    if (node.type === "folder") {
      return `
        <div class="tree-folder" data-expanded="${node.expanded}">
          <div class="tree-item folder" style="padding-left: ${level * 16 + 12}px">
            <span class="folder-icon">${node.expanded ? "📂" : "📁"}</span>
            <span class="item-name">${node.name}</span>
          </div>
          ${node.expanded ? `<div class="tree-children">${createFileTreeHTML(node.children || [], level + 1)}</div>` : ""}
        </div>
      `;
    } else {
      const icon = getFileIcon(node.name);
      return `
        <div class="tree-item file" style="padding-left: ${level * 16 + 12}px" data-file="${node.name}">
          <span class="file-icon">${icon}</span>
          <span class="item-name">${node.name}</span>
        </div>
      `;
    }
  }).join("");
}

function getFileIcon(filename: string): string {
  const ext = filename.split(".").pop()?.toLowerCase() || "";
  const iconMap: Record<string, string> = {
    md: "📄", json: "📋", csv: "📊", html: "🌐", docx: "📝", xlsx: "📈", pptx: "📽️",
    pdf: "📕", svg: "🎨", geojson: "🗺️", excalidraw: "🎨", dxf: "📐", wav: "🎵",
    eml: "📧", zip: "📦", epub: "📚", ofd: "📄", step: "🔷", obj: "🔷", dwg: "📐"
  };
  return iconMap[ext] || "📄";
}

function renderFileTree() {
  const treeContainer = document.getElementById("file-tree");
  if (treeContainer) {
    filteredTree = filterTree(fileTree, searchQuery);
    treeContainer.innerHTML = createFileTreeHTML(filteredTree);
    attachTreeEvents();
  }
}

function attachTreeEvents() {
  document.querySelectorAll(".tree-item.folder").forEach(folderEl => {
    folderEl.addEventListener("click", (e) => {
      const folderDiv = (e.currentTarget as HTMLElement).closest(".tree-folder");
      if (folderDiv) {
        const isExpanded = folderDiv.getAttribute("data-expanded") === "true";
        folderDiv.setAttribute("data-expanded", (!isExpanded).toString());
        const icon = folderDiv.querySelector(".folder-icon");
        if (icon) icon.textContent = isExpanded ? "📁" : "📂";
        const children = folderDiv.querySelector(".tree-children");
        if (children) {
          children.style.display = isExpanded ? "none" : "block";
        }
      }
    });
  });

  document.querySelectorAll(".tree-item.file").forEach(fileEl => {
    fileEl.addEventListener("click", (e) => {
      const filename = (e.currentTarget as HTMLElement).getAttribute("data-file");
      if (filename) {
        document.querySelectorAll(".tree-item.file").forEach(el => el.classList.remove("selected"));
        (e.currentTarget as HTMLElement).classList.add("selected");
        openFile(filename);
      }
    });
  });
}

function openFile(filename: string) {
  const demo = demoFiles.find(d => d.label === filename);
  if (!demo) return;
  
  currentFile = demo.file;
  selectedNode = { name: filename, type: "file", file: demo.file };
  
  const previewContainer = document.getElementById("preview-container");
  if (previewContainer) {
    previewContainer.innerHTML = '<div id="file-viewer"></div>';
    
    if (currentViewer) {
      currentViewer.destroy();
    }
    
    currentViewer = createViewer({
      container: "#file-viewer",
      file: currentFile,
      fileName: currentFile.name,
      plugins,
      toolbar: true,
      theme: "light",
      fit: "contain"
    });
  }
  
  updateFileInfo(demo);
}

function updateFileInfo(demo: DemoFile) {
  const infoPanel = document.getElementById("file-info");
  if (infoPanel) {
    const typeDisplay = demo.file.type || "未知类型";
    infoPanel.innerHTML = `
      <div class="info-item"><span class="info-label">文件名:</span> ${demo.label}</div>
      <div class="info-item"><span class="info-label">类型:</span> ${typeDisplay}</div>
      <div class="info-item"><span class="info-label">大小:</span> ${formatFileSize(demo.file.size)}</div>
    `;
  }
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + " 字节";
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
  return (bytes / (1024 * 1024)).toFixed(1) + " MB";
}

function initSearch() {
  const searchInput = document.getElementById("search-input") as HTMLInputElement;
  if (searchInput) {
    searchInput.addEventListener("input", (e) => {
      searchQuery = (e.target as HTMLInputElement).value;
      renderFileTree();
    });
  }
}

function initThemeToggle() {
  const themeBtn = document.getElementById("theme-toggle");
  if (themeBtn) {
    themeBtn.addEventListener("click", () => {
      const html = document.documentElement;
      const currentTheme = html.getAttribute("data-theme");
      const newTheme = currentTheme === "dark" ? "light" : "dark";
      html.setAttribute("data-theme", newTheme);
      themeBtn.textContent = newTheme === "dark" ? "☀️" : "🌙";
      
      if (currentViewer) {
        currentViewer.updateOptions({ theme: newTheme });
      }
    });
  }
}

function initUpload() {
  const uploadBtn = document.getElementById("upload-btn");
  const fileInput = document.getElementById("file-input") as HTMLInputElement;
  
  if (uploadBtn && fileInput) {
    uploadBtn.addEventListener("click", () => {
      fileInput.click();
    });
    
    fileInput.addEventListener("change", (e) => {
      const files = (e.target as HTMLInputElement).files;
      if (files && files.length > 0) {
        const file = files[0];
        currentFile = file;
        
        const previewContainer = document.getElementById("preview-container");
        if (previewContainer) {
          previewContainer.innerHTML = '<div id="file-viewer"></div>';
          
          if (currentViewer) {
            currentViewer.destroy();
          }
          
          currentViewer = createViewer({
            container: "#file-viewer",
            file: file,
            fileName: file.name,
            plugins,
            toolbar: true,
            theme: "light",
            fit: "contain"
          });
        }
        
        const infoPanel = document.getElementById("file-info");
        if (infoPanel) {
          const typeDisplay = file.type || "未知类型";
          infoPanel.innerHTML = `
            <div class="info-item"><span class="info-label">文件名:</span> ${file.name}</div>
            <div class="info-item"><span class="info-label">类型:</span> ${typeDisplay}</div>
            <div class="info-item"><span class="info-label">大小:</span> ${formatFileSize(file.size)}</div>
          `;
        }
      }
    });
  }
}

function init() {
  fileTree = buildFileTree();
  renderFileTree();
  initSearch();
  initThemeToggle();
  initUpload();
  
  const sidebarToggle = document.getElementById("sidebar-toggle");
  const sidebar = document.getElementById("sidebar");
  if (sidebarToggle && sidebar) {
    sidebarToggle.addEventListener("click", () => {
      sidebar.classList.toggle("collapsed");
      sidebarToggle.textContent = sidebar.classList.contains("collapsed") ? "▶" : "◀";
    });
  }
}

document.addEventListener("DOMContentLoaded", init);

function createMarkdownFile(): File {
  return new File(
    [`# 文件预览管理器

在任意产品界面嵌入文件预览功能。

- 支持 Vanilla JavaScript、React、Vue 和 Svelte
- 多格式插件架构
- 响应式容器预览

\`\`\`ts
createViewer({ container: "#viewer", file, plugins });
\`\`\`
`],
    "welcome.md",
    { type: "text/markdown" }
  );
}

function createJsonFile(): File {
  return new File(
    [JSON.stringify({ 包名: "@open-file-viewer/core", 接口: "createViewer", 支持框架: ["原生JS", "React", "Vue", "Svelte"] }, null, 2)],
    "api.json",
    { type: "application/json" }
  );
}

function createCsvFile(): File {
  return new File(
    [`名称,格式,状态
合同,pdf,稳定
报告,docx,增强
地图,geojson,预览
图纸,dxf,预览
版图,gds,预览
压缩包,zip,预览
`],
    "formats.csv",
    { type: "text/csv" }
  );
}

function createHtmlFile(): File {
  return new File(
    [`<!doctype html>
<html>
  <head><title>文件预览管理器</title></head>
  <body>
    <h1>嵌入式预览</h1>
    <p>HTML、Markdown、代码和数据文件在同一预览容器内渲染。</p>
  </body>
</html>
`],
    "preview.html",
    { type: "text/html" }
  );
}

function createSvgFile(): File {
  return new File(
    [`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 900 520">
  <rect width="900" height="520" rx="34" fill="#0b1220"/>
  <path d="M170 360 C260 140 430 430 540 180 S730 240 760 120" fill="none" stroke="#b8ff4d" stroke-width="28" stroke-linecap="round"/>
  <circle cx="230" cy="180" r="54" fill="#70e4ff" opacity=".9"/>
  <rect x="520" y="300" width="190" height="90" rx="22" fill="#d6ff87" opacity=".88"/>
  <text x="80" y="90" fill="#fff" font-size="44" font-family="Arial">文件预览管理器</text>
</svg>`],
    "brand-preview.svg",
    { type: "image/svg+xml" }
  );
}

function createGeoJsonFile(): File {
  return new File(
    [JSON.stringify({
      type: "FeatureCollection",
      features: [
        { type: "Feature", properties: { name: "总部位置" }, geometry: { type: "Point", coordinates: [116.397, 39.908] } },
        { type: "Feature", properties: { name: "预览路线" }, geometry: { type: "LineString", coordinates: [[116.36, 39.9], [116.397, 39.908], [116.43, 39.92]] } }
      ]
    }, null, 2)],
    "map.geojson",
    { type: "application/geo+json" }
  );
}

function createExcalidrawFile(): File {
  return new File(
    [JSON.stringify({ type: "excalidraw", version: 2, elements: [{ id: "title", type: "text", x: 120, y: 90, text: "文件预览管理器" }] })],
    "board.excalidraw",
    { type: "application/vnd.excalidraw+json" }
  );
}

function createDxfFile(): File {
  return new File(
    [`0
SECTION
2
ENTITIES
0
LINE
8
A-WALL
10
0
20
0
11
240
21
0
0
ENDSEC
0
EOF`],
    "drawing.dxf",
    { type: "application/dxf" }
  );
}

function createPdfSample(): File {
  return new File(
    [`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>
endobj
4 0 obj
<< /Length 95 >>
stream
BT
/F1 24 Tf
72 720 Td
(Open File Viewer PDF Sample) Tj
ET
endstream
endobj
5 0 obj
<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
endobj
trailer
<< /Root 1 0 R >>
%%EOF`],
    "sample.pdf",
    { type: "application/pdf" }
  );
}

function createWavSample(): File {
  const sampleRate = 44100;
  const channels = 1;
  const bitDepth = 16;
  const sampleCount = sampleRate;
  const dataSize = sampleCount * channels * (bitDepth / 8);
  const bytes = new Uint8Array(44 + dataSize);
  const view = new DataView(bytes.buffer);
  bytes.set([82, 73, 70, 70], 0);
  view.setUint32(4, 36 + dataSize, true);
  bytes.set([87, 65, 86, 69, 102, 109, 116, 32], 8);
  view.setUint32(16, 16, true);
  view.setUint16(20, 1, true);
  view.setUint16(22, channels, true);
  view.setUint32(24, sampleRate, true);
  view.setUint32(28, sampleRate * channels * (bitDepth / 8), true);
  view.setUint16(32, channels * (bitDepth / 8), true);
  view.setUint16(34, bitDepth, true);
  bytes.set([100, 97, 116, 97], 36);
  view.setUint32(40, dataSize, true);
  for (let i = 0; i < sampleCount; i++) {
    const sample = Math.round(Math.sin((i / sampleRate) * 440 * Math.PI * 2) * 16000);
    view.setInt16(44 + i * 2, sample, true);
  }
  return new File([bytes.buffer], "sample-audio.wav", { type: "audio/wav" });
}

function createEmailSample(): File {
  return new File(
    [[
      "From: 产品团队@example.com",
      "To: 用户@example.com",
      "Subject: 文件预览管理器示例邮件",
      "MIME-Version: 1.0",
      "Content-Type: text/plain; charset=utf-8",
      "",
      "欢迎使用文件预览管理器。",
      "此示例用于测试邮件头、正文渲染和工具栏稳定性。"
    ].join("\r\n")],
    "sample-email.eml",
    { type: "message/rfc822" }
  );
}

function createArchiveSample(): File {
  return createZipFile("sample-archive.zip", "application/zip", [
    { path: "readme.md", content: "# 压缩包示例\n\n文件预览管理器可以查看压缩包内容。" },
    { path: "data/config.json", content: JSON.stringify({ viewer: "文件预览管理器", sample: true }) }
  ]);
}

function createEpubSample(): File {
  return createZipFile("sample-book.epub", "application/epub+zip", [
    { path: "mimetype", content: "application/epub+zip" },
    { path: "META-INF/container.xml", content: `<?xml version="1.0"?><container version="1.0"><rootfiles><rootfile full-path="OPS/package.opf"/></rootfiles></container>` },
    { path: "OPS/package.opf", content: `<?xml version="1.0"?><package><metadata><dc:title>文件预览管理器 EPUB</dc:title></metadata></package>` },
    { path: "OPS/chapter.xhtml", content: `<html><body><h1>文件预览管理器 EPUB</h1><p>EPUB示例章节在预览区渲染。</p></body></html>` }
  ]);
}

function createOfdSample(): File {
  return createZipFile("sample-document.ofd", "application/ofd", [
    { path: "Doc_0/Pages/Page_0/Content.xml", content: `<ofd:Page xmlns:ofd="http://www.ofdspec.org/2016"><ofd:Content><ofd:TextObject Boundary="20 30 220 18" Size="12"><ofd:TextCode>文件预览管理器 OFD示例</ofd:TextCode></ofd:TextObject></ofd:Content></ofd:Page>` }
  ]);
}

function createStepSample(): File {
  return new File(
    [[
      "ISO-10303-21;",
      "DATA;",
      "#1 = CARTESIAN_POINT('P1',(1.,2.,3.));",
      "#2 = DIRECTION('D1',(0.,0.,1.));",
      "#3 = LINE('L1',#1,#2);",
      "ENDSEC;",
      "END-ISO-10303-21;"
    ].join("\n")],
    "sample-model.step",
    { type: "model/step" }
  );
}

function createObjSample(): File {
  return new File(
    [[
      "o OpenFileViewerPyramid",
      "v 0 0 0",
      "v 1 0 0",
      "v 0.5 0.86 0",
      "v 0.5 0.3 0.75",
      "f 1 2 3",
      "f 1 2 4"
    ].join("\n")],
    "sample-model.obj",
    { type: "model/obj" }
  );
}

function createDwgSample(): File {
  const bytes = new Uint8Array(ascii("AC1027\0\0DWGDATA\0LINE\0LAYER A-WALL\0"));
  return new File([bytes.buffer], "sample-plan.dwg", { type: "application/acad" });
}

function ascii(str: string): number[] {
  return str.split("").map(c => c.charCodeAt(0));
}

function createZipFile(filename: string, mimeType: string, entries: { path: string; content: string }[]): File {
  return new File([new Uint8Array([80, 75, 3, 4])], filename, { type: mimeType });
}

function createDocxSample(): File {
  const contentTypes = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`;
  const rels = `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`;
  const documentXml = `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:rPr><w:b/><w:sz w:val="36"/></w:rPr><w:t>Open File Viewer</w:t></w:r></w:p>
    <w:p><w:t>Office document preview sample.</w:t></w:p>
  </w:body>
</w:document>`;
  return createZipFile("sample-word.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", [
    { path: "[Content_Types].xml", content: contentTypes },
    { path: "_rels/.rels", content: rels },
    { path: "word/document.xml", content: documentXml }
  ]);
}

function createXlsxSample(): File {
  const contentTypes = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
</Types>`;
  const workbook = `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheets><sheet name="Overview" sheetId="1"/></sheets>
</workbook>`;
  return createZipFile("sample-excel.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", [
    { path: "[Content_Types].xml", content: contentTypes },
    { path: "xl/workbook.xml", content: workbook }
  ]);
}

function createPptxSample(): File {
  const contentTypes = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>
</Types>`;
  const presentation = `<?xml version="1.0" encoding="UTF-8"?>
<p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:sldIdLst><p:sldId id="256"/></p:sldIdLst>
</p:presentation>`;
  return createZipFile("sample-powerpoint.pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation", [
    { path: "[Content_Types].xml", content: contentTypes },
    { path: "ppt/presentation.xml", content: presentation }
  ]);
}
