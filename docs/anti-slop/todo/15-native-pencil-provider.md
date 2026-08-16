# 15. React Native와 design-document provider

> 이전: [Upstream과 Web provider](./14-web-gate.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Governance](./16-governance.md)

## Mapping

- [ ] canonical 63개 rule 각각에 Web-only, shared-intent, native supplement,
      design-document-computed 적용 상태를 기록한다.
- [ ] unsupported rule은 이유와 대체 evidence를 가진다.
- [ ] 기존 color/layout/motion/icon/typography/UX lint는 중복 구현하지 않고 dependency로 연결한다.

## 우선 구현

- [ ] side-tab과 rounded accent border structural lint를 추가한다.
- [ ] nested card와 separate accessory tile depth lint를 추가한다.
- [ ] icon-tile-above-heading과 decorative status dot을 검사한다.
- [ ] equal icon feature columns와 full-viewport centered composition을 platform evidence별로 검사한다.
- [ ] decorative glow/grid/stripe와 repeated container copy를 검사한다.
- [ ] design-document visitor에서 surface depth, heading adjacency, overflow/overlap과 token owner를 검사한다.
- [ ] React Native source provider에서 justify, raw animation과 hidden-at-rest 패턴을 검사한다.

## Evidence

- [ ] source diagnostic과 design-document node diagnostic에 stable rule id를 사용한다.
- [ ] layout visitor option, document fingerprint, node count와 issues를 report에 저장한다.
- [ ] native source lint가 simulator/device evidence로 오인되지 않게 한다.

## 완료 조건

- [ ] 모든 mapping이 exact-set test를 통과한다.
- [ ] 중립 native/design-document fixture의 적용 가능한 blocking finding과 warning이 0이다.

## 결정론적 코드 기준

- [rule registry](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
- [판정 구현](../references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)
- [snapshot usage boundary](../references/impeccable-detector-2026/README.md#usage-boundary)
- [Hallmark source mapping](../references/hallmark-eight-tells-2026/README.md)
