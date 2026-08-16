# Release process

AnslDes 소비자는 `latest`나 branch name을 사용하지 않고 release tag와 manifest SHA-256을 함께 고정한다.
제품 definition version은 AnslDes toolkit version과 독립적이다.

## Release candidate

1. `develop`에서 `npm run check`를 통과시킨다.
2. 모든 workspace package와 `deslint`의 release version을 맞춘다.
3. `npm run release:manifest`로 [`release/ansldes-release.json`](../release/ansldes-release.json)을 갱신한다.
4. `npm run release:check`가 source artifact fingerprint와 manifest freshness를 검증하는지 확인한다.
5. `develop`에서 `main`으로 release PR을 만들고 검증 결과를 첨부한다.

## Tag and consumer pin

Release PR이 병합된 `main`의 정확한 commit에 manifest의 `release.tag`를 붙인다. Release workflow는
Linux·macOS·Windows `deslint` binary와 네 package archive를 만들고 SHA-256 파일을 GitHub release asset으로
제공한다. tag가 `main`의 정확한 tip이 아니거나 manifest version과 다르면 배포는 실패한다. 제품 저장소는
다음 값을 lock artifact에 기록한다.

- repository URL
- release tag와 commit SHA
- release manifest SHA-256
- 사용하는 package 또는 `deslint` asset의 SHA-256
- product definition/policy version과 fingerprint

릴리스 tag 또는 checksum이 달라지면 product quality gate는 실패해야 한다. network의 현재 상태를 매 빌드마다
조회하지 않고, 승인된 lock artifact와 vendored manifest를 결정론적으로 검사한다.

## Rollback boundary

제품은 새 adapter의 dual-run parity report와 이전 실행 경로가 포함된 rollback release가 모두 확인되기 전까지
기존 디자인 시스템 구현을 제거하지 않는다. schema breaking change만 AnslDes major version을 올리고, 제품의
color·radius 같은 값 변경은 product definition version만 올린다.
