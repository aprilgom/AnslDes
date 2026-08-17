# Impeccable deterministic detector snapshot

이 디렉터리는 Anti-slop Gate TODO의 판정 근거를 skill의 서술이 아니라 실제 결정론적 detector
코드에 고정하기 위한 읽기 전용 upstream snapshot이다.

## Provenance

- upstream: [pbakaus/impeccable](https://github.com/pbakaus/impeccable)
- commit: [`7b646bafd60b9dd9828ce5c4c1a25691702c9e92`](https://github.com/pbakaus/impeccable/commit/7b646bafd60b9dd9828ce5c4c1a25691702c9e92)
- commit date: `2026-08-14T23:15:40Z`
- snapshot date: `2026-08-16`
- deterministic registry: 59 rules
- npm package: `impeccable@3.6.0`
- npm integrity: `sha512-nysc6/2OHTWqLrcSxTxZk4r4QMufhU8NTIuG2ic6p5zzyZe45AWBX3/18OA5S88pCWq+4z8pKsjUxhAM990RKg==`
- license: [Apache-2.0](./upstream/LICENSE)
- source snapshot tree SHA-256: `84d3d3a66be62ea4c361420a6662f42010b123b680e00d8306e6a627ab5f5954`
- repository canonical LF tree SHA-256: `d9d61f4e0239d35c3a00e178fe07e39614508ac041f844be051500fb102732e9`
  (`relative-path + NUL + file-sha256 + LF`, path 정렬)

Upstream 파일은 수정하지 않았다. 생성 산출물인 `detect-antipatterns-browser.js`는 중복된 browser bundle이라
snapshot에서 제외하고, 그 원본 모듈인 `browser/injected/index.mjs`와 각 engine을 보존했다.

## Primary references

- [rule registry](./upstream/cli/engine/registry/antipatterns.mjs): rule ID, category, scope,
  advisory와 engine support
- [package metadata](./upstream/package.json): exact version, Node engine과 parser/browser dependency
- [rule implementation](./upstream/cli/engine/rules/checks.mjs): 실제 판정 조건과 threshold
- [color math](./upstream/cli/engine/shared/color.mjs): 색 파싱, chroma, luminance와 contrast
- [design-system checks](./upstream/cli/engine/design-system.mjs): `DESIGN.md`와 design token drift
- [regex source engine](./upstream/cli/engine/engines/regex/detect-text.mjs): JSX, TSX, CSS source 검사
- [static HTML engine](./upstream/cli/engine/engines/static-html/detect-html.mjs): HTML과 computed CSS cascade
- [browser engine](./upstream/cli/engine/engines/browser/detect-url.mjs): rendered URL 검사
- [visual contrast engine](./upstream/cli/engine/engines/visual/screenshot-contrast.mjs): screenshot 기반 contrast
- [CLI entry](./upstream/cli/engine/cli/main.mjs): 입력, config, finding, exit code
- [CLI config](./upstream/cli/lib/impeccable-config.mjs): ignore와 design-system config 해석
- [skill shim](./upstream/skill/scripts/detect.mjs): skill에서 bundled detector를 호출하는 얇은 진입점

## Usage boundary

이 snapshot은 설계·이식·diff 검토를 위한 기준이며 npm package를 대신하는 설치된 실행 dependency가
아니다. parser dependency가 없는 환경에서 snapshot CLI를 직접 실행하면 regex fallback으로 내려가므로
그 결과를 full HTML/browser gate로 인정하지 않는다. Web gate는 exact package와 dependency를 실행하고,
React Native와 Pencil gate는 registry와 판정 구현의 의미를 각 platform evidence로 옮긴다. Skill 문서는
detector로 결정할 수 없는 LLM-only·native conformance 판단에만 보조적으로 사용한다.

업데이트할 때는 upstream commit, rule count, snapshot hash를 함께 변경하고 registry/rule diff와 migration
note 없이 기존 snapshot을 교체하지 않는다.

Machine registry는 `scripts/anti-slop-registry.mjs`가 vendored 59개 rule과 Hallmark 8개 mapping을 읽어
`deslint/internal/rules/anti_slop_catalog.json`의 canonical 63개를 생성한다. `npm run
check:anti-slop-registry`는 package/version/integrity, 두 tree hash, exact member set, mapping과 pack
fingerprint drift를 모두 실패시킨다. Upstream 갱신은 source snapshot과 generator pin을 함께 바꾸고 생성된
registry 및 migration note diff를 검토한 뒤에만 허용한다.
