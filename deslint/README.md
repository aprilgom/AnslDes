# deslint

`deslint`는 React Native 소스, 디자인 시스템 계약, Pencil 문서와 computed-layout evidence를 하나의
결정론적 규칙 체계로 검사하기 위한 독립 Go linter다.

현재 Tree-sitter 기반 TypeScript/TSX parser, definition/policy contract, v1 core rule engine과 세 report
format이 동작한다. 제품별 규칙과 source owner는 versioned policy로 외부에서 주입한다.

> **상태: bootstrap**
> 기존 product linter 제거와 cutover는 소비 저장소가 전체 fixture, diagnostic parity와 rollback 경계를
> 확인한 뒤 자체 승인 절차로 진행한다.

## 목표

- 같은 입력은 OS와 실행 순서에 관계없이 같은 diagnostic과 report를 만든다.
- TS/TSX의 import identity, alias, lexical shadowing과 JSX spread를 AST 기반으로 판별한다.
- color, typography, layout, motion, icon, component와 accessibility 계약을 언어 중립 데이터로 관리한다.
- Pencil document와 pen.dev computed-layout 결과를 source lint와 동일한 rule model로 검사한다.
- text, JSON과 SARIF report를 제공하고 error·warning·raw token·overflow·overlap budget을 강제한다.
- 기존 Node/TypeScript linter와 dual-run하여 rule ID, 위치, severity와 finding 수의 parity를 증명한다.

## 현재 구현 범위

| 영역                            | 상태                                                       |
| ------------------------------- | ---------------------------------------------------------- |
| Go CLI                          | `version`, `parse`, `lint` 명령 사용 가능                  |
| TypeScript/TSX parser           | Tree-sitter 기반 syntax summary 제공                       |
| Syntax error evidence           | `hasError`로 보존                                          |
| Normalized source IR            | object/JSX property literal과 source range 제공            |
| Import identity와 lexical scope | 예정                                                       |
| Design-system rule engine       | schema version, reference, raw value와 budget v1 제공      |
| Pencil/computed-layout lint     | raw property와 overflow/overlap/clipping/stale SHA v1 제공 |
| JSON/SARIF report               | text, native JSON, SARIF 2.1.0 제공                        |
| Legacy parity 및 cutover        | 예정                                                       |

상세한 순서와 완료 조건은 [Migration TODO](./TODO.md)에서 관리한다.

## 요구 사항

- Go 1.26 이상
- CGO를 사용할 수 있는 C toolchain

Tree-sitter Go binding이 CGO를 사용하므로 `CGO_ENABLED=0` 환경에서는 현재 빌드되지 않는다.

## 시작하기

저장소 checkout에서 다음 명령을 실행한다.

```bash
go mod download
go test ./...
go build -o ./bin/deslint ./cmd/deslint
./bin/deslint version
```

Go toolchain을 통해 versioned release의 CLI를 설치할 수도 있다.

```bash
go install github.com/aprilgom/AnslDes/deslint/cmd/deslint@v0.2.1
deslint version
```

## 사용법

### 버전 확인

```bash
go run ./cmd/deslint version
```

현재 개발 버전은 다음처럼 출력된다.

```text
0.1.0-dev
```

### TypeScript/TSX 파싱

```bash
go run ./cmd/deslint parse path/to/Component.tsx
```

`parse`는 현재 lint diagnostic 대신 parser bootstrap용 JSON summary를 출력한다.

```json
{
  "path": "path/to/Component.tsx",
  "language": "tsx",
  "rootKind": "program",
  "hasError": false,
  "namedNodes": 42,
  "nodeKindUses": {
    "identifier": 8,
    "jsx_element": 1
  }
}
```

- 지원 확장자: `.ts`, `.tsx`
- 지원하지 않는 확장자는 오류와 non-zero exit status를 반환한다.
- 구문 오류는 정상 입력으로 숨기지 않고 `hasError: true`로 남긴다.

### 통합 lint

```bash
go run ./cmd/deslint lint \
  --definition ../packages/schema/testdata/example-product.json \
  --policy ../packages/schema/testdata/example-policy.json \
  --source testdata/positive/Example.tsx \
  --pencil testdata/positive/document.pen.json \
  --layout testdata/positive/layout.json \
  --format json \
  --out report.json
```

