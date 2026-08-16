# Repository Guidelines

## Purpose

`deslint`는 제품 중립적인 독립 결정론적 linter다. 엔진은 제품 UI를 렌더링하거나 수정하지 않으며
주입된 source, contract, Pencil JSON과 computed-layout evidence를 분석해 diagnostic을 만든다.

## Architecture boundaries

- `cmd/deslint`: CLI wiring만 둔다.
- `internal/source`: TS/TSX parsing, lexical binding과 normalized source IR.
- `internal/rules`: platform-neutral rule evaluation.
- `internal/contract`: language-neutral contract와 schema validation.
- `internal/pen`: `.pen` document model과 Pencil rules.
- `internal/layout`: pen.dev computed-layout report와 overflow/overlap 분석.
- `internal/report`: stable JSON, text와 SARIF output.
- `testdata`: positive, negative와 false-positive golden fixture.

엔진에 소비 저장소의 이름, 절대 경로나 특정 사용처 개수를 하드코딩하지 않는다. 제품 고유 owner,
consumer, 예외와 budget은 versioned policy/config로 주입한다. 제품 snapshot과 parity report는 소비
저장소에만 둔다.

## Source analysis

- TS/TSX syntax는 pinned `tree-sitter`와 `tree-sitter-typescript`를 사용한다.
- parser CST를 그대로 rule에 노출하지 말고 normalized IR로 변환한다.
- import identity, alias와 shadowing은 lexical scope와 module resolver로 증명한다.
- type-aware 판단이 필요하면 `SemanticProvider` 경계 뒤에 TypeScript 7 IPC adapter를 둔다.
- TypeScript compiler의 `internal` Go package나 esbuild internal parser에 직접 의존하지 않는다.
- regex만으로 AST identity를 통과시키지 않는다. 보조 문자열 검사는 AST owner가 확정된 뒤에만 허용한다.

## Deterministic rule requirements

모든 rule은 다음을 가져야 한다.

1. 안정적인 rule ID
2. 명시적인 input/evidence kind
3. positive fixture
4. negative fixture
5. 우회·shadow·spread fixture
6. false-positive boundary
7. stable source location
8. 문서화된 severity와 exception policy

Broad ignore, 진단 후처리 삭제, 실행 오류의 pass 변환을 금지한다. 미실행과 미확보 runtime evidence는
`not-run` 또는 `deferred`이며 pass가 아니다.

## Build and test

Go 1.26 이상을 사용한다. Tree-sitter parser build는 CGO가 필요하다.

```bash
go fmt ./...
go vet ./...
go test ./...
go test -race ./...
```

변경 후 최소한 `go test ./...`와 `go vet ./...`를 실행한다. parser나 rule 변경은 관련 중립 golden
fixture를 함께 실행한다.

## Coding style

- `gofmt` 결과를 그대로 사용한다.
- panic으로 사용자 입력 오류를 처리하지 않는다.
- filesystem, process와 clock은 interface 뒤에 두어 fixture가 결정론적이어야 한다.
- diagnostic 정렬 순서를 명시해 OS와 map iteration에 따라 report가 바뀌지 않게 한다.
- exported API에는 Go doc을 작성한다.
- 새 dependency는 공식 source, 라이선스, 유지 상태와 pin 이유를 기록한다.

## Git and releases

- 기본 통합 브랜치는 `develop`, release 브랜치는 `main`이다.
- Conventional Commits와 한국어 설명을 사용한다.
- release는 macOS, Linux, Windows binary와 SHA-256을 제공한다.
- 소비 저장소는 release version과 checksum을 고정하고 floating latest를 사용하지 않는다.
- 기존 linter 제거와 rollback 판단은 각 소비 저장소가 자체 parity report와 승인 절차로 수행한다.
