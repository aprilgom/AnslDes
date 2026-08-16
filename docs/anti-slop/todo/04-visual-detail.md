# 04. Visual detail rules

> 이전: [Design-system awareness](./03-design-system-awareness.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Typography](./05-typography.md)

## Rules

- [ ] `side-tab`: card/list/callout 한쪽의 2px 이상 chromatic stripe를 검출한다.
- [ ] `border-accent-on-rounded`: rounded surface의 지배적인 accent edge를 검출한다.
- [ ] `gpt-thin-border-wide-shadow`: hairline border와 wide diffuse shadow 조합을 검출한다.
- [ ] `repeating-stripes-gradient`: 의미 없는 repeating gradient stripe를 검출한다.
- [ ] `codex-grid-background`: canvas/map/measurement가 아닌 decorative grid를 검출한다.

## Consumer-policy 해석

- [ ] error/status/selection edge는 semantic state와 exact component owner가 있을 때만 예외로 둔다.
- [ ] ListRow accessory chevron에 별도 tile/card wrapper가 생기는 회귀를 structural lint로 막는다.
- [ ] border와 elevation을 동시에 선언하는 surface recipe를 검사한다.
- [ ] design-document node와 native style provider에서 accent edge 방향·두께·radius 관계를 검사한다.
- [ ] grid/stripe는 실제 diagram owner와 데이터 의미를 증명하지 못하면 허용하지 않는다.

## 완료 조건

- [ ] 다섯 rule의 source, native, design-document fixture가 각각 정확한 owner를 보고한다.
- [ ] intentional status 예외가 broad card 예외로 확장되지 않는다.

## 결정론적 코드 기준

- [visual-detail 판정 구현](../references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)
- [rule registry](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
