import './style.css'
import {
  Scene,
  Color,
  Fog,
  PerspectiveCamera,
  WebGLRenderer,
  AmbientLight,
  DirectionalLight,
  PlaneGeometry,
  MeshBasicMaterial,
  Mesh,
  GridHelper,
  ShaderMaterial,
  Vector3,
  Vector2,
  Raycaster,
  Group,
  DataTexture,
  RGBAFormat,
  NearestFilter,
  MeshToonMaterial,
  Clock,
  BufferGeometry,
  Float32BufferAttribute,
} from 'three';

const SCENE_BG = 0x1a1a1a;
const MAX_PIXEL_RATIO = 2;
const GOPHER_COUNT = 20;

const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

function getViewportSize() {
  const width = Math.max(window.innerWidth || 0, 1);
  const height = Math.max(window.innerHeight || 0, 1);
  return { width, height };
}

const { width: initialWidth, height: initialHeight } = getViewportSize();

const scene = new Scene();
scene.background = new Color(SCENE_BG);
scene.fog = new Fog(SCENE_BG, 10, 50);

const camera = new PerspectiveCamera(75, initialWidth / initialHeight, 0.1, 1000);
camera.position.z = 10;
camera.position.y = 8;
camera.lookAt(0, 0, 0);

let renderer = null;
let isWebGLAvailable = false;
let isContextLost = false;

try {
  renderer = new WebGLRenderer({
    antialias: !prefersReducedMotion,
    powerPreference: 'high-performance',
    failIfMajorPerformanceCaveat: false,
  });
  renderer.setSize(initialWidth, initialHeight);
  renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, MAX_PIXEL_RATIO));
  renderer.shadowMap.enabled = false;
  renderer.domElement.setAttribute('aria-hidden', 'true');
  document.body.appendChild(renderer.domElement);
  isWebGLAvailable = true;

  renderer.domElement.addEventListener(
    'webglcontextlost',
    (event) => {
      event.preventDefault();
      isContextLost = true;
      console.warn('WebGL context lost. Pausing rendering.');
    },
    false
  );

  renderer.domElement.addEventListener(
    'webglcontextrestored',
    () => {
      isContextLost = false;
      console.info('WebGL context restored. Resuming rendering.');
      if (gridHelper && gridMaterial) {
        gridMaterial.needsUpdate = true;
      }
    },
    false
  );
} catch (error) {
  console.warn('WebGL initialization failed, falling back to 2D UI mode:', error);
  isWebGLAvailable = false;
}

const ambientLight = new AmbientLight(0xffffff, 0.6);
scene.add(ambientLight);

const dirLight = new DirectionalLight(0xffffff, 1.2);
dirLight.position.set(10, 20, 10);
dirLight.castShadow = false;
scene.add(dirLight);

const planeGeometry = new PlaneGeometry(100, 100);
const planeMaterial = new MeshBasicMaterial({ visible: false });
const groundPlane = new Mesh(planeGeometry, planeMaterial);
groundPlane.rotation.x = -Math.PI / 2;
scene.add(groundPlane);

const gridHelper = new GridHelper(100, 50, 0x00aaaa, 0x353535);

const gridVertexShader = `
  varying vec3 vColor;
  varying vec3 vWorldPosition;
  void main() {
    vColor = color;
    vec4 worldPosition = modelMatrix * vec4(position, 1.0);
    vWorldPosition = worldPosition.xyz;
    gl_Position = projectionMatrix * viewMatrix * worldPosition;
  }
`;

const gridFragmentShader = `
  uniform vec3 uCursor;
  uniform float uRadius;
  varying vec3 vColor;
  varying vec3 vWorldPosition;

  void main() {
    float dist = distance(vWorldPosition.xz, uCursor.xz);
    float intensity = 1.0 - smoothstep(0.0, uRadius, dist);
    vec3 glowColor = vec3(0.0, 1.0, 1.0);
    vec3 finalColor = vColor + (glowColor * intensity * 0.8);
    gl_FragColor = vec4(finalColor, 1.0);
  }
`;

const gridMaterial = new ShaderMaterial({
  vertexShader: gridVertexShader,
  fragmentShader: gridFragmentShader,
  vertexColors: true,
  uniforms: {
    uCursor: { value: new Vector3(0, 0, 0) },
    uRadius: { value: 10.0 },
  },
  transparent: true,
});

gridHelper.material = gridMaterial;
scene.add(gridHelper);

function createGradientMap() {
  const colors = new Uint8Array([
    150, 150, 150, 255,
    200, 200, 200, 255,
    255, 255, 255, 255,
  ]);
  const texture = new DataTexture(colors, 3, 1, RGBAFormat);
  texture.minFilter = NearestFilter;
  texture.magFilter = NearestFilter;
  texture.needsUpdate = true;
  return texture;
}

const toonGradient = createGradientMap();
const gophers = [];
let gopherTemplate = null;

