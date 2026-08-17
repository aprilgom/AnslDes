# Architecture

## 목표

하나의 Go binary가 framework-neutral contract, React Native source IR, Pencil document와 computed-layout
report를 동일한 diagnostic model로 검사한다. 제품 고유 정책은 config로 주입하고 engine과 분리한다.

```text
TS/TSX ── Tree-sitter ── Source IR ─┐
contract JSON ───────── Contract IR ─┼─ Rule Engine ─ Report(JSON/SARIF/text)
.pen JSON ────────────── Pencil IR ──┤
pen.dev output ───────── Layout IR ──┘
```

## Parser strategy

Tree-sitter는 구문과 concrete source range를 소유한다. `internal/source`가 import, declaration, JSX와
expression을 normalized IR로 바꾸고 lexical symbol table과 module resolver를 적용한다. Rule은 raw CST
node kind에 의존하지 않는다.

Tree-sitter가 제공하지 않는 TypeScript type inference는 `SemanticProvider` 선택 기능으로 격리한다.
TypeScript 7의 stable IPC API가 충분해질 때 adapter를 추가하며 compiler internal package에 결합하지 않는다.

## Contract ownership

Canonical contract는 JSON과 JSON Schema다. Go와 제품 TypeScript는 이를 각자의 typed model로 읽거나
generated adapter로 소비한다. 계약 fingerprint는 canonical JSON bytes로 계산한다.

## Evidence boundary

- `web-source`: Web AST, CSS와 static HTML
- `web-rendered`: browser DOM, computed style와 contrast engine
- `native-source`: React Native AST와 lexical/module binding
- `design-document-source`: Pencil document node, metadata와 variable resolution
- `design-document-computed`: pen.dev engine bounds와 problem output
- `simulator`, `emulator`, `physical-device`: 서로 독립된 native runtime evidence
- `consumer-conformance`: consumer가 추출한 control anatomy, state와 action consistency
- `native-source`: React Native source provider가 증명한 accessibility/performance/platform contract
- `simulator`, `emulator`, `physical-device`: native runtime provider의 독립 capture set

Runtime quality evidence는 `web-rendered`, `simulator`, `emulator`, `physical-device` 중 실제 capture kind를
그대로 사용한다. Web route 성공은 native evidence를 생성하지 않는다. Capture payload의 browser/app failure는
finding이지만 detector process failure는 report 이전 실행 오류다.

Native conformance는 `native-source-evidence`와 `native-runtime-evidence`를 별도 input으로 유지한다. Runtime
capture 안에서도 system gesture, safe area, keyboard/IME, accessibility, font scaling, motion, appearance,
adaptivity, performance와 bundle을 별도 object로 직렬화한다. Consumer policy의 runtime matrix는 capture ID와
platform/kind/device/window/theme/font-scale axes를 exact match한다.

한 evidence kind의 성공으로 다른 platform을 통과 처리하지 않는다.

Evidence result status는 `pass`, `fail`, `advisory`, `false-positive`, `not-run`, `deferred`다. Rule
activation의 `active`, `not-applicable`, `disabled`, `unsupported`와 혼합하지 않는다. 실행 오류는 finding이
아니며 CLI exit code `1`, blocking finding은 exit code `2`를 사용한다.

## Compatibility

Diagnostic의 외부 contract는 canonical `ruleId`, 별도 `sourceRuleIds`, `status`, `severity`, `message`,
`path`, `range`, `evidenceKind`, `platform`, `owner`와 `fingerprint`다. 동일 canonical finding의 upstream
provenance는 정렬·병합한다.

Report는 effective rule pack과 active rule exact set의 fingerprint를 포함한다. False positive는 exact
finding·owner fingerprint와 rationale을 가진 별도 record로 보존한다. Optional visual judgment는 별도
field에 저장하고 deterministic report fingerprint 계산에서 제외한다. 정렬과 직렬화는 OS, map iteration과
optional review 유무에 독립적이어야 한다.

## Rule registry와 profile

Rule metadata는 `internal/rules`의 공통 `RuleSpec` registry에 한 번 등록한다. Rule pack membership,
manifest fingerprint와 effective activation set은 이 registry에서 파생한다. 미지원 pack/version, stale
fingerprint, unknown rule ID와 wildcard override는 policy validation 단계에서 실패한다.

Consumer profile과 activation은 독립적이다. `operate`, `read`, `browse`, `create`는 stable profile ID이며
custom profile은 rationale, reviewer와 evidence owner가 필요하다. Profile은 severity, threshold와 required
evidence만 조정한다. Evidence가 없어 실행할 수 없는 rule은 `not-applicable`, 명시적으로 선택하지 않은 pack의
rule은 `unsupported`, 거버넌스 승인을 받은 exact override만 `disabled`로 직렬화한다.

Canonical definition에서 생성한 design context는 사람이 유지하는 두 번째 token source가 아니다.
Generator marker가 없는 기존 `DESIGN.md`나 sidecar는 덮어쓰지 않으며, sidecar의 source contract SHA가 현재
definition과 다르면 `evidence/stale`로 실패한다. Design-system awareness finding은 AnslDes canonical ID와
Impeccable의 `design-system-font`, `design-system-color`, `design-system-radius`,
`design-system-font-size` source ID를 분리해 보존한다.

Visual-detail provider는 accent edge의 방향·두께, radius, border/elevation, background pattern과 structural
accessory wrapper를 공통 IR로 전달한다. Status/selection과 diagram permission은 semantic state, data meaning,
component/diagram owner가 모두 exact match할 때만 적용되어 card 전체나 decorative background로 확장되지 않는다.

