# 01. Evidence contract와 audit 경계

> [Anti-slop 인덱스](../TODO.md) · 다음: [Consumer conformance profile](./02-consumer-profile.md)

고정된 detector의 source·rendered 판정과 AnslDes platform provider의 증거 수준을 하나의 evidence
contract로 정의한다.

## 할 일

- [x] Web source, Web rendered, native source, design-document computed, simulator/emulator, physical device evidence를 구분한다.
- [x] detector finding과 실행 실패를 각각 exit code 2와 1로 보존한다.
- [x] deterministic finding과 visual judgment를 별도 report field로 저장한다.
- [x] canonical rule ID와 Impeccable/Hallmark source catalog ID를 분리해 저장하고 중복 finding을 병합한다.
- [x] report에 effective rule pack id·version·fingerprint와 정렬된 active rule id exact set을 저장한다.
- [x] finding status와 별도로 rule activation을 `active`, `not-applicable`, `disabled`, `unsupported` exact enum으로 저장한다.
- [x] false-positive record는 exact owner fingerprint와 rationale 없이 생성하지 못하게 한다.
- [x] 미확보 browser/native/device evidence를 `not-run` 또는 `deferred`로 표시한다.
- [x] optional visual judgment의 유무가 deterministic verdict와 fingerprint를 변경하지 못하게 한다.

사람·LLM의 audit 해석과 refinement 검토는
[Optional TODO](../../../TODO_Optional.md)에서 관리한다.

## Consumer profile 경계

AnslDes는 `operate`, `read`, `browse`, `create` 같은 consumer profile의 schema와 판정 우선순위만
정의한다. 어떤 profile을 선택할지, 허용 type ratio와 identity가 무엇인지는 소비 저장소가 정책으로
주입하며 AnslDes 기본값에 특정 제품 유형을 고정하지 않는다.

## 완료 조건

- [x] 모든 anti-slop 결과가 evidence kind와 platform을 가진다.
- [x] Web detector가 native 품질 증거로 직렬화되는 경로가 없다.
- [x] `pass`, `fail`, `advisory`, `false-positive`, `not-run`, `deferred` 상태가 exact enum이다.
- [x] 같은 pack manifest와 activation policy가 같은 effective rule-set fingerprint를 만든다.
- [x] disabled·unsupported rule이 실행된 것처럼 `pass` finding을 만들지 않는다.

## 결정론적 코드 기준

- [snapshot provenance](../references/impeccable-detector-2026/README.md)
- [finding model](../references/impeccable-detector-2026/upstream/cli/engine/findings.mjs)
- [rule registry와 engine support](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
- [Hallmark source mapping](../references/hallmark-eight-tells-2026/README.md)
