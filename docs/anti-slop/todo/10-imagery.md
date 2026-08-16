# 10. Imagery rules

> 이전: [Copy](./09-copy.md) · [Anti-slop 인덱스](../TODO.md) · 다음: [Runtime](./11-runtime.md)

## Rules

- [ ] `shape-assembled-illustration`: hero-sized generic primitive SVG scene을 검출한다.
- [ ] `broken-image`: missing/empty/placeholder `src`와 asset load failure를 검출한다.

## Consumer-policy 해석

- [ ] icon, logo, data diagram과 pictorial hero illustration을 size/role/geometry로 구분한다.
- [ ] asset registry의 implementation source, consumer와 fingerprint를 재사용한다.
- [ ] Web `img`/video poster와 native Image source를 모두 검사한다.
- [ ] decorative asset의 screen-reader exclusion과 functional image label을 함께 검증한다.
- [ ] 실제 asset이 필요한 영역을 gradient/shape placeholder로 대체하지 않는다.

## 완료 조건

- [ ] 두 rule의 Web/native fixture와 asset owner drift 검사가 통과한다.
- [ ] broken source와 intentionally omitted illustration을 구분한다.

## 결정론적 코드 기준

- [image 관련 판정 구현](../references/impeccable-detector-2026/upstream/cli/engine/rules/checks.mjs)
- [browser engine](../references/impeccable-detector-2026/upstream/cli/engine/engines/browser/detect-url.mjs)
