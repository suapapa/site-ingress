import './style.css'
import * as THREE from 'three';
import { OBJLoader } from 'three/examples/jsm/loaders/OBJLoader.js';

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

const loader = new OBJLoader();
loader.load(
  '/model/go_gopher_high.obj',
  (object) => {
    // 1. Center and Scale the model
    const box = new THREE.Box3().setFromObject(object);
    const size = box.getSize(new THREE.Vector3());
    const center = box.getCenter(new THREE.Vector3());

    let maxDim = Math.max(size.x, size.y, size.z);

    if (maxDim < 0.001 || !isFinite(maxDim)) {
      maxDim = 40;
    }

    const targetSize = 1.5;
    const scale = targetSize / maxDim;

    // Apply corrected centering
    object.position.copy(center).multiplyScalar(-scale);
    object.scale.set(scale, scale, scale);

    // Create container
    gopherTemplate = new THREE.Group();
    gopherTemplate.add(object);

    // Apply Material - TOON (Cel) Shading
    object.traverse((child) => {
      if (child.isMesh) {
        child.geometry.computeVertexNormals();
        child.material = new THREE.MeshToonMaterial({
          color: 0x00ADD8,
          gradientMap: toonGradient,
        });
        // Shadows disabled
      }
    });

    // Spawn initial gophers
    for (let i = 0; i < 20; i++) {
      spawnGopher();
    }
  },
  (xhr) => { },
  (error) => {
    console.error('Error loading OBJ:', error);
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

      const colors = [
        0x00ADD8, // Gopher Blue
        0xFFD700, // Gold
        0xFF69B4, // Hot Pink
        0x32CD32, // Lime Green
        0xFF4500, // Orange Red
        0x9370DB, // Medium Purple
        0x00FFFF, // Cyan
      ];
      child.material.color.setHex(colors[Math.floor(Math.random() * colors.length)]);
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

// Header
const header = document.createElement('div');
header.className = 'site-header';
header.innerText = "🍀 HOMIN-DEV 🍀";
app.appendChild(header);

const linksContainer = document.createElement('div');
linksContainer.className = 'links-container';
app.appendChild(linksContainer);

// Fish (Movie Line) Container
const fishContainer = document.createElement('div');
fishContainer.className = 'fish-container';
app.appendChild(fishContainer);

// Footer
const footer = document.createElement('div');
footer.className = 'site-footer';
footer.innerHTML = "&copy; Homin Lee &lt;homin.crc@gmail.com&gt; All rights reserved.";
app.appendChild(footer);

// Links Fetcher
const fetchLinks = async () => {
  try {
    const res = await fetch('/api/links');
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
    card.target = "_blank";
    card.innerHTML = `
            <span>${link.Name}</span>
            <span class="link-desc">${link.Desc || ''}</span>
        `;
    linksContainer.appendChild(card);
  });
}

fetchLinks();
fetchFish();

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

// Animation
function animate() {
  requestAnimationFrame(animate);

  // Raycast to find mouse position on ground
  raycaster.setFromCamera(mouse, camera);
  const intersects = raycaster.intersectObject(groundPlane);

  let hasTarget = false;
  if (intersects.length > 0) {
    targetPoint.copy(intersects[0].point);
    hasTarget = true;
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
        const accel = g.speed * 0.5;
        g.velocity.add(diff.multiplyScalar(accel));
      } else {
        // Slowing down at target (Arrival behavior)
        g.velocity.multiplyScalar(0.85);
      }
    }

    // 2. Friction (Global damping)
    g.velocity.multiplyScalar(0.92);

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

        g.velocity.add(kick);
      }
    });

    // Limit Max Velocity
    g.velocity.clampLength(0, 1.5);

    // 4. Update Position
    g.pivot.position.add(g.velocity);

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
      g.pivot.rotation.y += rotDiff * 0.2;

      // Roll (X-axis) based on speed
      g.visual.rotation.x += speed / 0.75;
    } else {
      // Idle Animation
      g.visual.rotation.z = Math.sin(Date.now() * 0.005 + g.speed * 100) * 0.1;
    }
  });

  renderer.render(scene, camera);
}

animate();
