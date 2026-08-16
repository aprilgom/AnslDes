# Architecture

## 목표

하나의 Go binary가 framework-neutral contract, React Native source IR, Pencil document와 computed-layout
report를 동일한 diagnostic model로 검사한다. 제품 고유 정책은 config로 주입하고 engine과 분리한다.

```text
TS/TSX ── Tree-sitter ── Source IR ─┐
contract JSON ───────── Contract IR ─┼─ Rule Engine ─ Report(JSON/SARIF/text)
.pen JSON ────────────── Pencil IR ──┤
pen.dev output ───────── Layout IR ──┘
```

## Parser strategy

Tree-sitter는 구문과 concrete source range를 소유한다. `internal/source`가 import, declaration, JSX와
expression을 normalized IR로 바꾸고 lexical symbol table과 module resolver를 적용한다. Rule은 raw CST
node kind에 의존하지 않는다.

Tree-sitter가 제공하지 않는 TypeScript type inference는 `SemanticProvider` 선택 기능으로 격리한다.
TypeScript 7의 stable IPC API가 충분해질 때 adapter를 추가하며 compiler internal package에 결합하지 않는다.

## Contract ownership

Canonical contract는 JSON과 JSON Schema다. Go와 제품 TypeScript는 이를 각자의 typed model로 읽거나
generated adapter로 소비한다. 계약 fingerprint는 canonical JSON bytes로 계산한다.

## Evidence boundary

- source: AST와 lexical/module binding
- Pencil: document node, metadata와 variable resolution
- computed-layout: pen.dev engine bounds와 problem output
- browser: Web detector/DOM evidence
- native-runtime: simulator/device evidence

한 evidence kind의 성공으로 다른 platform을 통과 처리하지 않는다.

## Compatibility

Diagnostic의 외부 contract는 `ruleId`, `severity`, `message`, `path`, `range`, `evidenceKind`, `platform`,
`owner`와 `fingerprint`다. 정렬과 직렬화가 결정론적이어야 하며 minor release에서 기존 consumer를 깨지
않는다.