- `--source`는 반복할 수 있다.
- `--format`은 `text`, `json`, `sarif`를 지원한다.
- required evidence 누락과 stale layout SHA는 pass가 아니다.
- malformed input이나 policy validation 실패 시 기존 `--out` 파일을 덮어쓰지 않는다.
- finding이 policy budget을 초과하면 report를 먼저 원자적으로 기록하고 non-zero로 종료한다.

## 아키텍처

```text
TS/TSX ── Tree-sitter ── Source IR ─┐
contract JSON ───────── Contract IR ─┼─ Rule Engine ── text / JSON / SARIF
.pen JSON ────────────── Pencil IR ──┤
pen.dev output ───────── Layout IR ──┘
```

핵심 경계는 다음과 같다.

- `cmd/deslint`: CLI entry point
- `internal/source`: parser-independent source model
- `internal/source/treesitter`: TypeScript/TSX parser adapter
- `internal/rules`: platform-neutral rule engine 예정 영역
- `internal/contract`: language-neutral contract validator 예정 영역
- `internal/pen`: Pencil document analyzer 예정 영역
- `internal/layout`: computed overflow/overlap analyzer 예정 영역
- `internal/report`: stable text, JSON, SARIF formatter 예정 영역

Rule은 Tree-sitter CST에 직접 결합하지 않고 normalized IR을 소비한다. TypeScript type inference가 필요한
판단은 추후 `SemanticProvider` 경계 뒤의 별도 adapter로 격리한다. 자세한 설계는
[Architecture](./docs/architecture.md)를 참고한다.

## 결정론적 규칙 원칙

각 rule은 최소한 다음을 가져야 한다.

1. 안정적인 rule ID와 severity
2. 명시적인 evidence kind와 platform scope
3. positive 및 negative fixture
4. alias, shadow, spread와 우회 fixture
5. false-positive boundary
6. 안정적인 source range와 diagnostic 정렬
7. owner, rationale와 만료일을 가진 예외 정책

Parser 오류, 실행 실패, stale report 또는 확보하지 못한 runtime evidence는 pass로 처리하지 않는다.
미실행 항목은 `not-run`, 승인된 보류 항목은 `deferred`로 구분한다.

## 개발 및 검증

```bash
go fix -diff ./...
go fmt ./...
go tool -modfile=golangci-lint.mod golangci-lint run
go vet ./...
go test ./...
go test -race ./...
```

`golangci-lint`는 전용 module file에 고정되어 애플리케이션 의존성과 분리된다. 저장소 루트에서는
`npm run lint:deslint`로 같은 검사를 실행한다. `go fix` 제안을 적용하려면 `npm run fix:deslint`, 소스를
변경하지 않고 최신화 여부만 검사하려면 `npm run fix:deslint:check`를 사용한다.

parser나 rule을 변경할 때는 관련 중립 fixture와 우회 fixture를 함께 검증해야 한다. Map iteration이나
filesystem 순서가 report 결과에 영향을 주지 않도록 모든 외부 출력의 정렬 기준을 명시한다.

## 소비 저장소 이관 정책

AnslDes는 제품 구현이나 legacy snapshot을 보관하지 않는다. 소비 저장소가 기존 linter fixture, parity
report, migration note와 rollback release를 소유한다. `deslint`는 중립 fixture로 engine semantics를
검증하고, 제품별 owner·consumer·예외·budget은 외부 policy input으로만 받는다.

새 Go rule은 소비 저장소의 같은 fixture에서 기존 구현과 rule ID, severity, source location과 pass/fail
의미를 비교할 수 있어야 한다. 의도적인 차이는 소비 저장소의 migration note와 승인 fixture로 기록한다.

## 문서

- [Migration TODO](./TODO.md): 구현 순서와 cutover 완료 조건
- [Architecture](./docs/architecture.md): analyzer, contract와 evidence 경계
- [Migration boundary](./docs/migration-boundary.md): engine과 소비 저장소의 parity 소유권 경계
- [Repository guidelines](./AGENTS.md): 개발 규칙과 품질 기준

## 브랜치와 릴리스

- `develop`: 기본 통합 브랜치
- `main`: 검증된 릴리스 브랜치
- 커밋: Conventional Commits와 한국어 설명 사용

정식 릴리스에서는 macOS, Linux와 Windows binary 및 SHA-256 checksum을 제공하고, 소비 저장소는
floating latest 대신 정확한 버전과 checksum을 고정한다.
