# Ingress site of Homin.dev

Live > [HERE](https://homin.dev) <

![v2](img/v2.png)

## Notes

- **3D gopher**: Mesh is `asset/go_gopher_low_0.02.stl`, loaded in the browser with Three.js `STLLoader` at `/model/go_gopher_low_0.02.stl` (Vite `public/model/go_gopher_low_0.02.stl` is a symlink to that file in dev).
- **UI**: Dark typographic chrome over the WebGL scene (Bricolage Grotesque, Go cyan accent). Design tokens live in `DESIGN.md` / `frontend/src/style.css`.
- **Dev**: `npm run dev` in `frontend/`, `go run .` for the API. Production: `npm run build` then run the Go server.
- **Gophers only**: `?show_gophers_only=true` hides the link overlay.