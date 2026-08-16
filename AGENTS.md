# Repository Guidelines

## Purpose

AnslDes는 제품 중립 디자인 시스템 toolkit이다. 특정 제품의 token 값이나 source 경로를 이 저장소의
runtime, schema 또는 rule에 하드코딩하지 않는다.

## Ownership rules

- `packages/schema`: 언어 중립 definition schema만 소유한다.
- `packages/compiler`: reference, type, cycle과 canonicalization을 소유한다.
- `packages/react-native`: contract-driven 공용 component anatomy를 소유한다.
- `deslint`: 범용 검사 엔진을 소유한다.
- 제품 값, 브랜드 asset, 화면 flow, UX writing, owner/consumer 경로와 예외 registry는 제품 저장소가 소유한다.
- 제품 구현, snapshot, migration report와 rollback evidence를 AnslDes 안에 복사하지 않는다.

Generic package, 문서와 test fixture에 제품 이름, 제품 source 경로, 실제 브랜드 색상이나 고정 consumer
count를 추가하지 않는다. 예시는 `example-product`처럼 중립 fixture를 사용한다. 제품별 parity는 소비
저장소가 AnslDes release checksum과 함께 소유한다.

## Contract requirements

- canonical interchange format은 JSON이다.
- schema는 JSON Schema 2020-12를 사용한다.
- reference resolution 결과와 diagnostic ordering은 결정론적이어야 한다.
- missing reference, type mismatch, cycle, unknown key와 duplicate identity는 실패다.
- schema version과 product definition version은 독립적으로 관리한다.

## Checks

```bash
npm run check
```

변경 범위에 따라 schema fixture와 `deslint`의 Go unit/race test를 추가한다.

## Git

- 기본 통합 브랜치는 `develop`, release 브랜치는 `main`이다.
- Conventional Commits와 한국어 설명을 사용한다.
- 제품 migration과 generic library release를 한 커밋에 섞지 않는다.
