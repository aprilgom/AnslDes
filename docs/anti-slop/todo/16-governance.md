# 16. Exception과 governance

> 이전: [Native와 design-document provider](./15-native-pencil-provider.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [통합과 완료](./17-integration.md)

## Exception

- [x] `ruleId`, engine, platform, exact owner, reason, reviewer, expiry/review trigger를 요구한다.
- [x] `ignoreRules`, `ignoreFiles`, `ignoreValues`와 inline ignore를 exact allowlist로 검사한다.
- [x] broad glob, wildcard rule과 owner 없는 value ignore를 거부한다.
- [x] expired exception과 source owner drift를 실패시킨다.
- [x] rule `disabled` record는 exact pack/rule id, owner, rationale, reviewer와 expiry/review trigger를 요구한다.
- [x] `not-applicable`, `disabled`, `unsupported`를 서로 대체하거나 finding 삭제에 사용하지 못하게 한다.

## 우회 방지

- [x] `--no-config`, `--no-design-system`, `--no-inline-ignores` 사용을 CI에서 금지한다.
- [x] `--no-advisory` 정책을 contract에 고정한다.
- [x] exit code 2를 성공으로 바꾸는 wrapper를 fixture로 실패시킨다.
- [x] JSON finding 삭제, count 재작성과 passing report 강제 저장을 실패시킨다.
- [x] 한 provider의 ignore가 다른 provider mapping까지 면제하지 못하게 한다.
- [x] policy에서 required pack이나 rule을 누락·변조해 gate 범위를 축소하는 fixture를 실패시킨다.

## 운영

- [x] Impeccable version, Hallmark pinned commit, rule/mapping drift와 exception을 90일 governance review에 포함한다.
- [x] 신규 rule은 owner와 migration plan을 지정한 후 활성화한다.
- [x] 기존 rule 제거·교체는 pack major version, tombstone 또는 replacement id와 migration note를 요구한다.
- [x] additive rule은 pack minor version으로 추가하고 기존 consumer에 미치는 default activation을 명시한다.

## 완료 조건

- [x] expired exception 0, broad ignore 0, unowned exception 0이다.
- [x] duplicate/unknown rule, stale pack fingerprint와 승인 없는 disable 0이다.
- [x] config/report 우회 negative fixture가 정확한 diagnostic으로 실패한다.

구현 경계: exception과 ignore는 provider를 포함한 exact tuple로만 matching되며 만료되거나 owner가 달라지면
원 finding이 그대로 남는다. Governance policy는 90일 review subject, 네 개 금지 flag, advisory 보존,
exit-code/report 불변성과 passing-report-only 저장을 고정한다. `internal/governance`의 typed violation code는
wrapper exit rewrite, JSON mutation, count rewrite와 forced storage를 각각 구분한다. Pack evolution 검사는
exact member diff와 semver, owner, migration plan, tombstone/replacement 및 default activation을 함께 검증한다.

## 결정론적 코드 기준

- [snapshot provenance와 update policy](../references/impeccable-detector-2026/README.md)
- [inline ignore 구현](../references/impeccable-detector-2026/upstream/cli/engine/shared/inline-ignores.mjs)
- [CLI config·exit behavior](../references/impeccable-detector-2026/upstream/cli/engine/cli/main.mjs)
- [Hallmark provenance와 update policy](../references/hallmark-eight-tells-2026/README.md)