async function loadQuantizedBin(url) {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`HTTP ${res.status} loading ${url}`);
  const buf = await res.arrayBuffer();
  const view = new DataView(buf);

  const magic = String.fromCharCode(
    view.getUint8(0),
    view.getUint8(1),
    view.getUint8(2),
    view.getUint8(3)
  );
  if (magic !== 'GQ16') {
    throw new Error(`Invalid model magic: ${magic}`);
  }

  const vertexCount = view.getUint16(4, true);
  const triCount = view.getUint16(6, true);
  const minX = view.getFloat32(8, true);
  const minY = view.getFloat32(12, true);
  const minZ = view.getFloat32(16, true);
  const scaleX = view.getFloat32(20, true) / 65535;
  const scaleY = view.getFloat32(24, true) / 65535;
  const scaleZ = view.getFloat32(28, true) / 65535;

  const headerBytes = 32;
  const qVerts = new Uint16Array(buf, headerBytes, vertexCount * 3);
  const indices = new Uint16Array(buf, headerBytes + vertexCount * 6, triCount * 3);

  const positions = new Float32Array(triCount * 9);
  for (let i = 0; i < triCount * 3; i++) {
    const vi = indices[i] * 3;
    const pi = i * 3;
    positions[pi] = minX + qVerts[vi] * scaleX;
    positions[pi + 1] = minY + qVerts[vi + 1] * scaleY;
    positions[pi + 2] = minZ + qVerts[vi + 2] * scaleZ;
  }

  const geometry = new BufferGeometry();
  geometry.setAttribute('position', new Float32BufferAttribute(positions, 3));
  return geometry;
}

if (isWebGLAvailable) {
  loadQuantizedBin('/model/go_gopher_low_0.02.bin')
    .then((geometry) => {
      geometry.computeVertexNormals();

      const mesh = new Mesh(
        geometry,
        new MeshToonMaterial({
          color: 0x00ADD8,
          gradientMap: toonGradient,
        })
      );

      geometry.computeBoundingBox();
      const box = geometry.boundingBox;
      const size = new Vector3();
      const center = new Vector3();
      box.getSize(size);
      box.getCenter(center);

      let maxDim = Math.max(size.x, size.y, size.z);
      if (maxDim < 0.001 || !isFinite(maxDim)) {
        maxDim = 40;
      }

      const scale = 1.5 / maxDim;
      mesh.position.copy(center).multiplyScalar(-scale);
      mesh.scale.set(scale, scale, scale);

      gopherTemplate = new Group();
      gopherTemplate.add(mesh);

      for (let i = 0; i < GOPHER_COUNT; i++) {
        spawnGopher();
      }
    })
    .catch((error) => {
      console.warn('Could not load gopher model (will show grid only):', error);
    });
}

function spawnGopher() {
  if (!gopherTemplate) return;

  const pivot = new Group();
  const visual = gopherTemplate.clone();

  visual.traverse((child) => {
    if (child.isMesh) {
      child.material = child.material.clone();
      child.material.gradientMap = toonGradient;
      const hue = Math.random();
      const saturation = 0.8 + Math.random() * 0.2;
      const lightness = 0.5 + Math.random() * 0.1;
      child.material.color.setHSL(hue, saturation, lightness);
    }
  });

  visual.position.y = 0.75;
  pivot.add(visual);
  pivot.position.x = (Math.random() - 0.5) * 40;
  pivot.position.z = (Math.random() - 0.5) * 40;
  pivot.rotation.y = Math.random() * Math.PI * 2;
  scene.add(pivot);

  gophers.push({
    pivot,
    visual,
    speed: 0.05 + Math.random() * 0.15,
    targetOffset: new Vector3(
      (Math.random() - 0.5) * 5,
      0,
      (Math.random() - 0.5) * 5
    ),
    velocity: new Vector3(),
  });
}

const raycaster = new Raycaster();
const pointer = new Vector2();
const targetPoint = new Vector3();
const personalTarget = new Vector3();
const diff = new Vector3();
const motionDir = new Vector3();
const explosionDiff = new Vector3();
const kick = new Vector3();
const offscreenCursor = new Vector3(1000, 1000, 1000);

const app = document.querySelector('#app') || document.body;
const urlParams = new URLSearchParams(window.location.search);
const showGophersOnly = urlParams.get('show_gophers_only') === 'true';

if (!document.getElementById('scene-desc')) {
  const sceneDescription = document.createElement('p');
  sceneDescription.className = 'visually-hidden';
  sceneDescription.id = 'scene-desc';
  sceneDescription.textContent =
    'Interactive 3D scene: colorful Go gophers follow the pointer across a dark grid. Links and site chrome appear above the scene.';
  document.body.insertBefore(sceneDescription, app);
}

const ABOUT_PATH = '/about';
const ABOUT_BUTTON_LABEL = '\u{1F471}\u{1F3FB}\u200D\u2642\uFE0F About';
const WIP_PATH = '/wip';

function normalizePathname(pathname = window.location.pathname) {
  if (!pathname || pathname === '/') return '/';
  return pathname.endsWith('/') ? pathname.slice(0, -1) : pathname;
}

function isAboutRoute(pathname = window.location.pathname) {
  return normalizePathname(pathname) === ABOUT_PATH;
}

function isWipRoute(pathname = window.location.pathname) {
  return normalizePathname(pathname) === WIP_PATH;
}

function isAboutEmpty(about) {
  if (!about || typeof about !== 'object') return true;
  const hero = (about.Hero ?? about.hero ?? '').trim();
  const name = (about.Name ?? about.name ?? '').trim();
  const title = (about.Title ?? about.title ?? '').trim();
  const desc = (about.Description ?? about.description ?? '').trim();
  return hero === '' && name === '' && title === '' && desc === '';
}

