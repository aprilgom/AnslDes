# Anti-slop release와 소비 저장소 migration

AnslDes release는 generic schema, provider contract, rule registry와 중립 fixture만 배포한다. 제품 route,
token, owner, exception, baseline report, parity 결과와 rollback 승인은 소비 저장소가 소유한다.

## Exact pin 절차

1. `release/ansldes-release.json`의 version과 tag를 exact pin하고 배포 채널에서 얻은 manifest SHA-256을
   소비 저장소의 `consumer-release-lock.schema.json` 형식 lock에 기록한다.
2. 사용하는 package와 `deslint` binary 각각의 version·SHA-256, effective rule pack의 version·fingerprint를
   lock에 기록한다. `latest`, semver range와 checksum 없는 binary는 허용하지 않는다.
3. manifest의 `dependencies`에 있는 vendored Impeccable tree, Hallmark source, anti-slop catalog와 contract
   fingerprint를 lock과 대조한다. 하나라도 다르면 provider stage를 실행하지 않고 실패한다.
4. product definition의 `schemaVersion`과 `productDefinitionVersion`은 독립적으로 기록한다. 제품 token 값만
   바뀌면 AnslDes package version이 아니라 product definition version만 올린다.
5. local과 CI가 같은 lock, policy, definition, source snapshot과 evidence를 사용하게 하고 canonical JSON
   report fingerprint가 같은지 확인한다.

## Cutover와 rollback

- 소비 저장소는 기존 linter와 `deslint`를 동일 source snapshot에서 dual-run하고 rule ID, severity, location과
  verdict를 비교한다. 의도한 차이는 소비 저장소의 migration note에 reviewer와 근거를 기록한다.
- 통과 report는 원본 SHA-256 그대로 저장한다. finding count, exit code 또는 report를 wrapper가 다시 쓰면
  governance violation으로 취급한다.
- 제품 adapter는 lock 검증이 끝난 release만 읽는다. owner, consumer, exact exception과 budget은 product
  policy에 남기고 AnslDes source나 release manifest에 복사하지 않는다.
- parity 승인, rollback 가능한 이전 release와 migration evidence가 모두 있기 전에는 기존 구현을 제거하지 않는다.

## Rule과 pack lifecycle

1. 새 rule은 공통 `RuleSpec`에 stable ID, implementation version, category, platform, evidence kind, severity,
   provenance, provider와 dependency를 선언한다. evaluator는 공용 provider IR를 소비한다.
2. pack 추가는 exact member set과 fingerprint를 생성하고 minor version을 올리며 owner, migration plan과
   default activation을 기록한다. composer는 duplicate ID, missing dependency와 incompatible provider를 거부한다.
3. 동작 교체는 old/new rule mapping과 migration note를 남기고 major version을 올린다. 소비 policy의 exact
   activation fixture로 local/CI report parity를 검증한다.
4. 제거는 major version, tombstone 또는 exact replacement와 migration plan이 모두 필요하다. report 후처리,
   broad ignore와 engine 조건문으로 rule을 숨기지 않는다.
5. deprecated rule도 lock과 effective report에 남는다. 모든 소비자가 새 pack을 pin한 뒤 다음 major에서만
   제거할 수 있다.

로컬과 CI의 공통 gate는 `npm run check`다. 이 명령은 registry freshness, schema fixture, generated model,
Go rule/report test, product boundary와 release manifest freshness를 함께 검사한다.
