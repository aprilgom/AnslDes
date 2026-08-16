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

제품 경계를 처음 이관할 때는 product definition을 생성하고 기존 resolver와 새 compiler가 동일한 값을
내는지 검증한다. Definition v2 전환 자체는 호환 계층 없이 한 번에 수행하며 v1 입력이나 bundle을
runtime에 남기지 않는다.