const initialPath = normalizePathname();
const isSubPath = initialPath !== '/' && !isAboutRoute(initialPath) && !isWipRoute(initialPath);

function buildChrome() {
  const header = document.createElement('a');
  header.className = 'site-header';
  header.href = '/';
  header.textContent = 'HOMIN-DEV';
  app.appendChild(header);

  const main = document.createElement('main');
  main.className = 'center-content';
  main.setAttribute('aria-label', 'Site links');
  app.appendChild(main);

  if (isSubPath) {
    const backLink = document.createElement('a');
    backLink.className = 'back-link';
    backLink.href = '/';

    const arrow = document.createElement('span');
    arrow.className = 'back-arrow';
    arrow.setAttribute('aria-hidden', 'true');
    arrow.textContent = '←';

    const label = document.createElement('span');
    label.textContent = 'Home';

    backLink.append(arrow, ' ', label);
    main.appendChild(backLink);
  }

  const linksContainer = document.createElement('nav');
  linksContainer.className = 'links-container';
  linksContainer.setAttribute('aria-label', 'Destinations');
  main.appendChild(linksContainer);

  const status = document.createElement('p');
  status.className = 'links-status';
  status.hidden = true;
  main.appendChild(status);

  const fishContainer = document.createElement('blockquote');
  fishContainer.className = 'fish-container';
  main.appendChild(fishContainer);

  const footer = document.createElement('footer');
  footer.className = 'site-footer';
  footer.innerHTML =
    '&copy; Homin Lee &lt;i@homin.dev&gt; All rights reserved.<br>The Go gopher was designed by Renee French.';
  app.appendChild(footer);

  return { linksContainer, fishContainer, statusEl: status };
}

let linksContainer = null;
let fishContainer = null;
let statusEl = null;
let aboutOverlay = null;
let wipOverlay = null;
let aboutData = null;
let loadedLinks = null;
let lastAboutTrigger = null;
let lastWipTrigger = null;

function isSafeMarkdownURL(raw) {
  const trimmed = (raw || '').trim();
  if (!trimmed) return false;
  if (trimmed.startsWith('/') && !trimmed.startsWith('//')) return true;
  try {
    const parsed = new URL(trimmed);
    return parsed.protocol === 'http:' || parsed.protocol === 'https:' || parsed.protocol === 'mailto:';
  } catch {
    return false;
  }
}

function parseMarkdownHeading(line) {
  let level = 0;
  while (level < line.length && level < 6 && line[level] === '#') {
    level++;
  }
  if (level === 0 || level >= line.length || line[level] !== ' ') {
    return null;
  }
  return { level, content: line.slice(level + 1).trim() };
}

function parseUnorderedListItem(line) {
  for (const prefix of ['- ', '* ', '+ ']) {
    if (line.startsWith(prefix)) {
      return line.slice(prefix.length).trim();
    }
  }
  return null;
}

function parseOrderedListItem(line) {
  const match = /^(\d+)\.\s+(.*)$/.exec(line);
  return match ? match[2].trim() : null;
}

function findEarliestInlineToken(s) {
  let best = null;
  const consider = (kind, start, end, a, b = '') => {
    if (start >= 0 && (!best || start < best.start)) {
      best = { kind, start, end, a, b };
    }
  };

  // Inline code: `...`
  const codeStart = s.indexOf('`');
  if (codeStart >= 0) {
    const codeClose = s.indexOf('`', codeStart + 1);
    if (codeClose > codeStart + 1) {
      consider('code', codeStart, codeClose + 1, s.slice(codeStart + 1, codeClose));
    }
  }

  // Image: ![alt](url)
  const imgStart = s.indexOf('![');
  if (imgStart >= 0) {
    const mid = s.indexOf('](', imgStart + 2);
    if (mid >= 0) {
      const close = s.indexOf(')', mid + 2);
      if (close >= 0) {
        consider('img', imgStart, close + 1, s.slice(imgStart + 2, mid), s.slice(mid + 2, close));
      }
    }
  }

  // Link: [label](url)
  for (let searchFrom = 0; searchFrom < s.length; ) {
    const st = s.indexOf('[', searchFrom);
    if (st < 0) break;
    if (st > 0 && s[st - 1] === '!') {
      searchFrom = st + 1;
      continue;
    }
    const mid = s.indexOf('](', st + 1);
    if (mid >= 0) {
      const close = s.indexOf(')', mid + 2);
      if (close >= 0) {
        consider('link', st, close + 1, s.slice(st + 1, mid), s.slice(mid + 2, close));
        break;
      }
    }
    searchFrom = st + 1;
  }

  // Bold: **...** or __...__
  for (const delim of ['**', '__']) {
    const st = s.indexOf(delim);
    if (st >= 0) {
      const close = s.indexOf(delim, st + 2);
      if (close > st + 2) {
        consider('bold', st, close + 2, s.slice(st + 2, close));
      }
    }
  }

  // Italic: *...* or _..._
  for (const delim of ['*', '_']) {
    for (let searchFrom = 0; searchFrom < s.length; ) {
      const st = s.indexOf(delim, searchFrom);
      if (st < 0) break;
      if (st + 1 < s.length && s[st + 1] === delim) {
        searchFrom = st + 2;
        continue;
      }
      const close = s.indexOf(delim, st + 1);
      if (close > st + 1) {
        consider('italic', st, close + 1, s.slice(st + 1, close));
        break;
      }
      searchFrom = st + 1;
    }
  }

  return best;
}

