# 03. Design-system awareness rules

> 이전: [Consumer conformance profile](./02-consumer-profile.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Visual detail](./04-visual-detail.md)

## Rules

- [ ] `design-system-font`: canonical family/variant 밖의 font와 physical/logical weight drift를 실패시킨다.
- [ ] `design-system-color`: raw color와 semantic/component/asset exception 밖의 값을 실패시킨다.
- [ ] `design-system-radius`: surface/control/compact/pill 계약 밖의 radius를 실패시킨다.
- [ ] `design-system-font-size`: semantic typography ramp 밖의 size와 role mismatch를 실패시킨다.

## 구현

- [ ] 기존 framework-neutral contract에서 `DESIGN.md`와 `.impeccable/design.json`을 생성한다.
- [ ] 생성 파일의 source contract SHA와 generator version을 저장한다.
- [ ] 사람이 복사한 token 목록과 generated snapshot이 함께 존재하지 못하게 한다.
- [ ] upstream advisory보다 consumer policy의 명시적 error budget이 강하면 error를 유지한다.
- [ ] font/color/radius/size 각각 unknown, stale sidecar와 config-disabled fixture를 추가한다.

## 완료 조건

- [ ] 네 rule의 upstream id와 AnslDes rule mapping이 exact하다.
- [ ] generated design context와 canonical contract fingerprint가 최신이다.

## 결정론적 코드 기준

- [design-system 검사 구현](../references/impeccable-detector-2026/upstream/cli/engine/design-system.mjs)
- [rule registry](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
