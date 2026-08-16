# 06. Color와 contrast rules

> 이전: [Typography](./05-typography.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Layout](./07-layout.md)

## Rules

- [ ] `gradient-text`: text fill gradient를 검출한다.
- [ ] `ai-color-palette`: reflex purple/violet gradient와 cyan-on-dark 조합을 검출한다.
- [ ] `cream-palette`: 근거 없는 warm cream/beige page surface를 검출한다.
- [ ] `dark-glow`: chromatic blurred shadow/glow를 검출한다.
- [ ] `radial-halo`: saturated radial halo를 검출한다.
- [ ] `radial-spotlight-glow`: low-opacity decorative radial spotlight를 검출한다.
- [ ] `gray-on-color`: chromatic surface 위 neutral gray text를 검출한다.
- [ ] `low-contrast`: body 4.5:1, large text 3:1 기준을 검사한다.

## Consumer-policy 해석

- [ ] consumer definition의 contrast registry와 theme mapping을 source of truth로 사용한다.
- [ ] Hero/brand/asset exception도 text contrast는 면제하지 않는다.
- [ ] neutral elevation shadow와 decorative colored glow를 구분한다.
- [ ] 각 지원 theme의 screenshot과 computed color evidence를 별도로 저장한다.

## 완료 조건

- [ ] 여덟 rule의 literal, token, computed-color fixture가 통과한다.
- [ ] palette 판단은 hue 문자열만이 아니라 승인 palette와 사용 문맥을 확인한다.

## 결정론적 코드 기준

- [color·contrast 판정 구현](../references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)
- [색 파싱·휘도·contrast 계산](../references/impeccable-detector-2026/upstream/cli/engine/shared/color.mjs)
- [static CSS cascade](../references/impeccable-detector-2026/upstream/cli/engine/engines/static-html/css-cascade.mjs)
- [screenshot contrast](../references/impeccable-detector-2026/upstream/cli/engine/engines/visual/screenshot-contrast.mjs)
