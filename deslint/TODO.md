# deslint TODO

이 문서는 제품 중립적인 Go linter의 구현 순서와 완료 조건을 관리한다. 제품별 source snapshot, owner,
consumer, 예외와 parity report는 소비 저장소가 소유하며 이 저장소에 복사하지 않는다.

## 0. Bootstrap

- [x] 독립 Go module과 Tree-sitter TSX parser boundary를 만든다.
- [x] 중립 positive·negative fixture를 추가한다.
- [x] 언어 중립 definition/policy input 경계를 만든다.
- [x] 제품 snapshot 없이 자체 unit test가 실행된다.

## 1. 언어 중립 contract

- [x] canonical schema에서 TypeScript adapter와 Go model을 생성한다.
- [x] definition과 policy schema version을 독립적으로 검증한다.
- [x] reference existence, unknown key와 cycle을 진단한다.
- [ ] color, typography, layout, motion, icon과 component policy model을 확장한다.
- [ ] key 누락·추가·중복·unknown enum을 모든 adapter가 동일하게 거부한다.

## 2. TS/TSX source analyzer

- [x] Tree-sitter TypeScript와 TSX grammar version을 exact pin한다.
- [x] syntax error를 clean 결과로 바꾸지 않는다.
- [ ] import, export, declaration, JSX element, attribute, expression과 spread를 normalized IR로 만든다.
- [ ] lexical scope, alias, nested shadow와 duplicate local binding을 해석한다.
- [ ] relative import, package import와 `tsconfig.paths` module resolver를 구현한다.
- [ ] test/mock/generated exclusion을 policy 기반 exact matcher로 구현한다.
- [ ] type-aware 규칙은 `SemanticProvider` process boundary 뒤에 둔다.

## 3. Core rule engine과 report

- [x] rule registry, severity, evidence kind와 platform scope를 구현한다.
- [x] warning/error/raw/overflow/overlap budget을 구현한다.
- [x] text, JSON과 SARIF formatter를 구현한다.
- [x] exception은 exact rule/path, rationale, owner와 expiry를 요구한다.
- [x] broad ignore와 report 후처리 우회 negative fixture를 추가한다.
- [ ] root-cause dedup과 모든 diagnostic의 stable ordering을 완성한다.

## 4. Pencil과 computed layout

- [x] Pencil raw property와 computed overflow/overlap/clipping/stale SHA 입력 경계를 구현한다.
- [ ] `.pen` reusable instance와 variable resolution model을 구현한다.
- [ ] typography, color, layout, icon, guide와 UX-writing rule을 구현한다.
- [ ] pen.dev adapter를 process boundary로 구현한다.
- [ ] 100%·160%·235% large-text evidence contract를 구현한다.

## 5. React Native design-system rules

- [ ] color와 approved asset exception lint를 구현한다.
- [ ] spacing, radius, size, elevation, layer와 touch-target lint를 구현한다.
- [ ] motion owner, fallback과 Reduce Motion lint를 구현한다.
- [ ] icon geometry, consumer identity와 accessibility binding lint를 구현한다.
- [ ] Button, TextField와 native primitive ownership lint를 구현한다.
- [ ] selection, list, result, feedback와 overlay recipe lint를 구현한다.
- [ ] responsive, typography scaling, accessibility와 governance lint를 구현한다.

각 rule은 wrong module, alias, shadow, spread, noop handler, registry weakening과 false-positive fixture를
가져야 한다. 실제 owner, consumer path와 expected count는 외부 policy로만 주입한다.

## 6. Detector integration

- [ ] 외부 detector registry와 implementation snapshot을 version/checksum으로 고정한다.
- [ ] Web finding과 native/Pencil equivalent finding을 다른 evidence kind로 유지한다.
- [ ] detector rule mapping을 code reference와 함께 versioned policy로 받는다.
- [ ] LLM-only judgment를 deterministic pass로 직렬화하지 않는다.
- [ ] upstream drift와 migration note를 release gate에 포함한다.

## 7. Quality gate와 독립 배포

- [ ] stage runner가 실패 command, owner, exit code와 stdout/stderr를 report한다.
- [ ] passing report만 저장하고 dependency SHA freshness를 검사한다.
- [x] release manifest가 source artifact checksum을 고정한다.
- [x] release workflow가 macOS, Linux와 Windows binary 및 SHA-256을 생성한다.
- [x] Go unit, race, vet과 package checks를 root gate에서 실행한다.

## 8. 소비 저장소 migration boundary

- [ ] 외부 policy가 owner, consumer, exception과 budget을 완전히 표현할 수 있다.
- [ ] 동일 source snapshot에서 이전 linter와 `deslint`를 dual-run할 수 있는 report schema를 제공한다.
- [ ] 의도된 diagnostic 차이를 별도 migration note reference로 표현한다.
- [ ] exact release version, tag와 checksum을 검증하는 consumer lock schema를 제공한다.

제품별 fixture, parity 결과, rollback release와 기존 linter 제거 승인은 소비 저장소에만 둔다.

## 완료 조건

- [ ] AnslDes가 어떤 제품 checkout도 요구하지 않고 모든 check를 통과한다.
- [ ] 두 개 이상의 중립 fixture가 같은 schema/compiler/linter를 소비한다.
- [ ] parser error, missing evidence와 stale report가 절대 pass가 되지 않는다.
- [ ] 외부 policy만 바꾸어 서로 다른 제품을 같은 binary로 검사할 수 있다.
- [ ] 제품 이름, source 경로, token 값과 snapshot이 저장소에 존재하지 않는다.
