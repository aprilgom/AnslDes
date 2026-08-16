# 09. Copy rules

> 이전: [Motion](./08-motion.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Imagery](./10-imagery.md)

## Rules

- [ ] `em-dash-overuse`: 영문 em dash saturation을 advisory로 기록한다.
- [ ] `marketing-buzzword`: 구체적 기능 없는 SaaS buzzword를 검출한다.
- [ ] `aphoristic-cadence`: 반복되는 manufactured contrast 문장을 검출한다.
- [ ] `repeated-container-text`: 한 container 안 동일 literal의 구조적 반복을 검출한다.
- [ ] `theater-slop-phrase`: “theater” framing과 한국어 동등 과장 표현을 검토한다.

## Consumer-policy 해석

- [ ] consumer policy가 제공한 content role, intent와 recovery-copy registry를 재사용한다.
- [ ] 한국어에서는 영문 punctuation regex를 그대로 이식하지 않고 locale-aware fixture를 만든다.
- [ ] 질문, 이유, 위험, 결과와 다음 행동이 서로 같은 내용을 반복하지 않는지 검사한다.
- [ ] 도메인·법적 용어를 마케팅 문구로 오인하지 않도록 source/rationale를 연결한다.

## 완료 조건

- [ ] 다섯 rule의 한국어 적용·비적용 기준이 예문과 함께 고정된다.
- [ ] advisory가 silent pass나 blocking error로 임의 변환되지 않는다.

## 결정론적 코드 기준

- [copy pattern 판정 구현](../references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)
- [rule severity·advisory registry](../references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
