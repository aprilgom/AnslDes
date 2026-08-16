# 12. LLM-only critique

> 이전: [Runtime](./11-runtime.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Native conformance](./13-native.md)

공식 catalog의 deterministic detector 밖 판단 다섯 개를 review evidence로 관리한다.

## Judgments

- [ ] `Glassmorphism everywhere`: 실제 layer/material 문제를 해결하지 않는 blur/glow 남용을 검토한다.
- [ ] `Extreme border-radius on cards`: card 12–16px 기준과 pill/control 경계를 검토한다.
- [ ] `Amateurish hand-drawn SVG`: pictorial SVG의 craft와 asset medium 적합성을 검토한다.
- [ ] `Hero metric layout`: 근거 없는 big-number/stat template를 검토한다.
- [ ] `Identical card grids`: 동일 icon+heading+text card 반복이 정보 구조를 대신하는지 검토한다.

## Evidence

- [ ] 화면/root/node id, reviewer, date, before/after 또는 screenshot을 기록한다.
- [ ] 판단의 사용자 영향과 consumer-specific 대안을 기록한다.
- [ ] radius contract, asset registry, repeated structure처럼 결정론화 가능한 보조 규칙을 연결한다.
- [ ] `Hero metric layout`의 시각적 적합성 판단과 `unverified-social-proof`의 출처 검증을 구분한다.
- [ ] “LLM이 문제없다고 했다”는 단독 완료 증거로 인정하지 않는다.
- [ ] P3 취향 finding을 과다 생성하지 않고 systemic P1/P2를 우선한다.

## 완료 조건

- [ ] exact 5개 review record가 있고 `pass/fail/not-reviewed`가 명확하다.
- [ ] 미검토 judgment가 전체 gate의 통과 finding으로 직렬화되지 않는다.

## 결정론적 코드 경계

- [59개 deterministic registry](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
- [snapshot usage boundary](../references/impeccable-detector-2026/README.md#usage-boundary)
- [Hallmark source mapping](../references/hallmark-eight-tells-2026/README.md)

이 문서의 다섯 판단은 registry 밖에 있으므로 detector 통과 결과로 대체하지 않는다.
