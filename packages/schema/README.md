# @ansldes/schema

제품 디자인 시스템 definition의 언어 중립 schema다. 이 package에는 실제 브랜드 token 값이나 제품
source policy가 포함되지 않는다.

Definition v2는 다음을 분리한다.

- 제품 identity와 독립 version
- theme axis
- foundation token
- optional icon geometry, usage와 action recipe
- component anatomy와 recipe

제품은 [`design-system-definition.schema.json`](./design-system-definition.schema.json)을 참조하는 JSON을
자기 저장소에서 소유한다. Reference 형식은 `{collection.layer.name}`이다.

현재 schema는 구조와 값 종류를 검증한다. Reference 존재, cycle, theme exact-set과 layer compatibility는
compiler 및 `deslint`가 추가 검증한다.

Color, spacing, radius, size는 예외 없이 `primitive → semantic → component` 순서로만 참조한다.
숫자와 원시 색상은 primitive에서만 선언하며 semantic은 같은 collection의 primitive만, component는
같은 collection의 semantic만 참조한다. Component recipe는 이 네 collection의 component token만
소비한다. v1 입력과 bundle은 지원하지 않는다.

`foundations.icon`은 definition v2의 선택 그룹이다. size, stroke, optical alignment, geometry, usage와
action을 제품 저장소가 정의하며, schema에는 특정 glyph path나 제품별 consumer가 들어가지 않는다.

## Inputs

- `design-system-definition.schema.json`: 렌더링에 필요한 제품 token과 component recipe
- `design-system-policy.schema.json`: severity, profile, versioned rule pack, exact rule override, evidence와 budget
- `deslint-report.schema.json`: evidence kind, rule activation, finding, false-positive와 optional judgment를
  분리하는 결정론적 report
- `consumer-conformance.schema.json`: provider가 생성하는 control anatomy, state, action consistency와
  profile conformance evidence
- `consumer-release-lock.schema.json`: release manifest, package·binary checksum, rule pack과 detector dependency의 exact pin
- `design-context.schema.json`: canonical definition에서 생성한 `DESIGN.md` sidecar와 source contract SHA
- `visual-detail-evidence.schema.json`: Web source, native source와 design-document provider의 visual-detail IR
- `typography-evidence.schema.json`: profile별 type threshold와 font-scale을 포함하는 rendered/native typography IR
- `color-evidence.schema.json`: theme별 screenshot과 computed foreground/background contrast IR
- `layout-evidence.schema.json`: browser/native/design-document 공용 semantic spacing과 computed bounds IR
- `motion-evidence.schema.json`: source/runtime motion과 platform Reduce Motion fallback IR
- `copy-evidence.schema.json`: locale별 copy structure와 consumer content provenance IR
- `imagery-evidence.schema.json`: Web/native/design-document asset load, role, geometry와 accessibility IR
- `runtime-evidence.schema.json`: route별 Web/native failure, reveal fallback와 resolved text alignment IR
- `native-source-evidence.schema.json`: React Native 접근성, 목록/render/image/dependency와 platform contract IR
- `native-runtime-evidence.schema.json`: simulator/emulator/device별 gesture, inset, IME, appearance와 성능 IR
- `web-provider-evidence.schema.json`: source/static HTML/browser/visual-contrast provider의 실행 상태와 canonical finding IR
- `stage-execution-evidence.schema.json`: provider command, owner, exit code, stdout/stderr와 dependency SHA freshness

Definition의 optional `colorUsage`는 color evidence를 실행할 때 필요한 body/large-text contrast registry와
theme·context별 approved palette ID를 소유한다. 실제 palette 값과 제품별 ID는 소비 저장소 definition이 주입한다.

Definition과 policy는 별도 versioned input이다. 제품 source 경로와 예외는 definition이 아니라 policy에만
둔다. Policy의 optional `runtime` registry는 print/export justify 예외를 platform, surface, route, node,
owner와 context의 exact tuple로 소유한다.

Policy의 optional `native` registry는 iOS adjacent-target spacing, startup/frame/image/bundle threshold와
필수 runtime capture matrix를 소유한다. Source evidence와 simulator/emulator/physical-device evidence는
서로 다른 schema이며 어느 한쪽의 성공도 다른 쪽을 생성하지 않는다.

Native source evidence의 pattern inventory는 resolved text alignment, hidden-at-rest fallback과 raw animation/
transition registry binding을 기록한다. Design-document layout evidence는 document fingerprint, exact visitor
options, node count와 stable overflow/overlap/clipping issues를 함께 보존한다.

Policy의 optional `web` registry는 소비 저장소가 소유하는 build command와 route, viewport, theme,
font-scale, Reduce Motion capture matrix를 선언한다. Generated artifact 제외는 exact relative path와 SHA-256,
owner, rationale, reproduction command가 모두 일치할 때만 false positive로 분류된다.

필수 `governance` contract는 90일 review scope, CI 금지 flag, advisory 보존, exit-code/report 불변성과 exact
ignore allowlist를 소유한다. Exception과 disabled override는 rule/pack, evidence engine, platform, path, finding
owner, reviewer, expiry와 review trigger를 명시해야 하며 provider나 owner가 달라지면 적용되지 않는다.

## Generated models

```bash
npm run generate:models
npm run generate:models:check
```

contract schema에서 TypeScript model과 `deslint` Go model을 결정론적으로 생성한다. 생성물은 schema SHA-256을
포함하며 freshness 검사가 root quality gate에 연결돼 있다.
