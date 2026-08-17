# AnslDes Optional TODO

> [Required deterministic TODO](./TODO.md) · [Anti-slop gate](./docs/anti-slop/TODO.md)

같은 versioned input과 provider evidence에서 같은 결과를 재현할 수 없는 사람·LLM의 정성적 판단을
관리한다. 이 문서의 항목은 선택 사항이며 AnslDes release, `npm run check` 또는 deterministic
Anti-slop Gate의 완료 조건이 아니다.

Optional 결과는 deterministic finding과 별도 report field에 저장한다. 미검토 결과는
`not-reviewed`이며 `pass`로 승격하지 않고, optional review의 부재가 deterministic gate를
실패시키지도 않는다.

## 1. Qualitative audit record

- [ ] audit finding마다 P0–P3, 위치, 사용자 영향, 기준, 수정 방향과 reviewer를 기록한다.
- [ ] systemic issue와 isolated defect를 구분하고 positive observation도 보존한다.
- [ ] 명시된 brief, 기존 identity와 native affordance를 정성적 판단의 문맥으로 사용한다.
- [ ] `DESIGN.md` 부재를 greenfield나 visual authority 부재로 해석하지 않는다.
- [ ] refinement가 사실 문구·행동·범위 밖 identity를 바꾸지 않는지 검토한다.

## 2. Consumer UX review

- [ ] overlay가 첫 해결책인지 inline 또는 progressive disclosure가 가능한지 검토한다.
- [ ] 표준 navigation, 익숙한 control과 화면 간 반복 일관성을 positive observation으로 기록한다.

## 3. LLM-only critique

공식 Impeccable catalog의 deterministic detector 밖 판단 다섯 개를 optional review evidence로
관리한다.

- [ ] `Glassmorphism everywhere`: 실제 layer/material 문제를 해결하지 않는 blur/glow 남용을 검토한다.
- [ ] `Extreme border-radius on cards`: card 12–16px 기준과 pill/control 경계를 검토한다.
- [ ] `Amateurish hand-drawn SVG`: pictorial SVG의 craft와 asset medium 적합성을 검토한다.
- [ ] `Hero metric layout`: 근거 없는 big-number/stat template를 검토한다.
- [ ] `Identical card grids`: 동일 icon+heading+text card 반복이 정보 구조를 대신하는지 검토한다.
- [ ] 화면/root/node id, reviewer, date, before/after 또는 screenshot을 기록한다.
- [ ] 판단의 사용자 영향과 consumer-specific 대안을 기록한다.
- [ ] radius contract, asset registry, repeated structure처럼 결정론화 가능한 보조 규칙을 연결한다.
- [ ] `Hero metric layout`의 시각적 적합성 판단과 `unverified-social-proof`의 출처 검증을 구분한다.
- [ ] “LLM이 문제없다고 했다”는 단독 완료 증거로 인정하지 않는다.
- [ ] P3 취향 finding을 과다 생성하지 않고 systemic P1/P2를 우선한다.
- [ ] exact 5개 review record가 있고 `pass`, `fail`, `not-reviewed`가 명확하다.
- [ ] 미검토 judgment가 deterministic gate의 통과 finding으로 직렬화되지 않는다.

## 4. Native qualitative review

- [ ] Accessibility, Performance, Appearance/Theming, Platform Conformance, Adaptivity를 0–4로 평가한다.
- [ ] Web-shaped control, gratuitous motion, inconsistent affordance와 off-platform overlay를 검토한다.
- [ ] Android Dynamic Color 적용 가능성과 static fallback의 적합성을 검토한다.
- [ ] iOS의 hand-rolled glassmorphism, bespoke card stack과 off-platform iconography를 검토한다.
- [ ] Android의 Material component/motion과 iconography에서 iOS-shaped drift를 검토한다.
- [ ] 다섯 audit dimension의 합계 20점, rating band와 platform-conformance verdict를 기록한다.

## 5. Optional evidence integration

- [ ] 공식 LLM-only 5개 이름, URL과 snapshot date를 optional provenance registry에 저장한다.
- [ ] LLM review evidence stage는 명시적으로 요청된 경우에만 실행한다.
- [ ] optional review report가 deterministic report fingerprint와 pass/fail을 변경하지 않는지 검사한다.
- [ ] optional review 미실행을 `not-reviewed`로 기록하고 deterministic evidence의 `not-run`과 구분한다.

## 결정론적 경계

- [59개 deterministic registry](./docs/anti-slop/references/impeccable-detector-2026/upstream/cli/engine/registry/antipatterns.mjs)
- [snapshot usage boundary](./docs/anti-slop/references/impeccable-detector-2026/README.md#usage-boundary)
- [Hallmark source mapping](./docs/anti-slop/references/hallmark-eight-tells-2026/README.md)

이 문서의 결과는 detector 통과 결과를 대체하거나 필수 gate를 약화하지 않는다.
