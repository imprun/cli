# AGENTS.md

`imprun`은 Imprun Cloud와 연결된 Windforce Cell을 제어하는 공개 CLI다.

## 경계

- 사람 인증은 Imprun Identity의 OAuth 2.0 Device Authorization으로 처리한다.
- 테넌트 멤버십과 역할, Cloud에서 Cell로 위임하는 권한은 Imprun Cloud가 소유한다.
- Cell 실행 API와 중립적인 App, Release, Run, Workspace 계약은 windforce-core가 소유한다.
- CLI에는 서버, 워커, 상태 저장소 또는 Imprun 내부 인프라 코드를 포함하지 않는다.
- 공개 저장소이므로 조직 self-hosted runner와 비밀 값에 의존하지 않는다.

## 호환성

- 사용자 명령은 `imprun`이다. `wf` 실행 파일 별칭은 배포하지 않는다.
- 기존 `wf` 설정과 운영체제 보안 저장소의 자격 증명은 읽어서 `imprun` 영역으로 한 번 이관할 수 있다.
- 새 로그인은 공개 OAuth 클라이언트 `imprun-cli`를 사용한다. 클라이언트 시크릿은 없다.
- 기존 `wf-cli` 자격 증명은 교체 기간 동안 갱신과 해지가 가능해야 한다.

## 작업 방식

- 커밋은 DCO 서명(`git commit -s`)을 포함한다.
- Go 코드는 `gofmt`를 적용한다.
- 완료 전에 `make fmt`, `make build`, `make test`, `make snapshot`을 실행한다.
- 인증 코드, 토큰, 사용자 코드, 브라우저 쿠키 및 내부 주소를 로그나 저장소에 남기지 않는다.

