# AnslDes

AnslDes는 제품별 브랜드 값과 독립된 범용 디자인 시스템 toolkit이다.

AnslDes는 token의 실제 값을 소유하지 않는다. color, typography, spacing, radius, size, motion과
component recipe의 값은 각 제품 저장소가 AnslDes schema에 맞춰 정의한다.

```text
AnslDes                          product repository
├─ schema      ◀─────────────── product.design-system.json
├─ compiler                      product assets and exceptions
├─ runtimes                      product source policy
└─ deslint     ───────────────▶ source / Pencil / layout evidence
```

## Workspaces

- [`packages/schema`](./packages/schema/README.md): 제품 정의용 언어 중립 JSON Schema
- [`deslint`](./deslint/README.md): 디자인 시스템 결정론적 Go linter
- [`packages/compiler`](./packages/compiler/README.md): reference resolution과 canonical output compiler
- [`packages/core`](./packages/core/README.md): framework-neutral theme/token/recipe runtime
- [`packages/react-native`](./packages/react-native/README.md): contract-driven React Native component runtime

## Ownership boundary

AnslDes가 소유하는 것:

- token과 component recipe의 schema
- reference resolution과 cycle/type 검증
- 공용 component anatomy와 runtime boundary
- 제품 중립 lint rule과 report format

제품 저장소가 소유하는 것:

- 실제 color, typography, radius와 spacing 값
- light/dark 같은 theme mapping
- 제품별 component variant와 recipe
- 브랜드 asset, 화면 flow, UX writing과 예외 정책
- 제품 source owner와 consumer 경로

자세한 경계는 [Architecture](./docs/architecture.md)와
[Product boundary](./docs/product-boundary.md)를 참고한다.

## Anti-slop roadmap

Impeccable 2026의 결정론적 detector를 제품 중립 `deslint` evidence·rule provider로 수용하는 계획은
[Anti-slop Gate TODO](./docs/anti-slop/TODO.md)에서 관리한다. AnslDes는 generic registry와 provider
계약만 소유하며 실제 제품 화면의 baseline, owner, exception과 통과 report는 소비 저장소에 남긴다.

## Commands

```bash
npm install
npm run check
```

`npm run check`는 generic schema fixture, 제품 결합 방지 검사와 `deslint` Go test/vet을 실행한다.

릴리스 후보는 `npm run release:check`로
[`release/ansldes-release.json`](./release/ansldes-release.json)의 package version과 source artifact
checksum을 검증한다. tag와 제품 lock 절차는 [Release process](./docs/release-process.md)를 따른다.

## Status

현재 schema, compiler, framework-neutral runtime, React Native 기본 anatomy와 linter bootstrap이 있다.
제품은 기존 디자인 시스템 구현을 제거하기 전에 versioned AnslDes release를 exact pin하고 dual-run
migration과 rollback evidence를 유지해야 한다. 첫 정식 release는 `v0.1.0`이다.

- [Repository TODO](./TODO.md)
- [Anti-slop Gate TODO](./docs/anti-slop/TODO.md)
- [Repository rules](./AGENTS.md)
