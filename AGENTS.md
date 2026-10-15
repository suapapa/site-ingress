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
    -   Load 3D Gopher model from `../asset/go_gopher_low_0.02.stl` using Three.js `STLLoader` (served at `/model/go_gopher_low_0.02.stl` via `frontend/public/model/go_gopher_low_0.02.stl` → `asset/go_gopher_low_0.02.stl`).
    -   Implement "Gophers following mouse" interaction.
    -   Fetch links from `/api/links`.
    -   Display links in a UI overlay (HTML/CSS).
    -   Support `?show_gophers_only=true` to hide UI overlay.
- **Aesthetics**:
    -   Dark restrained chrome over the WebGL gopher scene (Go cyan accent).
    -   Quiet typographic overlay (no glassmorphism); see `DESIGN.md`.

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
7.  [x] Optimize animation speed for high refresh rate displays (Delta Time implementation).
8.  [x] Impeccable audit remediation (a11y, perf, quieter UI, DESIGN.md).

## Usage
- **Development**: Run `npm run dev` in `frontend/` for frontend dev. Run `go run .` for backend.
- **Production**: Build frontend (`npm run build`), then run Go server.

## Documentation Maintenance
- **Continuous Updates**: As the project evolves, strictly maintain `GEMINI.md` and `README.md` to reflect the current state, architectural decisions, and usage instructions.
- **Status Tracking**: Update the "Implementation Steps" in `GEMINI.md` to mark tasks as completed (e.g., [x]) as you progress.

---

## Server-side rendering of the SPA routes (크롤러 대응)

### 왜 필요한가
프론트엔드는 WebGL 클라이언트 렌더링이라, JS 를 실행하지 않는 크롤러에게는
`<div id="app"></div>` 만 있는 **빈 문서**로 보입니다. 구글은 JS 를 렌더하지만
**네이버 Yeti 는 그렇지 않아** 국내 검색에서 손해를 봅니다. (2026-10 SEO 감사 결과)

### 어떻게 동작하나
1. 백엔드가 기동할 때 `static/index.html`(Vite 빌드 shell)을 읽어 **라우트별 HTML 을 미리 렌더**합니다.
   설정을 기동 시 한 번만 읽는 기존 구조와 동일하므로 캐시 무효화 문제가 없습니다.
2. 렌더 단위는 `links.yaml` 의 **prefix 마다** 하나입니다 (배포 설정 기준 `/`, `/social`, `/archive`, `/support`).
   - `#app` 안에 `buildChrome()` 과 동일한 구조(header / 링크 카드 / footer)를 넣습니다.
   - `hide: true` 링크 제외, HTML 이스케이프, 라우트별 `<title>`·description·canonical·`og:url`, JSON-LD(WebSite+Person).
   - 문구는 하드코딩하지 않고 **설정에서 파생**합니다 (`HOMIN-DEV — Social` / 링크 라벨 나열).
     `/` 의 title·description 은 `index.html` 의 손으로 쓴 값을 그대로 둡니다.
3. `frontend/src/main.js` 는 하이드레이션 시 `app.replaceChildren()` 로 프리렌더 내용을 **교체**합니다
   — `appendChild` 로 바뀌면 링크가 두 번 그려집니다.
4. nginx 는 해당 라우트만 백엔드로 프록시하고, **백엔드 장애 시 정적 shell 로 폴백**합니다
   (`error_page 404 500 502 503 504 = @shell`) → 페이지는 예전처럼 클라이언트 렌더링으로 동작합니다.

### 관련 파일
| 파일 | 역할 |
|---|---|
| `render.go` | 프리렌더 본체 (`prerenderPages`, `setAppContent`, `addStructuredData`) |
| `main.go` | `-static` 플래그, `loadPrerenderShell()`, `rootHandler`/`pathHandler`/`servePrerendered` |
| `frontend/nginx.conf` | SPA 라우트 → 백엔드 프록시 + `@shell` 폴백 |
| `frontend/src/main.js` | 하이드레이션 시 `replaceChildren()` |
| `Dockerfile` | `node` 스테이지에서 프론트를 빌드해 `dist/index.html` 만 백엔드 이미지에 `static/` 로 복사 |

### 로컬에서 확인
```sh
cd frontend && npm ci && npm run build && cd ..
go run . -static frontend/dist -c <배포중인 ingress_links.yaml>
curl -s localhost:8080/social | grep -E '<title>|link-card'
go test ./...     # render_test.go: 라우트별 title/canonical, 이스케이프, ?show_gophers_only, 404/단축링크
```

브라우저까지 확인하려면 (WebGL 이라 headless 에서는 swiftshader 필요):
```sh
google-chrome --headless=new --disable-gpu --enable-unsafe-swiftshader --no-sandbox \
  --user-data-dir=/tmp/chrome-prof --virtual-time-budget=8000 --dump-dom \
  http://localhost:8080/social
# 카드 수가 curl 결과와 "같아야" 합니다 (2배면 replaceChildren 이 빠진 것)
```

### 주의
- 설정(`links.yaml`)을 바꾸면 **백엔드 재시작**이 필요합니다 (프리렌더·API 모두 기동 시 로드).
- 배포는 `./push2vultr.sh -b -r` — **백엔드 이미지도 반드시 재빌드**해야 합니다
  (프리렌더 템플릿이 백엔드 이미지에 들어갑니다). 프론트만 올리면 nginx 폴백만 동작합니다.
- nginx 정규식에 없는 **새 prefix** 도 `@backend` 경유로 동작하지만 정적 폴백은 받지 못합니다.
  새 prefix 를 추가하면 `frontend/nginx.conf` 의 정규식도 갱신하세요.
- `tidyPath("/")` 가 `""` 를 반환하던 버그를 이 작업에서 함께 수정했습니다 (루트 prefix 가 사라져 `/` 가 404 였음).
- `docker-compose.yml` 은 컨테이너가 분리돼 있어 nginx→`localhost:8080` 이 닿지 않습니다
  (k8s 는 같은 팟 → 네트워크 네임스페이스 공유). 로컬에서 팟처럼 재현하려면
  `docker run -d --name fe --network container:<backend-container> <frontend-image>` 를 쓰세요.

