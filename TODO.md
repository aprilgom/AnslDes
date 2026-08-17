# AnslDes TODO

> [Repository rules](./AGENTS.md) · [Optional non-deterministic work](./TODO_Optional.md)

## 1. Product-neutral schema

- [x] 제품 정의와 generic toolkit의 소유권 경계를 문서화한다.
- [x] JSON Schema 2020-12 기반 definition v2를 추가한다.
- [x] generic example과 product-coupling negative 검사를 추가한다.
- [x] token reference existence와 layer type을 schema/compiler 조합으로 검증한다.
- [x] theme name과 semantic mapping의 exact-set 일치를 검증한다.
- [x] color·spacing·radius·size의 `primitive → semantic → component` 단방향 참조를 강제한다.
- [x] component recipe의 primitive·semantic 직접 소비와 non-primitive raw token을 거부한다.

## 2. Compiler

- [x] reference resolver와 cycle detection을 구현한다.
- [x] canonical JSON serialization과 SHA-256 fingerprint를 구현한다.
- [x] JSON definition schema에서 TypeScript와 Go model을 생성한다.
- [x] resolved light/dark token bundle을 생성한다.
- [x] compiler 오류가 partial output을 남기지 않게 한다.

## 3. Runtime packages

- [x] framework-neutral recipe resolver를 구현한다.
- [x] React Native theme adapter를 구현한다.
- [x] Button, TextField, selection, list와 feedback anatomy를 제품 값 없이 구현한다.
- [x] React/React Native를 peer dependency로 고정한다.
- [x] 100%·160%·235% typography scaling fixture를 제공한다.
- [x] generic icon geometry·usage·action runtime을 제공한다.
- [x] motion/elevation recipe와 interaction helper를 framework-neutral core에서 제공한다.

## 4. deslint integration

- [x] definition/schema version 검사를 추가한다.
- [x] raw value, unknown token과 invalid reference rule을 추가한다.
- [x] product policy를 별도 input으로 받는다.
- [x] source/Pencil/computed-layout evidence를 분리한다.
- [x] text, JSON과 SARIF report를 구현한다.

## 5. Product migration

- [x] 제품 저장소에서 기존 token 값을 product definition으로 export한다.
- [x] 기존 TypeScript 계약과 JSON definition의 dual-run parity를 검증한다.
- [x] release candidate의 package version과 source checksum manifest를 결정론적으로 검증한다.
- [x] consumer release lock schema로 제품 adapter가 versioned AnslDes release를 exact pin하게 한다.
- [x] 제품 source owner, consumer와 예외를 product policy에만 두는 경계를 강제한다.
- [x] parity 승인과 rollback release 없이 기존 구현을 제거하지 않는 절차를 문서화한다.

## 6. Anti-slop gate

- [x] [Impeccable 2026과 Hallmark 기반 제품 중립 Anti-slop Gate](./docs/anti-slop/TODO.md)를 구현한다.
- [x] Impeccable 59개와 Hallmark eight tells의 중복 제거 결과인 63개 deterministic rule의
      canonical registry와 evidence 경계를 고정한다.
- [x] rule 구현을 공통 `RuleSpec` 계약과 versioned rule pack으로 등록해 engine core 수정 없이
      rule을 추가·교체할 수 있게 한다.
- [x] consumer policy가 rule pack version·fingerprint와 rule activation을 exact하게 선언하게 한다.
- [x] Web, React Native와 design-document provider 결과를 서로 다른 evidence kind로 유지한다.
- [x] 제품 profile, owner, consumer, exception과 runtime report를 consumer policy에만 둔다.

## Completion

- [x] AnslDes generic source에 제품 이름·경로·브랜드 값이 없다.
- [x] 두 개 이상의 중립 fixture가 같은 schema/compiler를 소비한다.
- [x] 제품 값 변경이 AnslDes release 없이 제품 definition version만 변경한다.
- [x] schema breaking change만 AnslDes major version을 요구한다.
- [x] rule 추가와 pack 구성 변경이 engine core의 조건문 수정 없이 registry·manifest·fixture 변경으로 끝난다.
