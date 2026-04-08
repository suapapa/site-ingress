import './style.css'
import * as THREE from 'three';
import { STLLoader } from 'three/examples/jsm/loaders/STLLoader.js';

// Setup
const scene = new THREE.Scene();
scene.background = new THREE.Color(0x202020); // Dark background
scene.fog = new THREE.Fog(0x202020, 10, 50);

const camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 1000);
camera.position.z = 10;
camera.position.y = 8;
camera.lookAt(0, 0, 0);

const renderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setSize(window.innerWidth, window.innerHeight);
renderer.setPixelRatio(window.devicePixelRatio);
renderer.shadowMap.enabled = false; // Disable shadows
document.body.appendChild(renderer.domElement);

// Lights
const ambientLight = new THREE.AmbientLight(0xffffff, 0.6);
scene.add(ambientLight);

const dirLight = new THREE.DirectionalLight(0xffffff, 1.2);
dirLight.position.set(10, 20, 10);
dirLight.castShadow = false; // Disable shadows
scene.add(dirLight);

// Ground (invisible plane for raycasting)
const planeGeometry = new THREE.PlaneGeometry(100, 100);
const planeMaterial = new THREE.MeshBasicMaterial({ visible: false });
const groundPlane = new THREE.Mesh(planeGeometry, planeMaterial);
groundPlane.rotation.x = -Math.PI / 2;
scene.add(groundPlane);

// Visual Floor: Infinite Grid
const gridSize = 100;
const gridDivisions = 50;
const colorCenterLine = 0x00AAAA; // Cyan center
const colorGrid = 0x353535;      // Subtle dark grey for the rest
const gridHelper = new THREE.GridHelper(gridSize, gridDivisions, colorCenterLine, colorGrid);

// Custom Shader for Glowing Grid Lines
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
    // Glow intensity: 1.0 at center, fading out to 0.0 at uRadius
    float intensity = 1.0 - smoothstep(0.0, uRadius, dist);
    
    // Additive Cyan Glow
    vec3 glowColor = vec3(0.0, 1.0, 1.0);
    // Brighter glow multiplier
    vec3 finalColor = vColor + (glowColor * intensity * 0.8); 
    
    gl_FragColor = vec4(finalColor, 1.0);
  }
