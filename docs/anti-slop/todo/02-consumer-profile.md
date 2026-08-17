# 02. Consumer conformance profile

> 이전: [Evidence contract](./01-evidence.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Design-system awareness](./03-design-system-awareness.md)

Detector의 landing-page 성향을 모든 interface에 일괄 적용하지 않도록 consumer가 surface 목적과
permission을 선언하는 제품 중립 profile 계약을 만든다.

## Profile schema

- [x] `operate`, `read`, `browse`, `create`를 stable profile id로 정의한다.
- [x] profile마다 primary user goal, density, novelty tolerance와 native-affordance priority를 선언한다.
- [x] consumer가 custom profile을 추가할 때 id, rationale, reviewer와 evidence owner를 요구한다.
- [x] profile이 rule을 끄지 않고 severity·threshold·required evidence만 좁게 조정하게 한다.
- [x] profile 선택이 없는 경우 detector 기본 판정을 보존한다.

## Rule activation

- [x] consumer policy가 required rule pack의 id, exact version과 SHA-256 fingerprint를 선언한다.
- [x] profile 선택과 rule activation을 분리해 profile이 암묵적으로 rule을 제거하지 못하게 한다.
- [x] per-rule activation은 canonical rule id만 허용하고 wildcard, glob과 unknown id를 거부한다.
- [x] `disabled`는 exact owner, rationale와 expiry/review trigger를 가진 별도 governance record를 요구한다.
- [x] 적용 platform/evidence가 없는 rule은 `disabled`가 아니라 `not-applicable`로 기록한다.

## Operate 중립 fixture

- [x] task completion을 우선하는 합성 form flow fixture를 추가한다.
- [x] 과장된 button, mismatched form control, gratuitous motion과 invented affordance를 검출한다.
- [x] 동일 action·state가 화면마다 다른 shape, label, icon과 feedback을 쓰는지 검사한다.
- [x] interactive component의 default, pressed, focused, disabled, loading, error 상태를 검사한다.

대안 UI와 익숙함에 대한 정성적 평가는 [Optional TODO](../../../TODO_Optional.md)에서 관리한다.

## Permission 경계

- [x] 단일 well-tuned font family와 조밀한 semantic type scale을 자동 slop으로 보지 않는다.
- [x] dense data UI와 prose의 line-length 기준을 분리한다.
- [x] approved motion recipe와 Reduce Motion 대체가 있으면 raw duration 휴리스틱보다 우선한다.
- [x] permission은 exact consumer policy fingerprint와 rationale 없이 broad ignore로 확장되지 않는다.

## 완료 조건

- [x] 중립 fixture에서 profile별 verdict와 systemic inconsistency count가 결정론적으로 재현된다.
- [x] detector finding과 profile conformance finding이 서로 다른 evidence kind를 가진다.
- [x] 같은 policy와 registry에서 active, not-applicable, disabled, unsupported exact set이 결정론적으로 재현된다.
- [x] AnslDes source에 실제 제품명, 브랜드 값, route와 consumer count가 없다.

## 결정론적 코드 기준

- [rule registry](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
- [판정 구현](../references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)
