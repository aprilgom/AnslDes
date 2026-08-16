# 08. Motion rules

> 이전: [Layout](./07-layout.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Copy](./09-copy.md)

## Rules

- [ ] `bounce-easing`: reflex bounce/elastic/overshoot easing을 검출한다.
- [ ] `pulsing-dot`: live data와 무관한 decorative status pulse를 검출한다.
- [ ] `blinking-cursor`: editable input 밖 fake caret를 검출한다.
- [ ] `marquee`: 사용자 제어 없는 perpetual auto-scroll을 검출한다.
- [ ] `layout-transition`: width/height/padding/margin 중심 animation을 검출한다.
- [ ] `image-hover-transform`: 목적 없는 hover scale/rotate를 검출한다.

## Consumer-policy 해석

- [ ] consumer definition의 motion owner, duration/easing resolver와 Reduce Motion 경계를 재사용한다.
- [ ] 상태 변화, feedback, loading, reveal 외 decorative motion을 실패시킨다.
- [ ] status pulse 예외는 실제 changing-data source와 accessible label을 요구한다.
- [ ] platform별 Reduce Motion 설정의 대체 상태를 기록한다.
- [ ] preference resolution이 완료된 animation/side effect를 재실행하지 않는지 검사한다.

## 완료 조건

- [ ] 여섯 rule과 consumer motion registry mapping이 exact하다.
- [ ] reduced-motion에서도 state와 다음 행동을 이해할 수 있다.

## 결정론적 코드 기준

- [motion 판정 구현](../references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)
- [regex source engine](../references/impeccable-detector-2026/upstream/cli/engine/engines/regex/detect-text.mjs)
