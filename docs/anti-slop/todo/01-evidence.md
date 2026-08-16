# 01. Evidence contract와 audit 경계

> [Anti-slop 인덱스](../TODO.md) · 다음: [Consumer conformance profile](./02-consumer-profile.md)

고정된 detector의 source·rendered 판정과 AnslDes platform provider의 증거 수준을 하나의 evidence
contract로 정의한다.

## 할 일

- [ ] Web source, Web rendered, native source, design-document computed, simulator/emulator, physical device evidence를 구분한다.
- [ ] detector finding과 실행 실패를 각각 exit code 2와 1로 보존한다.
- [ ] deterministic finding과 visual judgment를 별도 report field로 저장한다.
- [ ] false positive는 실제 문맥을 검증한 뒤에만 exact owner fingerprint로 분류한다.
- [ ] audit finding마다 P0–P3, 위치, 사용자 영향, 기준, 수정 방향과 owner를 기록한다.
- [ ] systemic issue와 isolated defect를 분리하고 positive finding도 보존한다.
- [ ] 미확보 browser/native/device evidence를 `not-run` 또는 `deferred`로 표시한다.
- [ ] 명시된 brief, consumer policy, 기존 identity와 native affordance를 자동 rule보다 우선한다.
- [ ] `DESIGN.md` 부재를 greenfield나 visual authority 부재로 해석하지 않는다.
- [ ] refinement가 사실 문구·행동·범위 밖 identity를 바꾸지 않는지 기록한다.

## Consumer profile 경계

AnslDes는 `operate`, `read`, `browse`, `create` 같은 consumer profile의 schema와 판정 우선순위만
정의한다. 어떤 profile을 선택할지, 허용 type ratio와 identity가 무엇인지는 소비 저장소가 정책으로
주입하며 AnslDes 기본값에 특정 제품 유형을 고정하지 않는다.

## 완료 조건

- [ ] 모든 anti-slop 결과가 evidence kind와 platform을 가진다.
- [ ] Web detector가 native 품질 증거로 직렬화되는 경로가 없다.
- [ ] `pass`, `fail`, `advisory`, `false-positive`, `not-run`, `deferred` 상태가 exact enum이다.

## 결정론적 코드 기준

- [snapshot provenance](../references/impeccable-detector-2026/README.md)
- [finding model](../references/impeccable-detector-2026/upstream/cli/engine/findings.mjs)
- [rule registry와 engine support](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