function appendInlineMarkdown(parent, text) {
  let remaining = text;
  while (remaining.length > 0) {
    const token = findEarliestInlineToken(remaining);
    if (!token) {
      parent.appendChild(document.createTextNode(remaining));
      return;
    }
    if (token.start > 0) {
      parent.appendChild(document.createTextNode(remaining.slice(0, token.start)));
    }
    switch (token.kind) {
      case 'code': {
        const code = document.createElement('code');
        code.textContent = token.a;
        parent.appendChild(code);
        break;
      }
      case 'img': {
        const src = token.b.trim();
        if (isSafeMarkdownURL(src)) {
          const img = document.createElement('img');
          img.src = src;
          img.alt = token.a;
          parent.appendChild(img);
        } else {
          parent.appendChild(document.createTextNode(token.a));
        }
        break;
      }
      case 'link': {
        const href = token.b.trim();
        if (isSafeMarkdownURL(href)) {
          const a = document.createElement('a');
          a.href = href;
          if (href.startsWith('http://') || href.startsWith('https://')) {
            a.target = '_blank';
            a.rel = 'noopener noreferrer';
          }
          appendInlineMarkdown(a, token.a);
          parent.appendChild(a);
        } else {
          appendInlineMarkdown(parent, token.a);
        }
        break;
      }
      case 'bold': {
        const strong = document.createElement('strong');
        appendInlineMarkdown(strong, token.a);
        parent.appendChild(strong);
        break;
      }
      case 'italic': {
        const em = document.createElement('em');
        appendInlineMarkdown(em, token.a);
        parent.appendChild(em);
        break;
      }
    }
    remaining = remaining.slice(token.end);
  }
}

function appendParagraphLines(p, lines) {
  lines.forEach((line, idx) => {
    if (idx > 0) {
      p.appendChild(document.createElement('br'));
    }
    appendInlineMarkdown(p, line);
  });
}

function renderMarkdownNodes(parent, md) {
  const lines = (md || '').replace(/\r\n/g, '\n').split('\n');
  let i = 0;
  while (i < lines.length) {
    const trimmed = lines[i].trim();
    if (!trimmed) {
      i++;
      continue;
    }

    if (trimmed.startsWith('```')) {
      i++;
      const codeLines = [];
      while (i < lines.length && !lines[i].trim().startsWith('```')) {
        codeLines.push(lines[i]);
        i++;
      }
      if (i < lines.length) i++;
      const pre = document.createElement('pre');
      const code = document.createElement('code');
      code.textContent = codeLines.join('\n');
      pre.appendChild(code);
      parent.appendChild(pre);
      continue;
    }

    if (trimmed === '---' || trimmed === '***' || trimmed === '___') {
      parent.appendChild(document.createElement('hr'));
      i++;
      continue;
    }

    const heading = parseMarkdownHeading(trimmed);
    if (heading) {
      const h = document.createElement(`h${heading.level}`);
      appendInlineMarkdown(h, heading.content);
      parent.appendChild(h);
      i++;
      continue;
    }

    if (trimmed.startsWith('>')) {
      const quoteLines = [];
      while (i < lines.length && lines[i].trim().startsWith('>')) {
        quoteLines.push(lines[i].trim().slice(1).trim());
        i++;
      }
      const bq = document.createElement('blockquote');
      const p = document.createElement('p');
      appendParagraphLines(p, quoteLines);
      bq.appendChild(p);
      parent.appendChild(bq);
      continue;
    }

    const ulItem = parseUnorderedListItem(trimmed);
    if (ulItem !== null) {
      const ul = document.createElement('ul');
      while (i < lines.length) {
        const item = parseUnorderedListItem(lines[i].trim());
        if (item === null) break;
        const li = document.createElement('li');
        appendInlineMarkdown(li, item);
        ul.appendChild(li);
        i++;
      }
      parent.appendChild(ul);
      continue;
    }

    const olItem = parseOrderedListItem(trimmed);
    if (olItem !== null) {
      const ol = document.createElement('ol');
      while (i < lines.length) {
        const item = parseOrderedListItem(lines[i].trim());
        if (item === null) break;
        const li = document.createElement('li');
        appendInlineMarkdown(li, item);
        ol.appendChild(li);
        i++;
      }
      parent.appendChild(ol);
      continue;
    }

    const paraLines = [];
    while (i < lines.length) {
      const t = lines[i].trim();
      if (
        !t ||
        t.startsWith('```') ||
        t === '---' ||
        t === '***' ||
        t === '___' ||
        t.startsWith('>') ||
        parseMarkdownHeading(t) ||
        parseUnorderedListItem(t) !== null ||
        parseOrderedListItem(t) !== null
      ) {
        break;
      }
      paraLines.push(t);
      i++;
    }
    if (paraLines.length > 0) {
      const p = document.createElement('p');
      appendParagraphLines(p, paraLines);
      parent.appendChild(p);
    }
  }
}

function bindAboutOverlayEvents(overlay) {
  if (!overlay || overlay.dataset.bound === 'true') return;
  overlay.dataset.bound = 'true';

  overlay.addEventListener('pointerdown', (event) => {
    event.stopPropagation();
  });

  overlay.addEventListener('click', (event) => {
    if (event.target === overlay) {
      closeAboutModal({ pushHistory: true });
    }
  });

  const closeBtn = overlay.querySelector('.about-close');
  if (closeBtn) {
    closeBtn.addEventListener('click', (event) => {
      event.preventDefault();
      closeAboutModal({ pushHistory: true });
    });
  }
}

