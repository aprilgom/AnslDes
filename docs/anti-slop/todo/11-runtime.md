# 11. Runtime과 general quality rules

> 이전: [Imagery](./10-imagery.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Native conformance](./13-native.md)

## Rules

- [x] `script-error`: Web load 중 uncaught error와 parse failure를 수집한다.
- [x] `content-hidden-at-rest`: reveal 종료 후 주요 content가 hidden 상태인지 검사한다.
- [x] `justified-text`: Web justify와 native text-justify equivalent를 검사한다.

## 구현

- [x] console error, unhandled rejection과 route render failure를 route별로 기록한다.
- [x] animation/reveal script가 실패해도 content가 default-visible인지 검사한다.
- [x] browser runtime failure와 detector process failure를 구분한다.
- [x] native에서는 app crash, redbox와 render error boundary evidence mapping을 별도로 정의한다.
- [x] body copy justify 금지와 print/export exception을 exact owner로 관리한다.

## 완료 조건

- [x] 세 rule의 runtime fixture가 정확한 route와 owner를 보고한다.
- [x] Web 성공만으로 native runtime을 완료 처리하지 않는다.

## 구현 경계

- `runtime-evidence.schema.json`은 completed capture와 detector-process failure를 상호 배타적으로 직렬화한다.
- Web은 console/Promise/route/uncaught/parse failure만, native는 app crash/redbox/render error boundary만 받는다.
- `runtime/script-error`, `runtime/content-hidden-at-rest`, `runtime/justified-text`는 공통 `RuleSpec` registry의
  `runtime` input으로 등록되어 evaluator 추가·교체가 engine 분기 변경을 요구하지 않는다.
- Print/export justify는 consumer policy `runtime.registryVersion`과 exact permission tuple이 일치해야 한다.

## 결정론적 코드 기준

- [runtime 판정 구현](../references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)
- [browser engine](../references/impeccable-detector-2026/upstream/cli/engine/engines/browser/detect-url.mjs)
- [browser injected detector](../references/impeccable-detector-2026/upstream/cli/engine/browser/injected/index.mjs)
