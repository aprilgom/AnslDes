# 13. Native platform conformance

> 이전: [LLM-only critique](./12-llm-review.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Upstream과 Web gate](./14-web-gate.md)

Native platform conformance를 React Native provider와 소비 저장소의 runtime evidence 계약으로 분리한다.

## 공통

- [ ] Accessibility, Performance, Appearance/Theming, Platform Conformance, Adaptivity를 0–4로 평가한다.
- [ ] Web-shaped control, gratuitous motion, inconsistent affordance와 off-platform overlay를 검출한다.
- [ ] system gesture, safe area, keyboard/IME, dark theme와 font scaling evidence를 분리한다.

## Accessibility

- [ ] interactive label, role/trait, state announcement와 reading/focus order를 검사한다.
- [ ] Dynamic Type/sp scaling에서 clipping, overlap과 unreachable control을 검사한다.
- [ ] iOS 44pt, Android 48dp target과 adjacent target spacing을 플랫폼별로 검사한다.
- [ ] Reduce Motion 대체와 light/dark contrast를 함께 검사한다.

## Performance

- [ ] first frame 전 synchronous startup work와 느린 초기화를 측정한다.
- [ ] 긴 목록이 FlatList 등 virtualization과 stable key를 사용하는지 검사한다.
- [ ] scroll/gesture path의 main-thread work와 60/120Hz frame drop evidence를 기록한다.
- [ ] React Native provider에서 불필요한 rerender, unstable callback/key와 memoization 누락을 검사한다.
- [ ] thumbnail에 full-size image를 decode하거나 cache 없이 반복 load하는지 검사한다.
- [ ] JS bundle/app binary의 unused dependency와 weight regression을 검사한다.

## Appearance와 theming

- [ ] raw color, broken dark appearance와 quick invert를 검사한다.
- [ ] Android Dynamic Color 적용 가능성과 static fallback을 함께 검토한다.
- [ ] hand-rolled material 대신 semantic system/material role을 사용했는지 검사한다.

## Adaptivity

- [ ] phone layout을 tablet에 단순 확대하지 않고 size/window class를 사용하는지 검사한다.
- [ ] portrait/landscape, iPad Split View, Android multi-window와 fold posture를 검사한다.
- [ ] keyboard/IME가 input과 primary action을 가리지 않는지 검사한다.

## iOS

- [ ] 44×44pt touch target과 breathing room을 검사한다.
- [ ] safe area, edge-swipe Back과 navigation stack/sheet 구조를 확인한다.
- [ ] Dynamic Type, 11pt floor와 Reduce Motion 대체를 확인한다.
- [ ] hand-rolled glassmorphism와 bespoke card stack을 검토한다.
- [ ] system navigation/controls와 iconography에서 off-platform drift를 검토한다.
- [ ] Simulator phone/iPad, light/dark와 large text evidence를 기록한다.

## Android

- [ ] 48×48dp touch target과 8dp separation을 검사한다.
- [ ] predictive Back, edge-to-edge insets와 IME를 확인한다.
- [ ] Material type role, sp scaling과 light/dark scheme을 확인한다.
- [ ] compact navigation과 expanded rail/drawer adaptation을 확인한다.
- [ ] Material component/motion과 iconography에서 iOS-shaped drift를 검토한다.
- [ ] emulator phone/tablet, night mode와 font scale evidence를 기록한다.

## 완료 조건

- [ ] iOS/Android source 계약과 runtime evidence가 서로 분리된다.
- [ ] 다섯 audit dimension과 합계 20점, rating band, platform-conformance verdict를 저장한다.
- [ ] 사용자가 defer한 physical-device 항목은 `deferred`이며 자동 pass가 아니다.

## 결정론적 코드 경계

- [snapshot usage boundary](../references/impeccable-detector-2026/README.md#usage-boundary)
- [Web engine support registry](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)

Web detector의 source/DOM 판정을 native runtime 증거로 직렬화하지 않는다.
