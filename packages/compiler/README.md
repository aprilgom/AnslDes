# @ansldes/compiler

AnslDes product definition을 검증하고 theme별 resolved bundle과 canonical SHA-256을 생성한다.

```bash
node packages/compiler/src/cli.mjs compile product.design-system.json \
  --out product.design-system.resolved.json
```

Compiler는 다음을 실패로 처리한다.

- JSON Schema 위반
- theme mapping 누락·추가
- 존재하지 않는 reference
- collection type을 벗어난 reference
- reference cycle
- typography token/weight 불일치
- icon size, stroke, optical alignment와 glyph reference 불일치

Compiler는 optional icon group과 motion/elevation recipe도 theme bundle에 canonical order로 포함한다.

제품 source owner, consumer와 exception은 definition이 아니라 별도 product policy 입력이다.