function renderAboutModal(about) {
  if (isAboutEmpty(about)) return null;

  const hero = (about.Hero ?? about.hero ?? '').trim();
  const name = (about.Name ?? about.name ?? '').trim();
  const title = (about.Title ?? about.title ?? '').trim();
  const description = (about.Description ?? about.description ?? '').trim();

  if (!aboutOverlay) {
    aboutOverlay = document.getElementById('about-overlay');
  }
  const wasOpen = aboutOverlay ? !aboutOverlay.hidden : isAboutRoute();

  if (!aboutOverlay) {
    aboutOverlay = document.createElement('div');
    aboutOverlay.className = 'about-overlay';
    aboutOverlay.id = 'about-overlay';
    aboutOverlay.hidden = !wasOpen;
    app.appendChild(aboutOverlay);
  }

  aboutOverlay.replaceChildren();
  delete aboutOverlay.dataset.bound;

  const modal = document.createElement('div');
  modal.className = 'about-modal';
  modal.setAttribute('role', 'dialog');
  modal.setAttribute('aria-modal', 'true');
  modal.setAttribute('aria-labelledby', 'about-modal-title');

  const closeBtn = document.createElement('a');
  closeBtn.className = 'about-close';
  closeBtn.href = '/';
  closeBtn.setAttribute('aria-label', 'Close about dialog');
  closeBtn.textContent = '×';
  modal.appendChild(closeBtn);

  if (hero && isSafeMarkdownURL(hero)) {
    const heroImg = document.createElement('img');
    heroImg.className = 'about-hero';
    heroImg.src = hero;
    heroImg.alt = name || title || 'HOMIN-DEV';
    modal.appendChild(heroImg);
  }

  if (name) {
    const nameEl = document.createElement('h2');
    nameEl.className = 'about-name';
    nameEl.id = 'about-modal-title';
    nameEl.textContent = name;
    modal.appendChild(nameEl);

    if (title) {
      const titleEl = document.createElement('p');
      titleEl.className = 'about-title';
      titleEl.textContent = title;
      modal.appendChild(titleEl);
    }
  } else if (title) {
    const titleEl = document.createElement('h2');
    titleEl.className = 'about-title';
    titleEl.id = 'about-modal-title';
    titleEl.textContent = title;
    modal.appendChild(titleEl);
  } else {
    const defaultEl = document.createElement('h2');
    defaultEl.className = 'about-title';
    defaultEl.id = 'about-modal-title';
    defaultEl.textContent = 'About';
    modal.appendChild(defaultEl);
  }

  if (description) {
    const descEl = document.createElement('div');
    descEl.className = 'about-description';
    renderMarkdownNodes(descEl, description);
    modal.appendChild(descEl);
  }

  aboutOverlay.appendChild(modal);
  aboutOverlay.hidden = !wasOpen;
  bindAboutOverlayEvents(aboutOverlay);
  return aboutOverlay;
}

function openAboutModal({ pushHistory = true, triggerEl = null } = {}) {
  if (!aboutOverlay && aboutData) {
    renderAboutModal(aboutData);
  }
  if (!aboutOverlay) return;

  if (wipOverlay && !wipOverlay.hidden) {
    closeWipModal({ pushHistory: false });
  }

  lastAboutTrigger =
    triggerEl || (linksContainer && linksContainer.querySelector('a[href="/about"]')) || null;
  aboutOverlay.hidden = false;

  if (pushHistory && !isAboutRoute()) {
    window.history.pushState({ aboutModal: true }, '', ABOUT_PATH);
  }

  const closeBtn = aboutOverlay.querySelector('.about-close');
  if (closeBtn && typeof closeBtn.focus === 'function') {
    closeBtn.focus();
  }
}

function closeAboutModal({ pushHistory = true } = {}) {
  if (!aboutOverlay || aboutOverlay.hidden) return;
  aboutOverlay.hidden = true;

  if (pushHistory && isAboutRoute()) {
    window.history.pushState(null, '', '/');
  }

  if (lastAboutTrigger && typeof lastAboutTrigger.focus === 'function') {
    lastAboutTrigger.focus();
  }
}

function isAboutLink(link) {
  if (!link) return false;
  const href = (link.Link || '').trim();
  const name = (link.Name || '').trim();
  return normalizePathname(href) === ABOUT_PATH || normalizePathname(name) === ABOUT_PATH;
}

function bindAboutCardClick(card) {
  if (!card || card.dataset.aboutBound === 'true') return;
  card.dataset.aboutBound = 'true';
  card.addEventListener('click', (event) => {
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey || event.button !== 0) {
      return;
    }
    event.preventDefault();
    openAboutModal({ pushHistory: true, triggerEl: card });
  });
}

function bindWipOverlayEvents(overlay) {
  if (!overlay || overlay.dataset.bound === 'true') return;
  overlay.dataset.bound = 'true';

  overlay.addEventListener('pointerdown', (event) => {
    event.stopPropagation();
  });

  overlay.addEventListener('click', (event) => {
    if (event.target === overlay) {
      closeWipModal({ pushHistory: true });
    }
  });

  const closeBtn = overlay.querySelector('.wip-close');
  if (closeBtn) {
    closeBtn.addEventListener('click', (event) => {
      event.preventDefault();
      closeWipModal({ pushHistory: true });
    });
  }
}

