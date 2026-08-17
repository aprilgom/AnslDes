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
- [x] color, typography, layout, motion, icon과 component policy model을 확장한다.
- [x] rule id, implementation version, category, platform, evidence kind, default severity, provenance와
      dependency를 표현하는 언어 중립 `RuleSpec` contract를 추가한다.
- [x] version, exact member와 SHA-256 fingerprint를 가진 rule pack manifest contract를 추가한다.
- [x] key 누락·추가·중복·unknown enum을 모든 adapter가 동일하게 거부한다.

## 2. TS/TSX source analyzer

- [x] Tree-sitter TypeScript와 TSX grammar version을 exact pin한다.
- [x] syntax error를 clean 결과로 바꾸지 않는다.
- [x] import, export, declaration, JSX element, attribute, expression과 spread를 normalized IR로 만든다.
- [x] lexical scope, alias, nested shadow와 duplicate local binding을 해석한다.
- [x] relative import, package import와 `tsconfig.paths` module resolver를 구현한다.
- [x] test/mock/generated exclusion을 policy 기반 exact matcher로 구현한다.
- [x] type-aware 규칙은 `SemanticProvider` process boundary 뒤에 둔다.

## 3. Core rule engine과 report

- [x] rule registry, severity, evidence kind와 platform scope를 구현한다.
- [x] warning/error/raw/overflow/overlap budget을 구현한다.
- [x] text, JSON과 SARIF formatter를 구현한다.
- [x] exception은 exact rule/path, rationale, owner와 expiry를 요구한다.
- [x] broad ignore와 report 후처리 우회 negative fixture를 추가한다.
- [x] rule 구현이 공통 interface로 self-describe하고 registry가 engine dispatch 조건문 없이 평가하게 한다.
- [x] registry composition에서 duplicate rule id, missing dependency, incompatible evidence/provider와
      unknown pack member를 거부한다.
- [x] report에 effective pack·rule exact set, activation status와 registry fingerprint를 저장한다.
- [x] root-cause dedup과 모든 diagnostic의 stable ordering을 완성한다.

## 4. Pencil과 computed layout

- [x] Pencil raw property와 computed overflow/overlap/clipping/stale SHA 입력 경계를 구현한다.
- [x] `.pen` reusable instance와 variable resolution model을 computed document visitor evidence로 구현한다.
- [x] typography, color, layout, icon, guide와 UX-writing rule을 구현한다.
- [x] pen.dev adapter를 design-document provider process boundary로 구현한다.
- [x] 100%·160%·235% large-text evidence contract를 구현한다.

## 5. React Native design-system rules

- [x] color와 approved asset exception lint를 구현한다.
- [x] spacing, radius, size, elevation, layer와 touch-target lint를 구현한다.
- [x] motion owner, fallback과 Reduce Motion lint를 구현한다.
- [x] icon geometry, consumer identity와 accessibility binding lint를 구현한다.
- [x] Button, TextField와 native primitive ownership lint를 구현한다.
- [x] selection, list, result, feedback와 overlay recipe lint를 구현한다.
- [x] responsive, typography scaling, accessibility와 governance lint를 구현한다.

각 rule은 wrong module, alias, shadow, spread, noop handler, registry weakening과 false-positive fixture를
가져야 한다. 실제 owner, consumer path와 expected count는 외부 policy로만 주입한다.

## 6. Detector integration

- [x] 외부 detector registry와 implementation snapshot을 version/checksum으로 고정한다.
- [x] Web finding과 native/Pencil equivalent finding을 다른 evidence kind로 유지한다.
- [x] detector rule mapping을 code reference와 함께 versioned catalog·pack contract로 받는다.
- [x] Impeccable/Hallmark canonical 63개를 engine에 하드코딩하지 않고 built-in rule pack으로 제공한다.
- [x] optional review evidence가 deterministic verdict와 fingerprint를 변경하지 못하게 한다.
- [x] upstream drift와 migration note를 release gate에 포함한다.

## 7. Quality gate와 독립 배포

- [x] stage runner가 실패 command, owner, exit code와 stdout/stderr를 report한다.
- [x] passing report만 release evidence로 게시하고 실패 report는 불변 진단으로 보존하며 dependency SHA freshness를 검사한다.
- [x] release manifest가 source artifact checksum을 고정한다.
- [x] release workflow가 macOS, Linux와 Windows binary 및 SHA-256을 생성한다.
- [x] Go unit, race, vet과 package checks를 root gate에서 실행한다.

## 8. 소비 저장소 migration boundary

- [x] 외부 policy가 owner, consumer, exception과 budget을 완전히 표현할 수 있다.
- [x] 외부 policy가 required rule pack의 exact version·fingerprint와 rule activation을 선언할 수 있다.
- [x] 동일 source snapshot에서 이전 linter와 `deslint`를 dual-run할 수 있는 report schema를 제공한다.
- [x] 의도된 diagnostic 차이를 별도 migration note reference로 표현한다.
- [x] exact release version, tag와 checksum을 검증하는 consumer lock schema를 제공한다.

제품별 fixture, parity 결과, rollback release와 기존 linter 제거 승인은 소비 저장소에만 둔다.

## 완료 조건

- [x] AnslDes가 어떤 제품 checkout도 요구하지 않고 모든 check를 통과한다.
- [x] 두 개 이상의 중립 fixture가 같은 schema/compiler/linter를 소비한다.
- [x] parser error, missing evidence와 stale report가 절대 pass가 되지 않는다.
- [x] 외부 policy만 바꾸어 서로 다른 제품을 같은 binary로 검사할 수 있다.
- [x] rule 추가·교체·비활성화 fixture가 engine core 수정 없이 같은 registry lifecycle을 사용한다.
- [x] 제품 이름, source 경로, token 값과 snapshot이 저장소에 존재하지 않는다.
