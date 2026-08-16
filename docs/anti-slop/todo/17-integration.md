# 17. AnslDes gate 통합과 완료 감사

> 이전: [Governance](./16-governance.md) · [Anti-slop 인덱스](../TODO.md) · [Repository TODO](../../../TODO.md)

## AnslDes 통합

- [ ] source, rendered Web, native/design-document provider와 LLM review evidence stage를 추가한다.
- [ ] blocking/advisory/exception/not-run/deferred budget schema를 추가한다.
- [ ] vendored detector snapshot과 anti-slop contract를 release dependency fingerprint에 포함한다.
- [ ] anti-slop stage가 빠지거나 report가 stale하면 실패하는 중립 fixture를 추가한다.
- [ ] 실패를 rule id, engine, platform, viewport와 injected owner에 귀속한다.
- [ ] `deslint` text, JSON과 SARIF가 같은 정렬·severity·location을 출력한다.

## 완료 감사

- [ ] deterministic registry exact 59, LLM judgment exact 5다.
- [ ] Web source/rendered와 native/design-document evidence가 report에서 분리된다.
- [ ] neutral fixtures의 blocking finding 0, execution error 0, expired exception 0이다.
- [ ] advisory는 0 또는 승인 review record를 가진다.
- [ ] LLM-only 미검증 항목을 pass로 표시하지 않는다.
- [ ] raw slop와 config/report 우회 fixture가 정확히 실패한다.
- [ ] AnslDes 공통 `npm run check`가 anti-slop contract와 deslint rule test를 포함한다.

## 소비 저장소 경계

- [ ] AnslDes release는 schema, registry, provider와 neutral fixture만 포함한다.
- [ ] 소비 저장소가 versioned release와 checksum을 exact pin하는 절차를 문서화한다.
- [ ] 실제 route, screen, token, owner, exception과 runtime report가 generic release에 포함되지 않는다.
- [ ] 제품 통합 실패가 generic passing report를 수정하지 못한다.

## 문서 상태

- [ ] 17개 세부 문서가 모두 완료된 후 Anti-slop 인덱스를 `[x]`로 바꾼다.
- [ ] README, deslint architecture, changelog와 release migration 문서를 갱신한다.

## 완료 조건

- [ ] local/CI에서 같은 중립 fixture가 같은 report를 만든다.
- [ ] 확보하지 않은 browser/native/device evidence가 완료로 과장되지 않는다.
- [ ] product-boundary check가 제품명·경로·브랜드 값 유입을 실패시킨다.

## 결정론적 코드 기준

- [pinned detector snapshot](../references/impeccable-detector-2026/README.md)
- [CLI entry](../references/impeccable-detector-2026/upstream/cli/engine/cli/main.mjs)
- [rule registry](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