`;

const gridMaterial = new THREE.ShaderMaterial({
  vertexShader: gridVertexShader,
  fragmentShader: gridFragmentShader,
  vertexColors: true, // Enables 'color' attribute usage
  uniforms: {
    uCursor: { value: new THREE.Vector3(0, 0, 0) },
    uRadius: { value: 10.0 } // Radius of the glow
  },
  transparent: true,
});

gridHelper.material = gridMaterial;
scene.add(gridHelper);

// Toon Gradient Map (Cel Shading)
function createGradientMap() {
  const format = THREE.RGBAFormat;
  const colors = new Uint8Array([
    150, 150, 150, 255, // Shadow
    200, 200, 200, 255, // Mid
    255, 255, 255, 255, // Highlight
  ]);
  const texture = new THREE.DataTexture(colors, 3, 1, format);
  texture.minFilter = THREE.NearestFilter;
  texture.magFilter = THREE.NearestFilter;
  texture.needsUpdate = true;
  return texture;
}

const toonGradient = createGradientMap();

// Gophers
const gophers = [];
let gopherTemplate = null;

const loader = new STLLoader();
loader.load(
  '/model/go_gopher_low_0.02.stl',
  (geometry) => {
    geometry.computeVertexNormals();

    const mesh = new THREE.Mesh(
      geometry,
      new THREE.MeshToonMaterial({
        color: 0x00ADD8,
        gradientMap: toonGradient,
      })
    );

    geometry.computeBoundingBox();
    const box = geometry.boundingBox;
    const size = new THREE.Vector3();
    const center = new THREE.Vector3();
    box.getSize(size);
    box.getCenter(center);

    let maxDim = Math.max(size.x, size.y, size.z);

    if (maxDim < 0.001 || !isFinite(maxDim)) {
      maxDim = 40;
    }

    const targetSize = 1.5;
    const scale = targetSize / maxDim;

    mesh.position.copy(center).multiplyScalar(-scale);
    mesh.scale.set(scale, scale, scale);

    gopherTemplate = new THREE.Group();
    gopherTemplate.add(mesh);

    // Spawn initial gophers
    for (let i = 0; i < 20; i++) {
      spawnGopher();
    }
  },
  (xhr) => { },
  (error) => {
    console.error('Error loading STL:', error);
  }
);


function spawnGopher() {
  if (!gopherTemplate) return;

  const pivot = new THREE.Group();

  const visual = gopherTemplate.clone();

  // Clone Material for Unique Colors
  visual.traverse((child) => {
    if (child.isMesh) {
      child.material = child.material.clone();
      child.material.gradientMap = toonGradient;

      // Generate vibrant HSL colors
      const hue = Math.random();
      const saturation = 0.8 + Math.random() * 0.2; // 80-100% Saturation for pop
      const lightness = 0.5 + Math.random() * 0.1;  // 50-60% Lightness for depth

      child.material.color.setHSL(hue, saturation, lightness);
    }
  });

  // Lift visual up so it sits on ground. 
  visual.position.y = 0.75;

  pivot.add(visual);

  // Random position on ground
  pivot.position.x = (Math.random() - 0.5) * 40;
  pivot.position.z = (Math.random() - 0.5) * 40;
  pivot.rotation.y = Math.random() * Math.PI * 2;

  scene.add(pivot);

  gophers.push({
    pivot: pivot,
    visual: visual,
    speed: 0.05 + Math.random() * 0.15,
    targetOffset: new THREE.Vector3(
      (Math.random() - 0.5) * 5,
      0,
      (Math.random() - 0.5) * 5
    ),
    velocity: new THREE.Vector3() // Physics velocity
  });
}

// Mouse interaction
const raycaster = new THREE.Raycaster();
const mouse = new THREE.Vector2();
const targetPoint = new THREE.Vector3();

// UI and Links
const app = document.querySelector('#app');
const urlParams = new URLSearchParams(window.location.search);
const showGophersOnly = urlParams.get('show_gophers_only') === 'true';

// Header
const header = document.createElement('div');
header.className = 'site-header';
header.innerText = "🍀 HOMIN-DEV 🍀";
header.style.cursor = 'pointer';
header.addEventListener('click', () => {
  window.location.href = '/';
});
if (!showGophersOnly) app.appendChild(header);

const centerContent = document.createElement('div');
centerContent.className = 'center-content';
if (!showGophersOnly) app.appendChild(centerContent);

const linksContainer = document.createElement('div');
linksContainer.className = 'links-container';
centerContent.appendChild(linksContainer);

// Fish (Movie Line) Container
const fishContainer = document.createElement('div');
fishContainer.className = 'fish-container';
centerContent.appendChild(fishContainer);



// Footer
const footer = document.createElement('div');
footer.className = 'site-footer';
footer.innerHTML = "&copy; Homin Lee &lt;i@homin.dev&gt; All rights reserved.<br>The Go gopher was designed by Renee French.";
app.appendChild(footer);



// Links Fetcher

const fetchLinks = async () => {
  try {
    let url = '/api/links';
    if (window.location.pathname !== '/') {
      const prefix = window.location.pathname.substring(1);
      url = `/api/links?prefix=${prefix}`;
    }

    const res = await fetch(url);
    if (!res.ok) throw new Error('Failed to fetch');
    const data = await res.json();
    renderLinks(data);
  } catch (e) {
    console.warn("Using mock data due to error:", e);
    // Mock data
    const mockData = [
      { Name: "Blog", Link: "https://homin.dev/blog", Desc: "Tech Blog" },
      { Name: "GitHub", Link: "https://github.com/suapapa", Desc: "My Code" },
      { Name: "LinkedIn", Link: "https://linkedin.com/in/homin", Desc: "Profile" }
    ];
    renderLinks(mockData);
  }
};


const fetchFish = async () => {
  try {
    const res = await fetch('/api/fish');
    if (!res.ok) return;
    const data = await res.json();
    // data matches MovieLine struct: { Movie, Line, Character }
    // Case sensitivity? Go JSON uses struct field names. 
    // Typically default gin JSON uses struct field names as is unless tagged.
    // wait, yaml tags were `yaml:"movie"`. 
    // Gin uses `json` package. If no json tags, it uses Field Name (PascalCase).
    // Let's check fish.go tags.
    // `yaml:"movie"` tags are for YAML.
    // So JSON output will be `Movie`, `Line`, `Character` (PascalCase).

    fishContainer.innerHTML = `
            <div class="fish-line">“${data.Line}”</div>
            <div class="fish-char">- ${data.Character} (${data.Movie})</div>
        `;
  } catch (e) {
    console.warn("Fish fetch failed:", e);
    fishContainer.style.display = 'none';
  }
}

function renderLinks(links) {
  linksContainer.innerHTML = '';
  links.forEach(link => {
    const card = document.createElement('a');
    card.href = link.Link;
    card.className = 'link-card';

    card.innerHTML = `
            <span>${link.Desc || link.Name}</span>
            <span class="link-sub">${link.Name}</span>
        `;
    linksContainer.appendChild(card);
  });
}

if (!showGophersOnly) {
  fetchLinks();
  fetchFish();
}

// Event listeners
window.addEventListener('mousemove', (event) => {
  mouse.x = (event.clientX / window.innerWidth) * 2 - 1;
  mouse.y = -(event.clientY / window.innerHeight) * 2 + 1;
});

window.addEventListener('resize', () => {
  camera.aspect = window.innerWidth / window.innerHeight;
  camera.updateProjectionMatrix();
  renderer.setSize(window.innerWidth, window.innerHeight);
});

window.addEventListener('mousedown', () => {
  raycaster.setFromCamera(mouse, camera);
  const intersects = raycaster.intersectObject(groundPlane);

  if (intersects.length > 0) {
    const explosionPoint = intersects[0].point;
    const explosionRadius = 15;
    const explosionForce = 3.0; // Strong enough to hit the speed cap instantly

    gophers.forEach(g => {
      const diff = new THREE.Vector3().subVectors(g.pivot.position, explosionPoint);
      diff.y = 0;
      const dist = diff.length();

      if (dist < explosionRadius) {
        // Force decreases with distance
        const strength = (1 - dist / explosionRadius) * explosionForce;
        diff.normalize();
        g.velocity.add(diff.multiplyScalar(strength));
      }
    });
  }
});

// Animation
const clock = new THREE.Clock();

function animate() {
  requestAnimationFrame(animate);

  const delta = clock.getDelta();
  const timeScale = Math.min(delta * 60, 2); // Normalize to 60fps, cap at 2x lag

  // Raycast to find mouse position on ground
  raycaster.setFromCamera(mouse, camera);
  const intersects = raycaster.intersectObject(groundPlane);

  let hasTarget = false;
  if (intersects.length > 0) {
    targetPoint.copy(intersects[0].point);
    hasTarget = true;

    // Update Grid Glow Position
    gridHelper.material.uniforms.uCursor.value.copy(targetPoint);
  } else {
    // Move glow off-screen if no intersection
    gridHelper.material.uniforms.uCursor.value.set(1000, 1000, 1000);
  }

  const gopherRadius = 0.65;
  const minSafeDist = gopherRadius * 2;
  const repulsionStrength = 0.5; // Impulse strength for "bounce"

  gophers.forEach(g => {
    // 1. Goal: Seek Target
    if (hasTarget) {
      const personalTarget = targetPoint.clone().add(g.targetOffset);
      const diff = new THREE.Vector3().subVectors(personalTarget, g.pivot.position);
      diff.y = 0;

      const distToTarget = diff.length();

      // Acceleration towards target
      if (distToTarget > 0.5) {
        diff.normalize();

        // Use individual speed stat as acceleration factor
        const accel = g.speed * 0.5 * (1 / 3); // Reduced to 1/3 speed
        g.velocity.add(diff.multiplyScalar(accel * timeScale));
      } else {
        // Slowing down at target (Arrival behavior)
        // Lerp-like friction: v = v * pow(0.85, timeScale)
        g.velocity.multiplyScalar(Math.pow(0.85, timeScale));
      }
    }

    // 2. Friction (Global damping)
    g.velocity.multiplyScalar(Math.pow(0.92, timeScale));

    // 3. Collision Resolution (Bounce)
    // Check against all others
    gophers.forEach(other => {
      if (g === other) return;

      const diff = new THREE.Vector3().subVectors(g.pivot.position, other.pivot.position);
      diff.y = 0;
      const distSq = diff.lengthSq();

      // If overlapping
      if (distSq < minSafeDist * minSafeDist && distSq > 0.0001) {
        const dist = Math.sqrt(distSq);
        const overlap = minSafeDist - dist;

        // Calculate Bounce Force
        // Force vector pushing AWAY from the other gopher
        const normal = diff.normalize();

        // Impulse: proportional to overlap plus base kick
        // This gives them a "velocity kick" outwards
        const kick = normal.multiplyScalar(repulsionStrength * (0.5 + overlap));

        g.velocity.add(kick.multiplyScalar(timeScale));
      }
    });

    // Limit Max Velocity
    g.velocity.clampLength(0, 1.5);

    // 4. Update Position
    g.pivot.position.add(g.velocity.clone().multiplyScalar(timeScale));

    // 5. Update Visual Rotation
    const speed = g.velocity.length();
    if (speed > 0.01) {
      // Face velocity direction
      const motionDir = g.velocity.clone().normalize();
      const targetRotation = Math.atan2(motionDir.x, motionDir.z);

      // Smooth rotation
      let rotDiff = targetRotation - g.pivot.rotation.y;
      while (rotDiff > Math.PI) rotDiff -= Math.PI * 2;
      while (rotDiff < -Math.PI) rotDiff += Math.PI * 2;
      g.pivot.rotation.y += rotDiff * 0.2 * timeScale;

      // Roll (X-axis) based on speed
      g.visual.rotation.x += (speed * timeScale) / 0.75;
    } else {
      // Idle Animation
      g.visual.rotation.z = Math.sin(Date.now() * 0.005 + g.speed * 100) * 0.1;
    }
  });

  renderer.render(scene, camera);
}

animate();
