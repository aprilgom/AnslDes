# Hallmark eight tells reference

Hallmark 발표 자료의 “Ban the tells” 여덟 항목을 AnslDes의 제품 중립 rule registry에 연결하는
고정 provenance와 mapping이다. Hallmark는 skill 기반의 authored guidance이며 결정론적 detector가
아니다. 따라서 source 문구를 실행 결과로 가장하지 않고, AnslDes provider가 검증할 수 있는 predicate로
canonicalize한다.

## Provenance

- upstream: [Nutlope/hallmark](https://github.com/Nutlope/hallmark)
- commit: [`13ac0ec7e148655948100b6396439e481361d690`](https://github.com/Nutlope/hallmark/commit/13ac0ec7e148655948100b6396439e481361d690)
- commit date: `2026-08-06T09:19:18-07:00`
- snapshot date: `2026-08-16`
- canonical source: [`docs/talk-slides.md`의 “Ban the tells”](https://github.com/Nutlope/hallmark/blob/13ac0ec7e148655948100b6396439e481361d690/docs/talk-slides.md#L156-L171)
- source rule count: 8
- source file SHA-256: `83083b5e37b99cb8211268b778aa0cb3f677f8fe53f8d817f54fcad799882e0e`
- license: [MIT](https://github.com/Nutlope/hallmark/blob/13ac0ec7e148655948100b6396439e481361d690/LICENSE)

## Canonical mapping

| Source ID | Hallmark tell | AnslDes canonical rule | 처리 |
| --- | --- | --- | --- |
| `hallmark-eight-01` | purple-to-blue gradient hero | `ai-color-palette` | Impeccable rule에 provenance 병합 |
| `hallmark-eight-02` | Inter/Roboto를 display와 body에 함께 사용 | `overused-font` | role 결합 predicate를 보존하고 provenance 병합 |
| `hallmark-eight-03` | icon 위주의 동일한 3열 feature layout | `equal-icon-feature-columns` | unique supplement |
| `hallmark-eight-04` | `100vh` 중심 정렬 hero | `full-viewport-centered-hero` | unique supplement |
| `hallmark-eight-05` | 의미 없는 card nesting | `nested-cards` | Impeccable rule에 provenance 병합 |
| `hallmark-eight-06` | headline의 clipped gradient text | `gradient-text` | Impeccable rule에 provenance 병합 |
| `hallmark-eight-07` | 순수 black/white surface | `pure-extreme-surface` | unique supplement |
| `hallmark-eight-08` | 근거 없는 metric, testimonial 또는 logo | `unverified-social-proof` | unique supplement, consumer provenance 필요 |

8개 source entry는 exact-set으로 유지한다. 기존 Impeccable rule 4개와 겹치므로 canonical registry에는
새 rule 4개만 더한다. source ID와 canonical rule ID를 둘 다 report provenance에 저장하되 같은 node와
predicate에 대해 diagnostic을 두 번 생성하지 않는다.

## Evidence boundary

- `equal-icon-feature-columns`는 동일한 세 column, 반복되는 icon-above-heading anatomy와 page-level
  feature region이 함께 확인될 때만 finding을 만든다. 단순 3열 data grid는 대상이 아니다.
- `full-viewport-centered-hero`는 full viewport minimum height와 수직·수평 중심 정렬의 결합을 검사한다.
  viewport 단위 자체를 금지하지 않는다.
- `pure-extreme-surface`는 rendered surface 또는 surface token의 순수 black/white를 검사한다. text,
  monochrome asset와 소비 정책이 지정한 매체 예외는 별도 owner evidence를 요구한다.
- `unverified-social-proof`는 내용의 거짓 여부를 추론하지 않는다. metric, testimonial, logo claim에
  consumer content registry의 source reference가 없음을 보고한다. registry를 실행하지 못한 경우는
  `not-run`이며 pass가 아니다.

## Update policy

Hallmark update는 commit, source file hash, exact source rule count와 mapping diff를 함께 기록한다.
Hallmark 전체 57개 slop-test gate나 8-state component checklist는 이 eight tells catalog와 다른
source set이므로 검토 없이 자동 편입하지 않는다.
