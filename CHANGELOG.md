# Changelog

## 1.0.0 - 2026-08-16

- Definition과 compiled bundle을 v2로 올리고 v1 호환 경로를 제거했다.
- Color, spacing, radius, size에 `primitive → semantic → component` 단방향 참조를 강제한다.
- Component recipe가 layered foundation의 primitive·semantic·asset token을 직접 소비하면 schema/compiler/`deslint` gate가 실패한다.
- Raw spacing·radius·size 값은 primitive에만 선언하도록 generic fixture와 회귀 테스트를 이관했다.

## 0.2.1 - 2026-08-16

- Go `deslint`가 definition v1의 optional `foundations.icon`을 TypeScript schema와 동일하게 허용하도록 수정했다.
- icon이 없는 기존 definition의 하위 호환과 알 수 없는 foundation 거부를 회귀 테스트로 고정했다.

## 0.2.0 - 2026-08-16

- definition v1에 optional generic icon geometry·usage·action group을 추가했다.
- compiler가 icon reference를 검증하고 canonical theme bundle에 포함하도록 확장했다.
- core runtime에 icon, motion transition, elevation과 interaction geometry/selection helper를 추가했다.

## 0.1.1 - 2026-08-16

- core와 React Native package export에 `require`와 `default` 조건을 추가해 CommonJS-aware test runner와 bundler에서도 동일 ESM build를 해석하도록 했다.
- package 간 내부 dependency와 release manifest를 0.1.1로 맞췄다.

## 0.1.0 - 2026-08-16

- JSON Schema 2020-12 product definition/policy contract와 결정론적 compiler를 추가했다.
- framework-neutral core와 React Native component runtime을 추가했다.
- Tree-sitter 기반 Go `deslint`와 source/Pencil/computed-layout evidence report를 추가했다.
- 제품 migration을 위한 versioned release manifest와 exact checksum 경계를 추가했다.
