# 14. Upstream snapshot과 Web provider

> 이전: [Native conformance](./13-native.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Native와 design-document provider](./15-native-pencil-provider.md)

## Upstream catalog와 snapshot

- [ ] detector package를 exact version과 integrity로 고정한다.
- [ ] 59개 rule id, category, scope와 advisory metadata를 machine registry로 저장한다.
- [ ] vendored detector commit과 snapshot tree fingerprint를 저장한다.
- [ ] 공식 LLM-only 5개 이름, URL과 snapshot date를 저장한다.
- [ ] Hallmark pinned commit의 source 8개와 canonical mapping exact set을 저장한다.
- [ ] Hallmark와 Impeccable 중복 4개를 하나의 diagnostic으로 canonicalize한다.
- [ ] canonical deterministic registry가 exact 63개인지 검사한다.
- [ ] upstream update는 registry·implementation·mapping diff와 migration note 없이는 통과하지 못하게 한다.

## Web provider

- [ ] source, static HTML, rendered browser와 visual-contrast engine을 별도 provider로 정의한다.
- [ ] viewport, theme, font scale와 Reduce Motion axis를 consumer policy에서 입력받는다.
- [ ] finding, advisory, false positive, not-run과 execution error를 분리한다.
- [ ] generated framework CSS/vendor false positive는 exact artifact fingerprint로만 격리한다.
- [ ] route와 build command는 consumer repository가 제공하고 AnslDes fixture에 고정하지 않는다.

## Neutral fixtures

- [ ] 합성 HTML/CSS/TSX fixture에서 모든 지원 engine의 positive와 negative finding을 검증한다.
- [ ] parser dependency 누락으로 regex fallback이 된 결과를 full Web pass로 인정하지 않는다.
- [ ] browser를 실행하지 못한 환경은 `not-run`이며 source-only pass와 구분한다.

## 완료 조건

- [ ] registry drift 0, provider execution error 0이다.
- [ ] false-positive exclusion이 artifact fingerprint와 재현 명령을 가진다.
- [ ] 실제 소비 제품의 Web baseline과 report가 AnslDes 저장소에 포함되지 않는다.

## 결정론적 코드 기준

- [pinned detector snapshot](../references/impeccable-detector-2026/README.md)
- [CLI entry](../references/impeccable-detector-2026/upstream/cli/engine/cli/main.mjs)
- [regex source engine](../references/impeccable-detector-2026/upstream/cli/engine/engines/regex/detect-text.mjs)
- [static HTML engine](../references/impeccable-detector-2026/upstream/cli/engine/engines/static-html/detect-html.mjs)
- [browser engine](../references/impeccable-detector-2026/upstream/cli/engine/engines/browser/detect-url.mjs)
- [Hallmark eight tells provenance와 mapping](../references/hallmark-eight-tells-2026/README.md)
