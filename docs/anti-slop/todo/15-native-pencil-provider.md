# 15. React Native와 design-document provider

> 이전: [Upstream과 Web provider](./14-web-gate.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Governance](./16-governance.md)

## Mapping

- [x] canonical 63개 rule 각각에 Web-only, shared-intent, native supplement,
      design-document-computed 적용 상태를 기록한다.
- [x] unsupported rule은 이유와 대체 evidence를 가진다.
- [x] 기존 color/layout/motion/icon/typography/UX lint는 중복 구현하지 않고 dependency로 연결한다.

## 우선 구현

- [x] side-tab과 rounded accent border structural lint를 추가한다.
- [x] nested card와 separate accessory tile depth lint를 추가한다.
- [x] icon-tile-above-heading과 decorative status dot을 검사한다.
- [x] equal icon feature columns와 full-viewport centered composition을 platform evidence별로 검사한다.
- [x] decorative glow/grid/stripe와 repeated container copy를 검사한다.
- [x] design-document visitor에서 surface depth, heading adjacency, overflow/overlap과 token owner를 검사한다.
- [x] React Native source provider에서 justify, raw animation과 hidden-at-rest 패턴을 검사한다.

## Evidence

- [x] source diagnostic과 design-document node diagnostic에 stable rule id를 사용한다.
- [x] layout visitor option, document fingerprint, node count와 issues를 report에 저장한다.
- [x] native source lint가 simulator/device evidence로 오인되지 않게 한다.

## 완료 조건

- [x] 모든 mapping이 exact-set test를 통과한다.
- [x] 중립 native/design-document fixture의 적용 가능한 blocking finding과 warning이 0이다.

구현 경계: anti-slop pack의 63개 `RuleSpec`은 Web, React Native, design-document target별 applicability를
각각 하나씩 가지며 unsupported mapping은 이유와 대체 evidence를 함께 기록한다. 기존 visual-detail,
typography, color, layout, motion, copy, imagery와 runtime analyzer를 dependency로 재사용하고,
`native/list-row-accessory-wrapper`만 icon-tile 의도의 native supplement로 연결한다. 중립 native source와
design-document computed positive fixture는 finding 0을 보장하고 source evidence를 runtime capture로 합성하지
않는다.

## 결정론적 코드 기준

- [rule registry](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
- [판정 구현](../references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)
- [snapshot usage boundary](../references/impeccable-detector-2026/README.md#usage-boundary)
- [Hallmark source mapping](../references/hallmark-eight-tells-2026/README.md)
