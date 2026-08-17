# AnslDes Architecture

## Dependency direction

```text
@ansldes/schema
       ↓
@ansldes/compiler ──────▶ canonical bundle + fingerprint
       ↓
@ansldes/core
       ↓
@ansldes/react-native
       ↓
product application

deslint ── reads schema + product definition + product policy + evidence
```

제품은 AnslDes를 소비한다. AnslDes package는 제품 package, 제품 source와 제품 asset을 import하지 않는다.

## Definition and policy

두 입력은 목적이 다르므로 분리한다.

- **definition**: color, typography, spacing, radius, size, motion과 component recipe 값
- **policy**: source owner, allowed consumer, asset exception, lint severity와 expiry

Definition은 UI를 렌더링하는 입력이고 policy는 `deslint`가 제품 source를 감사하는 입력이다. 제품 source
경로를 definition에 넣지 않는다.

Policy v1은 rule severity, source raw-property 분류, exact exclusion, required evidence, budget과
owner·rationale·expiry를 가진 exact exception만 허용한다. Asset, content와 runtime permission registry도
versioned policy input이며 runtime print/export 예외는 platform·surface·route·node·owner·context exact tuple이다.
Glob이나 absolute-path exclusion은 schema 이후 Go validator에서도 다시 거부한다.

Native policy는 source와 runtime provider가 공유하는 registry version, performance threshold, iOS adjacent-target
spacing과 exact runtime capture matrix만 소유한다. 실제 screen, device snapshot과 측정 report는 소비 저장소가
소유한다.

## Stable generic vocabulary

AnslDes는 token 이름 자체를 브랜드별로 고정하지 않는다. 대신 다음 구조와 reference type을 고정한다.

- foundations: color, spacing, radius, size, typography, motion, elevation, layer
- component: anatomy, slots, variants, sizes, states와 semantics
- theme axis: product-defined names와 하나의 default
- reference: `{collection.layer.name}`

공용 component는 제품 semantic token 이름을 직접 추측하지 않고 제품이 제공한 component recipe만 소비한다.

## Release boundary

- AnslDes schema/runtime: semantic version
- product definition: 제품 독립 version
- product policy: definition과 별도 fingerprint

제품의 blue 값이나 control radius 변경은 AnslDes release가 아니다. Schema field, component anatomy 또는
runtime behavior가 바뀔 때만 AnslDes release가 필요하다.

## Lint evidence and report

`deslint`는 definition, source, Pencil, computed-layout을 별도 evidence kind로 기록한다. 하나가 없거나 stale인
경우 다른 evidence의 성공으로 대체하지 않는다. diagnostic은 path·range·rule ID 순으로 정렬되고 text,
native JSON과 SARIF 2.1.0이 같은 finding fingerprint를 공유한다.

Report의 effective pack·rule exact set과 activation reason은 text, JSON, SARIF에 공통으로 들어간다. Rendered
finding은 provider engine, platform, viewport와 policy가 주입한 owner를 함께 보존한다. 소비 저장소는
`consumer-release-lock.schema.json`으로 release manifest, package·binary checksum, rule pack과 detector
dependency를 exact pin하며 실제 parity·rollback evidence는 자체 저장소에 둔다.
