# Anti-slop Gate — Impeccable 2026

> [AnslDes README](../../README.md) · [Repository TODO](../../TODO.md) · [Product boundary](../product-boundary.md)

Impeccable 2026 catalog와 고정된 결정론적 detector source를 AnslDes의 제품 중립 lint·evidence
계약으로 수용한다. AnslDes는 rule 의미, evidence schema, platform provider interface, diagnostic과
governance를 소유한다. 실제 화면, token 값, 브랜드 identity, source owner·consumer, exception과
통과 보고서는 각 소비 저장소가 소유한다.

## 기준

- catalog 기준일: `2026-08-16`
- detector package: `impeccable@3.6.0`, deterministic rule 59개
- vendored detector commit: `7b646bafd60b9dd9828ce5c4c1a25691702c9e92`
- local source: [Impeccable deterministic detector snapshot](./references/impeccable-detector-2026/README.md)
- 공식 catalog: [Impeccable Slop](https://impeccable.style/slop/)
- 공식 detector: [Detector CLI](https://impeccable.style/docs/detector/)
- 공식 source: [pbakaus/impeccable](https://github.com/pbakaus/impeccable)

결정론적 기준의 source of truth는 vendored
[rule registry](./references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)와
[rule implementation](./references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)이다.
Web detector 결과는 React Native, Pencil 또는 실기기 runtime 증거의 대체물이 아니다.

## 실행 순서

- [ ] [01. Evidence contract와 audit 경계](./todo/01-evidence.md)
- [ ] [02. Consumer conformance profile](./todo/02-consumer-profile.md)
- [ ] [03. Design-system awareness rules](./todo/03-design-system-awareness.md)
- [ ] [04. Visual detail rules](./todo/04-visual-detail.md)
- [ ] [05. Typography와 hierarchy rules](./todo/05-typography.md)
- [ ] [06. Color와 contrast rules](./todo/06-color.md)
- [ ] [07. Layout과 space rules](./todo/07-layout.md)
- [ ] [08. Motion rules](./todo/08-motion.md)
- [ ] [09. Copy rules](./todo/09-copy.md)
- [ ] [10. Imagery rules](./todo/10-imagery.md)
- [ ] [11. Runtime과 general quality rules](./todo/11-runtime.md)
- [ ] [12. LLM-only critique](./todo/12-llm-review.md)
- [ ] [13. Native platform conformance](./todo/13-native.md)
- [ ] [14. Upstream snapshot과 Web provider](./todo/14-web-gate.md)
- [ ] [15. React Native와 Pencil provider](./todo/15-native-pencil-provider.md)
- [ ] [16. Exception과 governance](./todo/16-governance.md)
- [ ] [17. AnslDes gate 통합과 완료 감사](./todo/17-integration.md)

각 단계는 deterministic finding, false positive, advisory, LLM judgment와 미확보 runtime evidence를
서로 다른 상태로 기록한다. 미검증을 통과로 표시하지 않으며 broad ignore나 report 후처리로 finding을
숨기지 않는다.

## 제품 중립 경계

- AnslDes fixture는 `example-product` 같은 중립 identity와 합성 token만 사용한다.
- 제품 모드, locale, platform, owner와 예외는 versioned consumer policy로 주입한다.
- 소비 저장소의 route 수, token 값, component 소비 횟수와 baseline finding을 복사하지 않는다.
- 실제 제품 통과 보고서와 rollback evidence는 소비 저장소에만 저장한다.

## 연결 문서

- [AnslDes architecture](../architecture.md)
- [Product boundary](../product-boundary.md)
- [deslint architecture](../../deslint/docs/architecture.md)
- [Pinned detector snapshot](./references/impeccable-detector-2026/README.md)
