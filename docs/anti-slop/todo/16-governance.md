# 16. Exception과 governance

> 이전: [Native와 design-document provider](./15-native-pencil-provider.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [통합과 완료](./17-integration.md)

## Exception

- [ ] `ruleId`, engine, platform, exact owner, reason, reviewer, expiry/review trigger를 요구한다.
- [ ] `ignoreRules`, `ignoreFiles`, `ignoreValues`와 inline ignore를 exact allowlist로 검사한다.
- [ ] broad glob, wildcard rule과 owner 없는 value ignore를 거부한다.
- [ ] expired exception과 source owner drift를 실패시킨다.

## 우회 방지

- [ ] `--no-config`, `--no-design-system`, `--no-inline-ignores` 사용을 CI에서 금지한다.
- [ ] `--no-advisory` 정책을 contract에 고정한다.
- [ ] exit code 2를 성공으로 바꾸는 wrapper를 fixture로 실패시킨다.
- [ ] JSON finding 삭제, count 재작성과 passing report 강제 저장을 실패시킨다.
- [ ] 한 provider의 ignore가 다른 provider mapping까지 면제하지 못하게 한다.

## 운영

- [ ] Impeccable version, rule drift와 exception을 90일 governance review에 포함한다.
- [ ] 신규 rule은 owner와 migration plan을 지정한 후 활성화한다.

## 완료 조건

- [ ] expired exception 0, broad ignore 0, unowned exception 0이다.
- [ ] config/report 우회 negative fixture가 정확한 diagnostic으로 실패한다.

## 결정론적 코드 기준

- [snapshot provenance와 update policy](../references/impeccable-detector-2026/README.md)
- [inline ignore 구현](../references/impeccable-detector-2026/upstream/cli/engine/shared/inline-ignores.mjs)
- [CLI config·exit behavior](../references/impeccable-detector-2026/upstream/cli/engine/cli/main.mjs)
