# Site Ingress V2 Migration Plan

## Overview
Transform the current Server-Side Rendered (SSR) Go application into a Client-Side Rendered (CSR) WebGL application. The Go server will serve the static frontend and provide a JSON API for links.

## Plan

### 1. Frontend Development (`/frontend`)
- **Setup**: Initialize a Vite project with Vanilla JavaScript.
- **Tech Stack**:
    -   Vite (Build tool)
    -   Three.js (WebGL Library)
- **Features**:
    -   Load 3D Gopher model from `../asset/go-gopher-model/go_gopher_high.obj`.
    -   Implement "Gophers following mouse" interaction.
    -   Fetch links from `/api/links`.
    -   Display links in a UI overlay (HTML/CSS).
- **Aesthetics**:
    -   Vibrant, modern design.
    -   Glassmorphism for the link overlay.

### 2. Backend Modifications (Go)
- **API Endpoint**:
    -   Create `GET /api/links` to return the list of links in JSON format.
- **Static File Serving**:
    -   Serve the built frontend assets (from `frontend/dist`) at the root path `/`.
    -   Serve 3D assets from `asset/`.
- **Cleanup**:
    -   Remove existing SSR templates (`tmpl/page.tmpl.html`) and associated logic (`page-root.go`).
- **Proxy Configuration**:
    -   Configure Nginx (production) and Vite (dev) to proxy `/api/*` and short links (e.g., `/gh`) to the backend.

### 3. Implementation Steps
1.  [x] Initialize frontend project in `frontend/`.
2.  [x] Install Three.js: `npm install three`.
3.  [x] Develop the 3D scene and logic in `frontend/src/main.js`.
4.  [x] Implement the link fetcher and UI.
5.  [x] Update `main.go` to expose the API and serve static files.
6.  Verify the integration.

## Usage
- **Development**: Run `npm run dev` in `frontend/` for frontend dev. Run `go run .` for backend.
- **Production**: Build frontend (`npm run build`), then run Go server.

## Documentation Maintenance
- **Continuous Updates**: As the project evolves, strictly maintain `GEMINI.md` and `README.md` to reflect the current state, architectural decisions, and usage instructions.
- **Status Tracking**: Update the "Implementation Steps" in `GEMINI.md` to mark tasks as completed (e.g., [x]) as you progress.