function renderWipModal() {
  if (!wipOverlay) {
    wipOverlay = document.getElementById('wip-overlay');
  }
  const wasOpen = wipOverlay ? !wipOverlay.hidden : isWipRoute();

  if (!wipOverlay) {
    wipOverlay = document.createElement('div');
    wipOverlay.className = 'wip-overlay';
    wipOverlay.id = 'wip-overlay';
    wipOverlay.hidden = !wasOpen;
    app.appendChild(wipOverlay);
  }

  wipOverlay.replaceChildren();
  delete wipOverlay.dataset.bound;

  const modal = document.createElement('div');
  modal.className = 'wip-modal';
  modal.setAttribute('role', 'dialog');
  modal.setAttribute('aria-modal', 'true');
  modal.setAttribute('aria-labelledby', 'wip-modal-title');

  const closeBtn = document.createElement('a');
  closeBtn.className = 'wip-close';
  closeBtn.href = '/';
  closeBtn.setAttribute('aria-label', 'Close dialog');
  closeBtn.textContent = '×';
  modal.appendChild(closeBtn);

  const titleEl = document.createElement('h2');
  titleEl.className = 'visually-hidden';
  titleEl.id = 'wip-modal-title';
  titleEl.textContent = 'Work In Progress';
  modal.appendChild(titleEl);

  const frame = document.createElement('div');
  frame.className = 'wip-frame';
  const img = document.createElement('img');
  img.className = 'wip-img';
  img.src = '/wip.webp';
  img.alt = 'Work In Progress — Gopher debugging: log.Print("h.e.r.e")';
  img.width = 711;
  img.height = 711;
  frame.appendChild(img);
  modal.appendChild(frame);

  const caption = document.createElement('p');
  caption.className = 'wip-caption';
  caption.textContent = 'Work in Progress';
  modal.appendChild(caption);

  wipOverlay.appendChild(modal);
  wipOverlay.hidden = !wasOpen;
  bindWipOverlayEvents(wipOverlay);
  return wipOverlay;
}

function openWipModal({ pushHistory = true, triggerEl = null } = {}) {
  if (!wipOverlay) {
    renderWipModal();
  }
  if (!wipOverlay) return;

  if (aboutOverlay && !aboutOverlay.hidden) {
    closeAboutModal({ pushHistory: false });
  }

  lastWipTrigger = triggerEl || (linksContainer && linksContainer.querySelector('a[href="/wip"]')) || null;
  wipOverlay.hidden = false;

  if (pushHistory && !isWipRoute()) {
    window.history.pushState({ wipModal: true }, '', WIP_PATH);
  }

  const closeBtn = wipOverlay.querySelector('.wip-close');
  if (closeBtn && typeof closeBtn.focus === 'function') {
    closeBtn.focus();
  }
}

function closeWipModal({ pushHistory = true } = {}) {
  if (!wipOverlay || wipOverlay.hidden) return;
  wipOverlay.hidden = true;

  if (pushHistory && isWipRoute()) {
    window.history.pushState(null, '', '/');
  }

  if (lastWipTrigger && typeof lastWipTrigger.focus === 'function') {
    lastWipTrigger.focus();
  }
}

function isWipLink(link) {
  if (!link) return false;
  const href = (link.Link || '').trim();
  const name = (link.Name || '').trim();
  return normalizePathname(href) === WIP_PATH || normalizePathname(name) === WIP_PATH;
}

function bindWipCardClick(card) {
  if (!card || card.dataset.wipBound === 'true') return;
  card.dataset.wipBound = 'true';
  card.addEventListener('click', (event) => {
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey || event.button !== 0) {
      return;
    }
    event.preventDefault();
    openWipModal({ pushHistory: true, triggerEl: card });
  });
}

window.addEventListener('keydown', (event) => {
  if (event.key === 'Escape') {
    if (aboutOverlay && !aboutOverlay.hidden) {
      closeAboutModal({ pushHistory: true });
    } else if (wipOverlay && !wipOverlay.hidden) {
      closeWipModal({ pushHistory: true });
    }
  }
});

window.addEventListener('popstate', () => {
  if (isAboutRoute()) {
    if (wipOverlay && !wipOverlay.hidden) {
      closeWipModal({ pushHistory: false });
    }
    openAboutModal({ pushHistory: false });
  } else if (isWipRoute()) {
    if (aboutOverlay && !aboutOverlay.hidden) {
      closeAboutModal({ pushHistory: false });
    }
    openWipModal({ pushHistory: false });
  } else {
    closeAboutModal({ pushHistory: false });
    closeWipModal({ pushHistory: false });
  }
});

