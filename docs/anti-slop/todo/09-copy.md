# 09. Copy rules

> 이전: [Motion](./08-motion.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Imagery](./10-imagery.md)

## Rules

- [x] `em-dash-overuse`: 영문 em dash saturation을 advisory로 기록한다.
- [x] `marketing-buzzword`: 구체적 기능 없는 SaaS buzzword를 검출한다.
- [x] `aphoristic-cadence`: 반복되는 manufactured contrast 문장을 검출한다.
- [x] `repeated-container-text`: 한 container 안 동일 literal의 구조적 반복을 검출한다.
- [x] `theater-slop-phrase`: versioned locale phrase registry와 일치하는 “theater” framing을 검출한다.
- [x] `unverified-social-proof`: metric, testimonial과 logo claim의 source reference 누락을 검출한다.

## Consumer-policy 해석

- [x] consumer policy가 제공한 content role, intent와 recovery-copy registry를 재사용한다.
- [x] 한국어에서는 영문 punctuation regex를 그대로 이식하지 않고 locale-aware fixture를 만든다.
- [x] 질문, 이유, 위험, 결과와 다음 행동이 서로 같은 내용을 반복하지 않는지 검사한다.
- [x] 도메인·법적 용어를 마케팅 문구로 오인하지 않도록 source/rationale를 연결한다.
- [x] claim truth를 lint가 추론하지 않고 consumer content registry의 stable source reference만 검증한다.

## 완료 조건

- [x] 여섯 rule의 한국어 적용·비적용 기준이 예문과 함께 고정된다.
- [x] advisory가 silent pass나 blocking error로 임의 변환되지 않는다.
- [x] content registry 미실행은 `not-run`이며 proof가 없는 claim을 pass로 표시하지 않는다.

## 한국어 적용 경계

| Rule | 적용 예 | 비적용 예 |
| --- | --- | --- |
| `em-dash-overuse` | `en-*` body에서 8회 이상·약 500자당 1회 이상 | `ko-KR` 문장부호에는 영문 regex 미적용 |
| `marketing-buzzword` | “혁신적인 경험”에 feature/rationale reference 없음 | 근거가 연결된 도메인·법적 용어 |
| `aphoristic-cadence` | 같은 container에 “X가 아닙니다. Y입니다.” 3회 이상 | 단발성 대비 문장 |
| `repeated-container-text` | 질문과 다음 행동이 모두 “다시 시도하세요” | approved recovery-copy ID가 같은 반복 |
| `theater-slop-phrase` | locale registry의 “보안 연극”과 일치 | registry에 없는 일반 설명 |
| `unverified-social-proof` | 출처 ID가 없는 사용자 수·추천사·logo claim | registry의 stable source reference와 exact match |

## 결정론적 코드 기준

- [copy pattern 판정 구현](../references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)
- [rule severity·advisory registry](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
- [Hallmark eight tells mapping](../references/hallmark-eight-tells-2026/README.md)
