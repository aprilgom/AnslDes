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
| JSON/SARIF report               | versioned JSON Schema, text, native JSON, SARIF 2.1.0 제공 |
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
  --conformance ../packages/schema/testdata/operate-conformance.json \
  --design-context ../packages/schema/testdata/generated-design-context/.impeccable/design.json \
  --visual-detail ../packages/schema/testdata/visual-detail-native.json \
  --typography ../packages/schema/testdata/typography-positive-235.json \
  --color ../packages/schema/testdata/color-negative-light.json \
  --color ../packages/schema/testdata/color-permissions-dark.json \
  --layout-detail ../packages/schema/testdata/layout-negative-web.json \
  --motion ../packages/schema/testdata/motion-reduced-simulator.json \
  --copy ../packages/schema/testdata/copy-ko-positive.json \
  --imagery ../packages/schema/testdata/imagery-permissions.json \
  --runtime ../packages/schema/testdata/runtime-permissions.json \
  --native-source-conformance ../packages/schema/testdata/native-source-positive.json \
  --native-runtime-conformance ../packages/schema/testdata/native-runtime-positive-ios.json \
  --native-runtime-conformance ../packages/schema/testdata/native-runtime-positive-android.json \
  --format json \
  --out report.json
```

- `--source`는 반복할 수 있다.
- `--format`은 `text`, `json`, `sarif`를 지원한다.
- required evidence 누락과 stale layout SHA는 pass가 아니다.
- evidence는 `web-source`, `web-rendered`, `native-source`, `design-document-source`,
  `design-document-computed`, `simulator`, `emulator`, `physical-device`, `consumer-conformance`를 서로
  다른 kind로 기록한다.
- malformed input이나 policy validation 실패 시 기존 `--out` 파일을 덮어쓰지 않는다.
- finding이 policy budget을 초과하면 report를 먼저 원자적으로 기록하고 exit code `2`로 종료한다.
- malformed input, provider 실패와 policy validation 오류는 report를 덮어쓰지 않고 exit code `1`로 종료한다.

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
- `internal/conformance`: consumer control/state/action 적합성 analyzer
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

Native JSON report는 [`deslint-report.schema.json`](../packages/schema/deslint-report.schema.json)을 따른다.
Finding status와 rule activation status를 분리하며, exact false-positive record를 report에 보존한다. Optional
visual judgment는 별도 field에 저장하고 deterministic verdict와 report fingerprint에는 포함하지 않는다.

`internal/rules`의 단일 `RuleSpec` registry가 rule ID, evidence kind와 platform applicability를 소유한다.
Versioned rule pack은 exact member set의 SHA-256을 가지며 consumer policy는 id, version과 fingerprint를
모두 고정한다. Profile은 severity, threshold와 required evidence만 조정하고 rule 비활성화는 exact rule ID,
owner, rationale, expiry와 review trigger가 있는 별도 `ruleOverrides` record로만 선언한다.

Compiler의 `design-context` 명령은 canonical definition에서 generated `DESIGN.md`와
`.impeccable/design.json`을 만든다. Sidecar는 generator version과 source contract SHA-256을 포함하며,
`deslint --design-context`는 stale context를 실패시키고 font, color, radius와 typography ramp drift를
각각 exact upstream rule ID provenance와 함께 검사한다.

`--visual-detail`은 반복 가능하며 Web source, native source와 design-document source가 생성한 동일한
구조 IR을 받는다. Accent edge, border/elevation, decorative stripe/grid 판정은 공통 규칙을 사용하지만
diagnostic evidence kind와 exact node owner는 provider별로 보존한다.

`--typography`도 반복 가능하며 profile별 ratio, role floor, platform floor와 font scale을 evidence에 고정한다.
단일 family 사용은 허용하고 family/variant drift, semantic hierarchy, leading, tracking, heading order와
display template을 결합 판정한다.

`--color`는 지원 theme마다 별도 screenshot·computed-color evidence를 받는다. Body/large-text contrast
threshold와 palette 승인은 canonical definition의 `colorUsage` registry를 사용하며, palette ID가 theme와
context에 exact match할 때만 승인한다. Brand·asset owner 예외도 text contrast를 면제하지 않는다.

`--layout-detail`은 반복 가능한 browser/native/design-document 공용 IR이다. Semantic spacing relation과
computed bounds를 함께 기록하고 consumer profile의 density 및 prose/data/table role을 적용한다. 기존
`--layout`의 Pencil overflow/overlap report와 별도로 보존되므로 provider별 증거가 서로를 덮어쓰지 않는다.

`--motion`은 source와 runtime에서 해석된 animation을 반복 입력한다. `transitionId`가 있으면 canonical
definition motion registry의 owner, purpose, resolved duration/easing과 exact match해야 한다. Reduce Motion
evidence는 preference를 effect보다 먼저 해석하고 완료된 side effect를 재실행하지 않으며, state와 다음 행동을
정적으로도 이해할 수 있음을 함께 기록해야 한다.

`--copy`는 locale, content role, intent와 container 구조를 기록한다. Phrase/recovery/source-reference 값은
consumer policy의 versioned `content` registry에서만 주입한다. Claim truth는 추론하지 않으며 registry가
실행되지 않았으면 별도 `consumer-content-registry` evidence를 `not-run`으로 남긴다. English em-dash saturation은
항상 advisory이고 blocking budget으로 승격되지 않는다.

`--imagery`는 Web `img`/poster, native image와 inline SVG를 공통 IR로 받는다. Consumer policy의 versioned
asset registry에서 owner, implementation source, consumer, role과 fingerprint를 exact match하고 decorative
screen-reader exclusion 및 functional label을 함께 검사한다. Registry가 명시한 intentional omission만 source
부재에서 제외된다.

`--runtime`은 route별 console error, unhandled rejection, render failure와 native crash/redbox/error-boundary를
각 platform runtime evidence로 받는다. Reveal이 끝난 primary content와 script 실패 시 default visibility도
같이 검사한다. Body justify는 screen에서 금지하며 print/export는 policy의 versioned `runtime` registry에 있는
platform·surface·route·node·owner·context exact tuple만 허용한다. Browser detector launch/navigation/injection/
collection 실패는 `runtime/script-error` finding이 아니라 실행 오류로 반환한다.

`--native-source-conformance`는 React Native source에서 label/role/state/order, list virtualization/stable key,
render identity, thumbnail decode/cache, dependency use, semantic material, size/window class와 platform Back 계약을
받는다. `--native-runtime-conformance`는 simulator/emulator/physical-device capture를 별도로 받아 system gesture,
safe-area/edge-to-edge, IME, 44pt/48dp target, scaled text, Reduce Motion/contrast, startup/frame/bundle과 adaptivity를
검사한다. Policy의 exact capture matrix가 phone/tablet/foldable, theme, font scale, orientation, window/fold mode를
고정하며 Web이나 source 성공은 runtime capture를 완료시키지 않는다.

Native source provider는 justify, hidden-at-rest fallback과 raw animation registry bypass도 같은
`native-source` evidence로 전달한다. Design-document computed layout은 document SHA-256, resolved-instance/
computed-bounds visitor option, node count와 stable issue identity를 입력에 보존한다. 이 두 provider는 canonical
anti-slop rule의 target별 applicability mapping을 공유하며 unsupported mapping은 대체 evidence를 명시한다.

Governance는 broad ignore나 wrapper 후처리가 아니라 policy input과 immutable execution evidence로 검사한다.
Exact exception은 finding owner와 provider evidence kind까지 일치해야 하며 disabled rule은 owning pack의 exact
version을 가리킨다. `--no-config`, `--no-design-system`, `--no-inline-ignores`, `--no-advisory`, exit code 2
rewrite, report SHA/count 변경과 fail report의 강제 저장은 각각 typed governance violation이다.

`--web-provider`는 반복 가능하며 pinned detector의 `regex-source`, `static-html`, `browser`,
`visual-contrast` 결과를 받는다. Upstream rule ID는 embedded 63-rule catalog의 canonical ID로 변환하고,
advisory는 warning으로 보존한다. `completed/full`만 capture matrix를 충족하며 parser dependency가 없어
`regex-fallback`이 된 결과는 `not-run`이다. Provider launch/navigation 실패는 finding이 아닌 exit code 1의
execution error다. Route와 build command, 실제 baseline은 소비 저장소가 소유한다.

`--stage-execution`은 provider process가 남긴 command, owner, platform, exit code, stdout/stderr와 locked/observed
dependency SHA-256을 받는다. 실패 command와 stale dependency는 execution evidence와 blocking diagnostic으로
남으며 다른 stage의 성공으로 대체되지 않는다.

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
