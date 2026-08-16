# 11. Runtime과 general quality rules

> 이전: [Imagery](./10-imagery.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [LLM-only critique](./12-llm-review.md)

## Rules

- [ ] `script-error`: Web load 중 uncaught error와 parse failure를 수집한다.
- [ ] `content-hidden-at-rest`: reveal 종료 후 주요 content가 hidden 상태인지 검사한다.
- [ ] `justified-text`: Web justify와 native text-justify equivalent를 검사한다.

## 구현

- [ ] console error, unhandled rejection과 route render failure를 route별로 기록한다.
- [ ] animation/reveal script가 실패해도 content가 default-visible인지 검사한다.
- [ ] browser runtime failure와 detector process failure를 구분한다.
- [ ] native에서는 app crash, redbox와 render error boundary evidence mapping을 별도로 정의한다.
- [ ] body copy justify 금지와 print/export exception을 exact owner로 관리한다.

## 완료 조건

- [ ] 세 rule의 runtime fixture가 정확한 route와 owner를 보고한다.
- [ ] Web 성공만으로 native runtime을 완료 처리하지 않는다.

## 결정론적 코드 기준

- [runtime 판정 구현](../references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)
- [browser engine](../references/impeccable-detector-2026/upstream/cli/engine/engines/browser/detect-url.mjs)
- [browser injected detector](../references/impeccable-detector-2026/upstream/cli/engine/browser/injected/index.mjs)