if (showGophersOnly) {
  app.replaceChildren();
  const footer = document.createElement('footer');
  footer.className = 'site-footer';
  footer.innerHTML =
    '&copy; Homin Lee &lt;i@homin.dev&gt; All rights reserved.<br>The Go gopher was designed by Renee French.';
  app.appendChild(footer);
} else {
  // Check if backend already prerendered the chrome into #app (see render.go).
  // Re-use it directly so there is zero empty-screen flash while hydration and
  // API fetch occur.
  const existingLinks = app.querySelector('.links-container');
  if (existingLinks) {
    linksContainer = existingLinks;
    fishContainer = app.querySelector('.fish-container');
    statusEl = app.querySelector('.links-status');
    aboutOverlay = document.getElementById('about-overlay');

    if (aboutOverlay) {
      bindAboutOverlayEvents(aboutOverlay);
      if (isAboutRoute()) {
        aboutOverlay.hidden = false;
      }
    }

    wipOverlay = document.getElementById('wip-overlay');
    if (wipOverlay) {
      bindWipOverlayEvents(wipOverlay);
      if (isWipRoute()) {
        wipOverlay.hidden = false;
        const closeBtn = wipOverlay.querySelector('.wip-close');
        if (closeBtn && typeof closeBtn.focus === 'function') {
          closeBtn.focus();
        }
      }
    }

    const prerenderedAboutCard = linksContainer.querySelector('a[href="/about"]');
    if (prerenderedAboutCard) {
      bindAboutCardClick(prerenderedAboutCard);
    }

    const prerenderedWipCard = linksContainer.querySelector('a[href="/wip"]');
    if (prerenderedWipCard) {
      bindWipCardClick(prerenderedWipCard);
    }

    const main = app.querySelector('.center-content') || app;
    if (!fishContainer) {
      fishContainer = document.createElement('blockquote');
      fishContainer.className = 'fish-container';
      main.appendChild(fishContainer);
    }
    if (!statusEl) {
      statusEl = document.createElement('p');
      statusEl.className = 'links-status';
      statusEl.hidden = true;
      main.appendChild(statusEl);
    }
  } else {
    // Prerendered DOM not present (e.g. static shell fallback) -> build chrome
    app.replaceChildren();
    ({ linksContainer, fishContainer, statusEl } = buildChrome());
    renderWipModal();
    if (isWipRoute()) {
      openWipModal({ pushHistory: false });
    }
  }
}

const mockLinks = [
  { Name: 'Blog', Link: 'https://homin.dev/blog', Desc: 'Tech Blog' },
  { Name: 'GitHub', Link: 'https://github.com/suapapa', Desc: 'My Code' },
  { Name: 'LinkedIn', Link: 'https://linkedin.com/in/homin', Desc: 'Profile' },
];

async function fetchAbout() {
  try {
    const res = await fetch('/api/about');
    if (!res.ok) return;
    const data = await res.json();
    if (isAboutEmpty(data)) return;

    aboutData = data;
    renderAboutModal(aboutData);

    if (loadedLinks !== null) {
      renderLinks(loadedLinks);
    }
    if (isAboutRoute()) {
      openAboutModal({ pushHistory: false });
    }
  } catch (e) {
    console.warn('Fetch about failed:', e);
  }
}

async function fetchLinks() {
  const hasExistingLinks = linksContainer && linksContainer.children.length > 0;
  if (!hasExistingLinks && statusEl) {
    statusEl.hidden = false;
    statusEl.textContent = 'Loading links…';
  }

  try {
    let url = '/api/links';
    if (isSubPath) {
      const prefix = normalizePathname().substring(1);
      url = `/api/links?prefix=${encodeURIComponent(prefix)}`;
    }

    const res = await fetch(url);
    if (!res.ok) throw new Error('Failed to fetch');
    const data = await res.json();
    loadedLinks = Array.isArray(data) ? data : [];
    renderLinks(loadedLinks);
  } catch (e) {
    console.warn('Fetch links failed:', e);
    if (!hasExistingLinks) {
      if (statusEl) {
        statusEl.hidden = false;
        statusEl.textContent = 'Showing offline links (API unavailable).';
      }
      loadedLinks = mockLinks;
      renderLinks(loadedLinks);
    } else if (statusEl) {
      statusEl.hidden = true;
    }
  }
}

async function fetchFish() {
  if (!fishContainer) return;
  try {
    const res = await fetch('/api/fish');
    if (!res.ok) return;
    const data = await res.json();
    fishContainer.replaceChildren();

    const line = document.createElement('p');
    line.className = 'fish-line';
    line.textContent = `\u201C${data.Line ?? ''}\u201D`;

    const char = document.createElement('cite');
    char.className = 'fish-char';
    char.textContent = `- ${data.Character ?? ''} (${data.Movie ?? ''})`;

    fishContainer.append(line, char);
  } catch (e) {
    console.warn('Fish fetch failed:', e);
    fishContainer.replaceChildren();
  }
}

function renderLinks(links) {
  if (!linksContainer) return;
  linksContainer.replaceChildren();
  if (statusEl) statusEl.hidden = true;

  const effectiveLinks = Array.isArray(links) ? [...links] : [];
  if (!isSubPath && !isAboutEmpty(aboutData) && !effectiveLinks.some(isAboutLink)) {
    effectiveLinks.unshift({
      Name: ABOUT_PATH,
      Link: ABOUT_PATH,
      Desc: ABOUT_BUTTON_LABEL,
    });
  }

  if (effectiveLinks.length === 0) {
    if (statusEl) {
      statusEl.hidden = false;
      statusEl.textContent = 'No links available.';
    }
    return;
  }

  effectiveLinks.forEach((link) => {
    const card = document.createElement('a');
    card.href = link.Link;
    card.className = 'link-card';

    const label = document.createElement('span');
    label.textContent = link.Desc || link.Name || 'Link';

    const sub = document.createElement('span');
    sub.className = 'link-sub';
    sub.textContent = link.Name || '';

    card.append(label, sub);
    if (isAboutLink(link)) {
      bindAboutCardClick(card);
    } else if (isWipLink(link)) {
      bindWipCardClick(card);
    }
    linksContainer.appendChild(card);
  });
}

