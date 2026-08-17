# 14. Upstream snapshot과 Web provider

> 이전: [Native conformance](./13-native.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Native와 design-document provider](./15-native-pencil-provider.md)

## Upstream catalog와 snapshot

- [x] detector package를 exact version과 integrity로 고정한다.
- [x] 59개 rule id, category, scope와 advisory metadata를 machine registry로 저장한다.
- [x] vendored detector commit과 snapshot tree fingerprint를 저장한다.
- [x] Hallmark pinned commit의 source 8개와 canonical mapping exact set을 저장한다.
- [x] Hallmark와 Impeccable 중복 4개를 하나의 diagnostic으로 canonicalize한다.
- [x] canonical deterministic registry가 exact 63개인지 검사한다.
- [x] canonical 63개 rule을 exact member, pack version과 fingerprint를 가진 built-in pack manifest로 제공한다.
- [x] 각 rule이 id, implementation version, category, platform, evidence kind, default severity,
      provenance와 dependency를 가진 `RuleSpec`으로 등록되게 한다.
- [x] registry composer가 여러 pack을 stable ordering으로 병합하고 duplicate id, missing dependency,
      incompatible provider와 unknown member를 실패시킨다.
- [x] upstream update는 registry·implementation·mapping diff와 migration note 없이는 통과하지 못하게 한다.

## Web provider

- [x] source, static HTML, rendered browser와 visual-contrast engine을 별도 provider로 정의한다.
- [x] viewport, theme, font scale와 Reduce Motion axis를 consumer policy에서 입력받는다.
- [x] finding, advisory, false positive, not-run과 execution error를 분리한다.
- [x] generated framework CSS/vendor false positive는 exact artifact fingerprint로만 격리한다.
- [x] route와 build command는 consumer repository가 제공하고 AnslDes fixture에 고정하지 않는다.

## Neutral fixtures

- [x] 합성 HTML/CSS/TSX fixture에서 모든 지원 engine의 positive와 negative finding을 검증한다.
- [x] 합성 rule과 pack을 추가·제거·교체하는 fixture가 engine dispatch code를 수정하지 않고 통과한다.
- [x] parser dependency 누락으로 regex fallback이 된 결과를 full Web pass로 인정하지 않는다.
- [x] browser를 실행하지 못한 환경은 `not-run`이며 source-only pass와 구분한다.

## 완료 조건

- [x] registry drift 0, provider execution error 0이다.
- [x] manifest member exact set, implementation registry exact set과 report effective rule exact set이 일치한다.
- [x] false-positive exclusion이 artifact fingerprint와 재현 명령을 가진다.
- [x] 실제 소비 제품의 Web baseline과 report가 AnslDes 저장소에 포함되지 않는다.

구현 경계: `anti_slop_catalog.json`은 59개 Impeccable source와 Hallmark exact mapping에서 생성되며,
`ansldes-anti-slop@1.1.0`의 63개 member와 fingerprint를 고정한다. Web provider payload는 소비 정책의 exact
capture matrix에 대조되고 `completed/full`만 충족으로 인정된다. 이 저장소에는 중립 `example-*` fixture만
있으며 실제 route baseline, owner별 예외와 통과 report는 소비 저장소가 소유한다.

## 결정론적 코드 기준

- [pinned detector snapshot](../references/impeccable-detector-2026/README.md)
- [CLI entry](../references/impeccable-detector-2026/upstream/cli/engine/cli/main.mjs)
- [regex source engine](../references/impeccable-detector-2026/upstream/cli/engine/engines/regex/detect-text.mjs)
- [static HTML engine](../references/impeccable-detector-2026/upstream/cli/engine/engines/static-html/detect-html.mjs)
- [browser engine](../references/impeccable-detector-2026/upstream/cli/engine/engines/browser/detect-url.mjs)
- [Hallmark eight tells provenance와 mapping](../references/hallmark-eight-tells-2026/README.md)
