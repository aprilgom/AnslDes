# Product Boundary

## Keep in AnslDes

- definition schema와 validator
- reference resolver와 compiler
- component anatomy와 accessibility 기본 동작
- generic lint engine와 evidence model
- generic fixture와 portability test

## Keep in each product

- primitive와 semantic token의 실제 값
- component variant의 시각 recipe
- 브랜드 font, icon, illustration과 media
- 화면 heading, route, responsive flow와 UX writing
- consumer source, exception reviewer와 expiry
- Pencil 제품 화면과 runtime evidence

## Forbidden coupling examples

다음은 generic package에 들어가면 안 된다.

```text
<product-source-root>/features/...
consumerCount: 3
brand.hero.overlay
특정 제품 도메인의 화면 intent
```

해당 정보는 product policy가 소유하고 `deslint` 실행 시 별도 인자로 전달한다.

## Migration rule

기존 제품 계약을 즉시 삭제하지 않는다. 먼저 product definition을 생성하고 기존 resolver와 새 compiler가
동일 값을 내는지 dual-run한다. 결과 fingerprint와 diagnostic parity가 일치한 뒤 제품 adapter를 전환한다.