if (!showGophersOnly) {
  fetchAbout();
  fetchLinks();
  fetchFish();
}

function updatePointerFromEvent(event) {
  const x = event.clientX ?? (event.touches && event.touches[0]?.clientX);
  const y = event.clientY ?? (event.touches && event.touches[0]?.clientY);
  if (x == null || y == null) return;
  const { width, height } = getViewportSize();
  pointer.x = (x / width) * 2 - 1;
  pointer.y = -(y / height) * 2 + 1;
}

function applyExplosion() {
  if (!isWebGLAvailable || !groundPlane) return;
  raycaster.setFromCamera(pointer, camera);
  const intersects = raycaster.intersectObject(groundPlane);
  if (intersects.length === 0) return;

  const explosionPoint = intersects[0].point;
  const explosionRadius = 15;
  const explosionForce = 3.0;

  gophers.forEach((g) => {
    explosionDiff.subVectors(g.pivot.position, explosionPoint);
    explosionDiff.y = 0;
    const dist = explosionDiff.length();
    if (dist < explosionRadius && dist > 0) {
      const strength = (1 - dist / explosionRadius) * explosionForce;
      explosionDiff.normalize().multiplyScalar(strength);
      g.velocity.add(explosionDiff);
    }
  });
}

window.addEventListener('pointermove', updatePointerFromEvent, { passive: true });
window.addEventListener('pointerdown', (event) => {
  updatePointerFromEvent(event);
  if (!prefersReducedMotion) applyExplosion();
});

window.addEventListener('resize', () => {
  const { width, height } = getViewportSize();
  camera.aspect = width / height;
  camera.updateProjectionMatrix();
  if (isWebGLAvailable && renderer) {
    renderer.setSize(width, height);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, MAX_PIXEL_RATIO));
  }
});

window.addEventListener('pageshow', (event) => {
  if (event.persisted) {
    clock.start();
  }
});

const clock = new Clock();
const gopherRadius = 0.65;
const minSafeDist = gopherRadius * 2;
const repulsionStrength = 0.5;

function animate() {
  requestAnimationFrame(animate);

  if (!isWebGLAvailable || isContextLost || !renderer) {
    return;
  }

  const delta = clock.getDelta();
  const timeScale = prefersReducedMotion ? 0 : Math.min(delta * 60, 2);

  raycaster.setFromCamera(pointer, camera);
  const intersects = raycaster.intersectObject(groundPlane);

  let hasTarget = false;
  if (intersects.length > 0) {
    targetPoint.copy(intersects[0].point);
    hasTarget = true;
    gridHelper.material.uniforms.uCursor.value.copy(targetPoint);
  } else {
    gridHelper.material.uniforms.uCursor.value.copy(offscreenCursor);
  }

  if (!prefersReducedMotion) {
    for (let i = 0; i < gophers.length; i++) {
      const g = gophers[i];

      if (hasTarget) {
        personalTarget.copy(targetPoint).add(g.targetOffset);
        diff.subVectors(personalTarget, g.pivot.position);
        diff.y = 0;
        const distToTarget = diff.length();

        if (distToTarget > 0.5) {
          diff.normalize();
          const accel = g.speed * 0.5 * (1 / 3);
          g.velocity.addScaledVector(diff, accel * timeScale);
        } else {
          g.velocity.multiplyScalar(Math.pow(0.85, timeScale));
        }
      }

      g.velocity.multiplyScalar(Math.pow(0.92, timeScale));

      for (let j = 0; j < gophers.length; j++) {
        if (i === j) continue;
        const other = gophers[j];
        diff.subVectors(g.pivot.position, other.pivot.position);
        diff.y = 0;
        const distSq = diff.lengthSq();

        if (distSq < minSafeDist * minSafeDist && distSq > 0.0001) {
          const dist = Math.sqrt(distSq);
          const overlap = minSafeDist - dist;
          kick.copy(diff).normalize().multiplyScalar(repulsionStrength * (0.5 + overlap) * timeScale);
          g.velocity.add(kick);
        }
      }

      g.velocity.clampLength(0, 1.5);
      g.pivot.position.addScaledVector(g.velocity, timeScale);

      const speed = g.velocity.length();
      if (speed > 0.01) {
        motionDir.copy(g.velocity).normalize();
        const targetRotation = Math.atan2(motionDir.x, motionDir.z);
        let rotDiff = targetRotation - g.pivot.rotation.y;
        while (rotDiff > Math.PI) rotDiff -= Math.PI * 2;
        while (rotDiff < -Math.PI) rotDiff += Math.PI * 2;
        g.pivot.rotation.y += rotDiff * 0.2 * timeScale;
        g.visual.rotation.x += (speed * timeScale) / 0.75;
      } else {
        g.visual.rotation.z = Math.sin(Date.now() * 0.005 + g.speed * 100) * 0.1;
      }
    }
  }

  renderer.render(scene, camera);
}

if (isWebGLAvailable) {
  animate();
}
