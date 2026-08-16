# 07. Layout과 space rules

> 이전: [Color](./06-color.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Motion](./08-motion.md)

## Rules

- [ ] `nested-cards`: card/surface 안의 반복 card depth를 검출한다.
- [ ] `monotonous-spacing`: 모든 관계에 같은 gap을 쓰는 평평한 rhythm을 검출한다.
- [ ] `numbered-section-labels`: 의미 없는 tiny section number scaffolding을 검출한다.
- [ ] `edge-flush-cards`: horizontal scroller의 비대칭 gutter를 검사한다.
- [ ] `text-occlusion`: opaque layer가 text bounds를 덮는지 검사한다.
- [ ] `first-viewport-column-overflow`: opening multi-column 높이 불균형을 검사한다.
- [ ] `heading-rhythm`: heading 위 간격이 아래 간격보다 작거나 같은지 검사한다.
- [ ] `line-length`: prose 65–75ch, 경고 상한 약 80ch를 검사한다.
- [ ] `cramped-padding`: bounded text container의 내부 여백을 검사한다.
- [ ] `body-text-viewport-edge`: viewport horizontal inset을 검사한다.
- [ ] `text-overflow`: container 밖 content와 의도하지 않은 horizontal scroll을 검사한다.
- [ ] `clipped-overflow-container`: positioned overlay가 clipping ancestor에 갇히는지 검사한다.

## Consumer-policy 해석

- [ ] data/table density와 prose line length를 consumer profile로 구분한다.
- [ ] semantic spacing 관계와 component anatomy를 사용해 rhythm을 판단한다.
- [ ] design-document computed overflow/overlap과 Web browser bounds를 독립 증거로 저장한다.
- [ ] card 제거 후 정보 경계는 typography, divider, spacing으로 유지되는지 확인한다.

## 완료 조건

- [ ] 12개 rule의 browser/native/design-document 적용 가능성 표가 완성된다.
- [ ] nested accessory tile과 반복 guide skeleton을 negative fixture로 검출한다.

## 결정론적 코드 기준

- [layout 판정 구현](../references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)
- [browser engine](../references/impeccable-detector-2026/upstream/cli/engine/engines/browser/detect-url.mjs)
- [static HTML engine](../references/impeccable-detector-2026/upstream/cli/engine/engines/static-html/detect-html.mjs)
