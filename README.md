# Ingress site of Homin.dev

Live > [HERE](https://homin.dev) <

![v2](img/v2.png)

## Notes

- **3D gopher**: Mesh is `asset/go_gopher_low_0.02.stl`, loaded in the browser with Three.js `STLLoader` at `/model/go_gopher_low_0.02.stl` (Vite `public/model/go_gopher_low_0.02.stl` is a symlink to that file in dev).
- **UI**: Dark typographic chrome over the WebGL scene (Bricolage Grotesque, Go cyan accent). Design tokens live in `DESIGN.md` / `frontend/src/style.css`.
- **Dev**: `npm run dev` in `frontend/`, `go run .` for the API. Production: `npm run build` then run the Go server.
- **Gophers only**: `?show_gophers_only=true` hides the link overlay.
- **Crawlers / SSR**: JS 를 실행하지 않는 크롤러(특히 네이버 Yeti)에게 이 앱은 빈 `<div id="app">` 로 보입니다. 그래서 Go 백엔드가 기동 시 링크 prefix 마다 HTML 을 미리 렌더해서 `#app` 안에 넣어 줍니다 (`render.go`) — API 와 같은 설정을 쓰므로 브라우저와 크롤러가 보는 내용이 갈라지지 않습니다. `frontend/src/main.js` 는 하이드레이션 시 `replaceChildren()` 으로 그 내용을 교체하고, nginx 는 백엔드 장애 시 정적 shell 로 폴백합니다.
  - 로컬 확인: `cd frontend && npm run build && cd .. && go run . -static frontend/dist`
  - 검증: `go test ./...` (`render_test.go` — 라우트별 title/canonical, 이스케이프, `?show_gophers_only`, 404/단축링크)
  - 설정을 바꾸면 백엔드 재시작이 필요합니다(프리렌더·API 모두 기동 시 로드).