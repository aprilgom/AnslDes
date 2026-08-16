# Migration boundary

AnslDes는 제품별 기존 linter, source snapshot, UI 문서나 parity report를 저장하지 않는다. 이 자료는
소비 저장소가 소유하고, AnslDes release는 제품을 알지 못한 채 언어 중립 definition과 policy를 처리한다.

## AnslDes가 제공하는 것

- versioned definition/policy schema
- 결정론적 compiler와 runtime
- TS/TSX, Pencil과 computed-layout evidence adapter
- stable rule ID, diagnostic ordering과 report format
- 중립 positive, negative, shadow, spread와 false-positive fixture

## 소비 저장소가 제공하는 것

- 제품 token 값과 theme mapping
- source owner, consumer path와 exact count
- 브랜드 asset과 승인 예외
- 기존 linter fixture와 dual-run parity report
- migration note, rollback release와 cutover 승인

## Cutover rule

소비 저장소는 같은 source snapshot에서 기존 linter와 pinned `deslint` release를 실행한다. rule ID,
severity, location, message와 exit status가 승인된 parity 조건을 만족하고 rollback release가 존재하기 전에는
기존 구현을 제거하지 않는다. 이 evidence는 AnslDes로 복사하지 않는다.
