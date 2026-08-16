# 05. Typography와 hierarchy rules

> 이전: [Visual detail](./04-visual-detail.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Color](./06-color.md)

## Rules

- [ ] `overused-font`: consumer가 승인한 family registry 밖의 font와 variant drift를 검사한다.
- [ ] `flat-type-hierarchy`: consumer profile의 type ratio와 semantic role 대비를 함께 검사한다.
- [ ] `icon-tile-stack`: heading 위 rounded icon tile template를 검출한다.
- [ ] `italic-serif-display`: 승인되지 않은 italic serif hero display를 검출한다.
- [ ] `hero-eyebrow-chip`: hero headline 위 eyebrow/pill을 검출한다.
- [ ] `kicker-above-heading`: heading 위 별도 tracked kicker를 검출한다.
- [ ] `oversized-h1`: 긴 문장형 display title과 first-viewport 점유를 검사한다.
- [ ] `extreme-negative-tracking`: 글자 형태를 무너뜨리는 negative tracking을 검출한다.
- [ ] `tight-leading`: multi-line text의 role별 최소 line-height를 검사한다.
- [ ] `tiny-text`: body content의 최소 크기를 검사한다.
- [ ] `undersized-ui-text`: platform floor와 consumer가 선언한 더 강한 기능 텍스트 기준을 적용한다.
- [ ] `all-caps-body`: 긴 영문 all-caps 본문을 검출한다.
- [ ] `wide-tracking`: body의 과도한 letter spacing을 검출한다.
- [ ] `skipped-heading`: Web heading outline과 native accessibility heading order를 검사한다.

## Consumer-policy 해석

- [ ] 단일 family 자체를 실패시키지 않고 역할·크기·weight·line-height 결합을 판단한다.
- [ ] consumer가 선언한 heading, body, label과 metadata semantic order를 재사용한다.
- [ ] display, guide, legal copy와 task flow의 type-scale 정책을 분리한다.
- [ ] consumer가 지원하는 font-scale matrix에서 hierarchy가 보존되는지 report에 연결한다.

## 완료 조건

- [ ] 14개 upstream id와 typography contract mapping이 exact하다.
- [ ] 특정 surface용 ratio가 다른 consumer profile에 오탐으로 강제되지 않는다.

## 결정론적 코드 기준

- [typography 판정 구현](../references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)
- [font helpers](../references/impeccable-detector-2026/upstream/cli/engine/shared/fonts.mjs)
- [rule registry](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
