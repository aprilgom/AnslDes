# 02. Consumer conformance profile

> 이전: [Evidence contract](./01-evidence.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Design-system awareness](./03-design-system-awareness.md)

Detector의 landing-page 성향을 모든 interface에 일괄 적용하지 않도록 consumer가 surface 목적과
permission을 선언하는 제품 중립 profile 계약을 만든다.

## Profile schema

- [ ] `operate`, `read`, `browse`, `create`를 stable profile id로 정의한다.
- [ ] profile마다 primary user goal, density, novelty tolerance와 native-affordance priority를 선언한다.
- [ ] consumer가 custom profile을 추가할 때 id, rationale, reviewer와 evidence owner를 요구한다.
- [ ] profile이 rule을 끄지 않고 severity·threshold·required evidence만 좁게 조정하게 한다.
- [ ] profile 선택이 없는 경우 detector 기본 판정을 보존한다.

## Operate 중립 fixture

- [ ] task completion을 우선하는 합성 form flow fixture를 추가한다.
- [ ] 과장된 button, mismatched form control, gratuitous motion과 invented affordance를 검출한다.
- [ ] 동일 action·state가 화면마다 다른 shape, label, icon과 feedback을 쓰는지 검사한다.
- [ ] interactive component의 default, pressed, focused, disabled, loading, error 상태를 검사한다.
- [ ] overlay가 첫 해결책인지, inline 또는 progressive disclosure가 가능한지 review evidence로 남긴다.
- [ ] 표준 navigation, 익숙한 control과 화면 간 반복 일관성을 positive finding으로 보존한다.

## Permission 경계

- [ ] 단일 well-tuned font family와 조밀한 semantic type scale을 자동 slop으로 보지 않는다.
- [ ] dense data UI와 prose의 line-length 기준을 분리한다.
- [ ] approved motion recipe와 Reduce Motion 대체가 있으면 raw duration 휴리스틱보다 우선한다.
- [ ] permission은 exact consumer policy fingerprint와 rationale 없이 broad ignore로 확장되지 않는다.

## 완료 조건

- [ ] 중립 fixture에서 profile별 verdict와 systemic inconsistency count가 결정론적으로 재현된다.
- [ ] detector finding과 profile conformance finding이 서로 다른 evidence kind를 가진다.
- [ ] AnslDes source에 실제 제품명, 브랜드 값, route와 consumer count가 없다.

## 결정론적 코드 기준

- [rule registry](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
- [판정 구현](../references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)
