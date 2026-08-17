# 13. Native platform conformance

> 이전: [Runtime](./11-runtime.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Upstream과 Web gate](./14-web-gate.md)

Native platform conformance를 React Native provider와 소비 저장소의 runtime evidence 계약으로 분리한다.

## 공통

- [x] system gesture, safe area, keyboard/IME, dark theme와 font scaling evidence를 분리한다.

## Accessibility

- [x] interactive label, role/trait, state announcement와 reading/focus order를 검사한다.
- [x] Dynamic Type/sp scaling에서 clipping, overlap과 unreachable control을 검사한다.
- [x] iOS 44pt, Android 48dp target과 adjacent target spacing을 플랫폼별로 검사한다.
- [x] Reduce Motion 대체와 light/dark contrast를 함께 검사한다.

## Performance

- [x] first frame 전 synchronous startup work와 느린 초기화를 측정한다.
- [x] 긴 목록이 FlatList 등 virtualization과 stable key를 사용하는지 검사한다.
- [x] scroll/gesture path의 main-thread work와 60/120Hz frame drop evidence를 기록한다.
- [x] React Native provider에서 불필요한 rerender, unstable callback/key와 memoization 누락을 검사한다.
- [x] thumbnail에 full-size image를 decode하거나 cache 없이 반복 load하는지 검사한다.
- [x] JS bundle/app binary의 unused dependency와 weight regression을 검사한다.

## Appearance와 theming

- [x] raw color, broken dark appearance와 quick invert를 검사한다.
- [x] hand-rolled material 대신 semantic system/material role을 사용했는지 검사한다.

## Adaptivity

- [x] phone layout을 tablet에 단순 확대하지 않고 size/window class를 사용하는지 검사한다.
- [x] portrait/landscape, iPad Split View, Android multi-window와 fold posture를 검사한다.
- [x] keyboard/IME가 input과 primary action을 가리지 않는지 검사한다.

## iOS

- [x] 44×44pt touch target과 consumer policy의 adjacent-target spacing을 검사한다.
- [x] safe area, edge-swipe Back과 navigation stack/sheet 구조를 확인한다.
- [x] Dynamic Type, 11pt floor와 Reduce Motion 대체를 확인한다.
- [x] Simulator phone/iPad, light/dark와 large text evidence를 기록한다.

## Android

- [x] 48×48dp touch target과 8dp separation을 검사한다.
- [x] predictive Back, edge-to-edge insets와 IME를 확인한다.
- [x] Material type role, sp scaling과 light/dark scheme을 확인한다.
- [x] compact navigation과 expanded rail/drawer adaptation을 확인한다.
- [x] emulator phone/tablet, night mode와 font scale evidence를 기록한다.

## 완료 조건

- [x] iOS/Android source 계약과 runtime evidence가 서로 분리된다.
- [x] 사용자가 defer한 physical-device 항목은 `deferred`이며 자동 pass가 아니다.

## 구현 경계

- `native-source-evidence.schema.json`과 `native-runtime-evidence.schema.json`은 서로 대체할 수 없다.
- 15개 native rule은 source-only, runtime-only 또는 공통 `RequiredInputs` metadata로 applicability를 선언한다.
- Consumer policy `native` registry가 threshold, iOS spacing과 required runtime capture matrix를 versioning한다.
- Simulator/emulator matrix의 누락은 `native/runtime-matrix`; physical-device defer는 evidence `deferred`다.

platform 적합성의 정성적 점수와 visual drift 평가는
[Optional TODO](../../../TODO_Optional.md)에서 관리한다.

## 결정론적 코드 경계

- [snapshot usage boundary](../references/impeccable-detector-2026/README.md#usage-boundary)
- [Web engine support registry](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)

Web detector의 source/DOM 판정을 native runtime 증거로 직렬화하지 않는다.