Typography evidence는 surface profile의 threshold와 실제 font scale을 함께 기록한다. 따라서 generic engine은
하나의 고정 ratio를 모든 surface에 적용하지 않으며 Web heading level과 native accessibility heading을 동일한
semantic order로 정규화한다. Impeccable ID와 Hallmark `hallmark-eight-02` provenance는 canonical finding에
별도 source rule ID로 병합한다.

Color evidence는 한 theme의 screenshot 경로와 computed foreground/background 값을 하나의 provider-neutral
IR로 묶는다. Runner는 canonical definition의 모든 theme가 별도 evidence로 제공됐는지 확인하고 `colorUsage`
contrast registry 및 context·theme scoped palette permission을 analyzer에 주입한다. Provider의 승인 boolean만으로
palette finding을 우회할 수 없고, exact media owner 예외도 text contrast 계산에는 영향을 주지 않는다.

Layout-detail evidence는 semantic spacing relation, structural anatomy와 computed viewport bounds를 한 IR에
정규화하되 provider의 capture path와 computed-bounds path를 각각 보존한다. Consumer profile density는 bounded
text padding floor에, prose/data/table role은 line-length applicability에만 영향을 준다. Data grid와 immersive
canvas는 구조적으로 명시해야 Hallmark layout supplement의 negative boundary로 처리된다.

Motion evidence는 Web/native/design-document source와 runtime capture를 동일한 resolved animation IR로
정규화한다. Optional `transitionId`가 있으면 canonical definition의 owner·purpose·duration·easing registry와
대조하고, reduced preference에서는 reduced duration과 fallback을 사용한다. Preference resolution 이전 effect나
resolution 후 완료 effect 재실행은 finding으로 축소하지 않고 malformed execution evidence로 실패시킨다.

Copy evidence는 locale, content role, intent, container와 literal을 정규화한다. Phrase, protected term,
recovery-copy ID와 claim source reference는 consumer policy의 versioned content registry가 소유한다. Copy source
evidence와 content-registry execution evidence를 분리하므로 registry 미실행은 claim pass로 변환되지 않는다.
English em-dash saturation은 고정 advisory이며 budget enforcement에서 blocking count에 포함하지 않는다.

Imagery evidence는 medium, semantic role, geometry, load status와 accessibility state만 전달한다. Asset bytes나
제품 경로는 AnslDes에 복사하지 않고 consumer policy registry의 owner·implementation source·consumer·fingerprint와
exact match한다. Intentional omission은 같은 registry entry에 명시된 경우만 허용된다.

Runtime quality evidence는 surface 아래 route owner와 failure/content/text observation을 안정적인 ID로 정규화한다.
Web failure kind와 native crash/redbox/error-boundary kind는 서로 호환되지 않는다. Print/export body justify는
consumer policy registry의 platform·surface·route·node·owner·context가 모두 일치할 때만 제외된다.

Native source evaluator와 runtime evaluator는 공통 policy threshold와 RuleSpec metadata를 공유하되 evidence를
합성하지 않는다. iOS target은 44pt와 policy spacing, Android target은 48dp와 8dp spacing을 사용한다. 60/120Hz
frame path는 동일한 drop-ratio/main-thread threshold에 대해 refresh-rate axis와 함께 보존된다. Physical-device
미확보 상태는 기존 evidence policy에 따라 `not-run` 또는 명시적 `deferred`로 남는다.

Web provider adapter는 source regex, static HTML, rendered browser와 visual contrast capability를 별도
identity로 유지한다. Provider payload는 consumer policy의 route owner와 viewport/theme/font-scale/
Reduce-Motion exact axis에 대조한다. Embedded catalog가 upstream ID, canonical ID, provider capability와
advisory metadata의 유일한 mapping이며 runner dispatch에는 rule별 조건문이 없다. Generated CSS/vendor
finding은 exact artifact SHA-256과 재현 명령이 있는 정책 record로만 false positive가 되고, fingerprint가
달라지면 즉시 일반 finding으로 복귀한다.

Canonical anti-slop `RuleSpec`은 Web, React Native와 design-document target별 mapping을 exact 3개 가진다.
Mapping state는 `supported`, `shared-intent`, `native-supplement`, `design-document-computed`, `unsupported` 중
하나이며 unsupported에는 이유와 대체 evidence가 필수다. Platform adapter는 이 metadata와 기존 공용 IR
analyzer를 소비하므로 rule별 engine dispatch를 추가하지 않는다. Native source의 justify/reveal/raw-animation
관찰은 `native-source`로 남고 simulator·device evidence로 승격되지 않는다.

Governance exception과 ignore는 rule, engine/evidence kind, platform, path, owner 및 선택적인 property/value의
exact tuple이다. 만료, provider drift와 owner drift는 suppression을 만들지 않는다. Rule activation 상태는
finding status와 독립이며 inactive rule을 false positive로 재분류하는 report도 거부한다. CI execution
verifier는 금지 flag, tool/wrapper exit code, 원본/저장 report SHA와 finding count, 저장 여부를 비교한다.

Budget은 severity/category 집계와 별도로 blocking, advisory, exception, not-run, deferred 상태를 센다. Text와
SARIF도 JSON과 같은 effective rule set을 직렬화하며 SARIF result properties는 evidence kind, platform,
viewport와 owner를 보존한다. Release dependency와 소비 저장소 lock 절차는
[`docs/anti-slop/release-migration.md`](../../docs/anti-slop/release-migration.md)에 고정한다.
