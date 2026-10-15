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
} from 'three';
import { STLLoader } from 'three/examples/jsm/loaders/STLLoader.js';

const SCENE_BG = 0x1a1a1a;
const MAX_PIXEL_RATIO = 2;
const GOPHER_COUNT = 20;

const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

const scene = new Scene();
scene.background = new Color(SCENE_BG);
scene.fog = new Fog(SCENE_BG, 10, 50);

const camera = new PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 1000);
camera.position.z = 10;
camera.position.y = 8;
camera.lookAt(0, 0, 0);

const renderer = new WebGLRenderer({ antialias: !prefersReducedMotion, powerPreference: 'high-performance' });
renderer.setSize(window.innerWidth, window.innerHeight);
renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, MAX_PIXEL_RATIO));
renderer.shadowMap.enabled = false;
renderer.domElement.setAttribute('aria-hidden', 'true');
document.body.appendChild(renderer.domElement);

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

const loader = new STLLoader();
loader.load(
  '/model/go_gopher_low_0.02.stl',
  (geometry) => {
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
  },
  undefined,
  (error) => {
    console.error('Error loading STL:', error);
  }
);

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

const app = document.querySelector('#app');
const urlParams = new URLSearchParams(window.location.search);
const showGophersOnly = urlParams.get('show_gophers_only') === 'true';

const sceneDescription = document.createElement('p');
sceneDescription.className = 'visually-hidden';
sceneDescription.id = 'scene-desc';
sceneDescription.textContent =
  'Interactive 3D scene: colorful Go gophers follow the pointer across a dark grid. Links and site chrome appear above the scene.';
document.body.insertBefore(sceneDescription, app);

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

let linksContainer;
let fishContainer;
let statusEl;

// The backend prerenders this chrome into #app for crawlers that do not run JavaScript
// (see render.go). Replace it with the live DOM instead of appending to it, or the page
// would show every destination twice.
app.replaceChildren();

if (!showGophersOnly) {
  ({ linksContainer, fishContainer, statusEl } = buildChrome());
} else {
  const footer = document.createElement('footer');
  footer.className = 'site-footer';
  footer.innerHTML =
    '&copy; Homin Lee &lt;i@homin.dev&gt; All rights reserved.<br>The Go gopher was designed by Renee French.';
  app.appendChild(footer);
}

const mockLinks = [
  { Name: 'Blog', Link: 'https://homin.dev/blog', Desc: 'Tech Blog' },
  { Name: 'GitHub', Link: 'https://github.com/suapapa', Desc: 'My Code' },
  { Name: 'LinkedIn', Link: 'https://linkedin.com/in/homin', Desc: 'Profile' },
];

async function fetchLinks() {
  statusEl.hidden = false;
  statusEl.textContent = 'Loading links…';

  try {
    let url = '/api/links';
    if (window.location.pathname !== '/') {
      const prefix = window.location.pathname.substring(1);
      url = `/api/links?prefix=${encodeURIComponent(prefix)}`;
    }

    const res = await fetch(url);
    if (!res.ok) throw new Error('Failed to fetch');
    const data = await res.json();
    renderLinks(data);
  } catch (e) {
    console.warn('Using mock data due to error:', e);
    statusEl.textContent = 'Showing offline links (API unavailable).';
    renderLinks(mockLinks);
  }
}

async function fetchFish() {
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
  linksContainer.replaceChildren();
  statusEl.hidden = true;

  if (!Array.isArray(links) || links.length === 0) {
    statusEl.hidden = false;
    statusEl.textContent = 'No links available.';
    return;
  }

  links.forEach((link) => {
    const card = document.createElement('a');
    card.href = link.Link;
    card.className = 'link-card';

    const label = document.createElement('span');
    label.textContent = link.Desc || link.Name || 'Link';

    const sub = document.createElement('span');
    sub.className = 'link-sub';
    sub.textContent = link.Name || '';

    card.append(label, sub);
    linksContainer.appendChild(card);
  });
}

if (!showGophersOnly) {
  fetchLinks();
  fetchFish();
}

function updatePointerFromEvent(event) {
  const x = event.clientX ?? (event.touches && event.touches[0]?.clientX);
  const y = event.clientY ?? (event.touches && event.touches[0]?.clientY);
  if (x == null || y == null) return;
  pointer.x = (x / window.innerWidth) * 2 - 1;
  pointer.y = -(y / window.innerHeight) * 2 + 1;
}

function applyExplosion() {
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
  camera.aspect = window.innerWidth / window.innerHeight;
  camera.updateProjectionMatrix();
  renderer.setSize(window.innerWidth, window.innerHeight);
  renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, MAX_PIXEL_RATIO));
});

const clock = new Clock();
const gopherRadius = 0.65;
const minSafeDist = gopherRadius * 2;
const repulsionStrength = 0.5;

function animate() {
  requestAnimationFrame(animate);

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

animate();
