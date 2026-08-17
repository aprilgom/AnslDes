# 17. AnslDes gate 통합과 완료 감사

> 이전: [Governance](./16-governance.md) · [Anti-slop 인덱스](../TODO.md) · [Repository TODO](../../../TODO.md)

## AnslDes 통합

- [x] source, rendered Web와 native/design-document provider evidence stage를 추가한다.
- [x] blocking/advisory/exception/not-run/deferred budget schema를 추가한다.
- [x] vendored detector snapshot, Hallmark source fingerprint와 anti-slop contract를 release dependency fingerprint에 포함한다.
- [x] effective rule pack·rule exact set, activation reason과 registry fingerprint를 text, JSON과 SARIF report에 포함한다.
- [x] anti-slop stage가 빠지거나 report가 stale하면 실패하는 중립 fixture를 추가한다.
- [x] 실패를 rule ID, engine, platform, viewport와 injected owner에 귀속한다.
- [x] `deslint` text, JSON과 SARIF가 같은 정렬·severity·location을 출력한다.

## 완료 감사

- [x] Impeccable source registry exact 59, Hallmark source catalog exact 8, 중복 제거 후 canonical deterministic registry exact 63이다.
- [x] Web source/rendered와 native/design-document evidence가 report에서 분리된다.
- [x] neutral fixture의 blocking finding, execution error와 expired exception이 모두 0이다.
- [x] neutral fixture의 advisory는 0이며 advisory budget을 별도로 감사한다.
- [x] raw slop과 config/report 우회 fixture가 정확히 실패한다.
- [x] rule pack 추가·제거·교체와 policy activation fixture가 local/CI에서 같은 report를 만든다.
- [x] AnslDes 공통 `npm run check`가 anti-slop contract와 deslint rule test를 포함한다.

## 소비 저장소 경계

- [x] AnslDes release는 schema, registry, provider와 neutral fixture만 포함한다.
- [x] 소비 저장소가 versioned release와 checksum을 exact pin하는 절차를 문서화한다.
- [x] 실제 route, screen, token, owner, exception과 runtime report가 generic release에 포함되지 않는다.
- [x] 제품 통합 실패가 generic passing report를 수정하지 못한다.

## 문서 상태

- [x] 16개 필수 단계 문서가 모두 완료되어 Anti-slop 인덱스를 `[x]`로 바꾼다.
- [x] README, deslint architecture, changelog와 release migration 문서를 갱신한다.
- [x] rule authoring, pack publishing, deprecation과 removal lifecycle을 문서화한다.

## 완료 조건

- [x] local/CI에서 같은 중립 fixture가 같은 report를 만든다.
- [x] 같은 registry·manifest·policy fingerprint에서 effective rule exact set과 ordering이 같다.
- [x] 확보하지 않은 browser/native/device evidence가 완료로 과장되지 않는다.
- [x] product-boundary check가 제품명·경로·브랜드 값 유입을 실패시킨다.

## 구현 근거

- 예산·governance 계약: `design-system-policy.schema.json`, `internal/policy`, `internal/governance`
- provider와 stage 분리: `internal/webcheck`, `internal/nativecheck`, layout/runtime evidence analyzer
- canonical report: `internal/report`, `deslint-report.schema.json`
- provenance와 release pin: `anti_slop_catalog.json`, `ansldes-release.json`, `consumer-release-lock.schema.json`
- ordering 회귀: `TestRunnerProducesTheSameNeutralMultiStageReportForAnyInputOrder`
- lifecycle 회귀: `internal/rules/evolution_test.go`, registry/composer tests

## 결정론적 코드 기준

- [pinned detector snapshot](../references/impeccable-detector-2026/README.md)
- [CLI entry](../references/impeccable-detector-2026/upstream/cli/engine/cli/main.mjs)
- [rule registry](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
- [Hallmark eight tells mapping](../references/hallmark-eight-tells-2026/README.md)
