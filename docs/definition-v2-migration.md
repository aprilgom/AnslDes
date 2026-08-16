# Definition v2 migration

Definition v2와 compiled bundle v2는 v1을 지원하지 않는 breaking contract다. 소비 저장소는 legacy
parser나 dual-read adapter를 추가하지 않고 definition, compiler, runtime pin을 같은 변경에서 교체한다.

## Required token graph

Color, spacing, radius, size는 다음 그래프만 허용한다.

```text
raw value → primitive → semantic → component → component recipe
```

- primitive만 raw color·number를 선언한다.
- semantic은 같은 collection의 primitive만 참조한다.
- component는 같은 collection의 semantic만 참조한다.
- component recipe는 layered foundation의 component만 참조한다.
- semantic self-reference, component→primitive, recipe→semantic·primitive·asset은 모두 오류다.

Typography, motion, elevation, layer와 icon은 각자의 명시적 schema를 따른다. Component recipe 내부의
상태, 접근성, opacity 같은 비-token scalar가 자동으로 foundation token이 되는 것은 아니다.

## Direct cutover

1. 제품 exporter가 schemaVersion 2 정의를 직접 생성하게 한다.
2. 기존 semantic/component raw 값을 primitive에 올리고 모든 중간 reference를 명시한다.
3. component recipe reference를 component layer로 교체한다.
4. `@ansldes/schema`, compiler, core와 React Native adapter를 같은 `v1.0.0` release로 pin한다.
5. v2 bundle을 다시 만들고 schema, compiler, `deslint`, product parity gate를 실행한다.
6. v1 산출물과 compatibility branch를 남기지 않는다.

잘못된 layer jump는 JSON Schema, compiler와 Go `deslint`에서 각각 독립적으로 실패한다.
