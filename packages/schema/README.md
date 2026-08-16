# @ansldes/schema

제품 디자인 시스템 definition의 언어 중립 schema다. 이 package에는 실제 브랜드 token 값이나 제품
source policy가 포함되지 않는다.

Definition v1은 다음을 분리한다.

- 제품 identity와 독립 version
- theme axis
- foundation token
- optional icon geometry, usage와 action recipe
- component anatomy와 recipe

제품은 [`design-system-definition.schema.json`](./design-system-definition.schema.json)을 참조하는 JSON을
자기 저장소에서 소유한다. Reference 형식은 `{collection.layer.name}`이다.

현재 schema는 구조와 값 종류를 검증한다. Reference 존재, cycle, theme exact-set과 layer compatibility는
compiler 및 `deslint`가 추가 검증한다.

`foundations.icon`은 definition v1의 선택 그룹이다. size, stroke, optical alignment, geometry, usage와
action을 제품 저장소가 정의하며, schema에는 특정 glyph path나 제품별 consumer가 들어가지 않는다.

## Inputs

- `design-system-definition.schema.json`: 렌더링에 필요한 제품 token과 component recipe
- `design-system-policy.schema.json`: severity, raw-property 분류, exact exclusion, evidence와 budget

Definition과 policy는 별도 versioned input이다. 제품 source 경로와 예외는 definition이 아니라 policy에만
둔다.

## Generated models

```bash
npm run generate:models
npm run generate:models:check
```

두 schema에서 TypeScript model과 `deslint` Go model을 결정론적으로 생성한다. 생성물은 schema SHA-256을
포함하며 freshness 검사가 root quality gate에 연결돼 있다.
