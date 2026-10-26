# Ingress site of Homin.dev

Live > [HERE](https://homin.dev) <

![v2](img/v2.png)

## Notes

- **3D gopher**: Mesh is `asset/go_gopher_low_0.02.bin` (indexed 16-bit quantized binary converted from `asset/go_gopher_low_0.02.stl` via `./scripts/stl2qbin.py`), loaded in the browser at `/model/go_gopher_low_0.02.bin` (Vite `public/model/go_gopher_low_0.02.bin` is a symlink to that file in dev).
- **UI**: Dark typographic chrome over the WebGL scene (Bricolage Grotesque, Go cyan accent). Design tokens live in `DESIGN.md` / `frontend/src/style.css`.
- **Dev**: `npm run dev` in `frontend/`, `go run .` for the API. Or run `make help` for all commands.
- **Gophers only**: `?show_gophers_only=true` hides the link overlay.
- **Crawlers / SSR**: SPA 라우트는 백엔드가 서버 렌더링합니다 — JS 를 실행하지 않는 크롤러(특히 네이버 Yeti)는 빈 `#app` 만 보기 때문입니다. 구조·검증법·주의사항은 `AGENTS.md` 참고.
- **LLMs / AI Scrapers**: `/llms.txt` (`/llms-full.txt`), 라우트별 `.md` (`/index.md`, `/social.md` 등), 그리고 `Accept: text/markdown` 콘텐츠 협상을 통해 전체 공개 링크와 단축 링크 인덱스를 마크다운으로 바로 제공합니다.

## Makefile Commands

자주 사용하는 명령어는 `Makefile`로 정의되어 있습니다. `make` 또는 `make help`를 입력하면 사용 가능한 모든 명령어와 변수를 확인할 수 있습니다.

- **개발 및 테스트**:
  - `make dev-fe`: 프론트엔드 Vite 개발 서버 실행
  - `make dev-be`: 프론트엔드 빌드 결과와 설정 파일을 연동하여 백엔드 실행
  - `make run`: 기본 백엔드 서버 실행
  - `make test`: Go 테스트 실행 (`go test -v ./...`)
- **빌드**:
  - `make build`: 프론트엔드 및 백엔드 바이너리 전체 빌드
  - `make build-fe`: 프론트엔드 에셋 빌드
  - `make build-be`: 백엔드 Go 바이너리 빌드
  - `make model`: STL 모델을 최적화된 양자화 바이너리로 변환
- **배포 (Vultr / K8s)**:
  - `make deploy`: 백엔드/프론트엔드 이미지 빌드 & 푸시 후 k8s 롤아웃 재시작
  - `make deploy-be`: 백엔드 이미지 빌드 & 푸시 후 재시작
  - `make deploy-fe`: 프론트엔드 이미지 빌드 & 푸시 후 재시작
  - `make restart`: k8s 배포 롤아웃 재시작
  - `make status`: k8s 롤아웃 상태 및 파드 확인