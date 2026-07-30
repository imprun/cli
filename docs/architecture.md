# Imprun CLI 아키텍처

이 문서는 공개 명령줄 도구 `imprun`의 현재 책임과 제품 간 계약을 고정한다.
관련 구현 마스터 이슈는
[imprun/imprun#105](https://github.com/imprun/imprun/issues/105)다.

## 책임 경계

| 구성 요소 | 소유 책임 |
| --- | --- |
| Imprun CLI (`imprun/cli`) | 사용자 명령, 로컬 컨텍스트, OS 보안 저장소, Device Authorization 클라이언트, Cloud/Cell API 호출 |
| Imprun Identity | 사람 인증, Device Authorization, 토큰 발급·갱신·해지 |
| Imprun Cloud (`imprun/imprun`) | 테넌트 멤버십과 역할, Cell 연결, 요청 위임과 감사 |
| Windforce Core (`imprun/windforce-core`) | 제품 중립적인 Workspace, App, Release, Run 계약과 Cell 실행 |

CLI 저장소에는 서버, 워커, 데이터베이스, Imprun 내부 인프라 설정을 넣지 않는다.
Cloud 서버 프로세스의 실행 파일은 `imprun-server`, Cell 런타임은
`windforce-core`다.

## 로그인과 요청 경로

```text
imprun auth login
  -> Cloud의 /.well-known/imprun-cli.json 조회
  -> Imprun Identity의 OAuth 2.0 Device Authorization
  -> OS 보안 저장소에 계정별 자격 증명 저장
  -> Cloud에서 테넌트 멤버십과 역할 검증
  -> 허용된 요청만 대상 Cell로 위임
```

- 새 로그인은 시크릿이 없는 공개 클라이언트 `imprun-cli`를 사용한다.
- 설정 파일에는 컨텍스트와 계정 식별자만 저장하고 토큰은 저장하지 않는다.
- Cell은 Identity 토큰만 보고 Cloud 멤버십을 추론하지 않는다. Cloud가 권한을
  확인하고 짧은 범위의 위임 문맥을 만들어 Cell로 전달한다.
- Identity 로그아웃과 로컬 자격 증명 삭제는 구분한다. 일반 로그아웃은 원격
  refresh token 해지가 성공한 뒤 로컬 값을 지운다.

## `wf`에서 이전

`wf` 실행 파일과 영구 별칭은 새 릴리스에 포함하지 않는다. 기존 릴리스 자산은
수정하지 않고 그대로 보존한다.

처음 실행할 때 새 설정이 없으면 기존 `wf/config.json`을 읽을 수 있다. 기존
운영체제 보안 저장소의 자격 증명을 찾으면 `imprun` 영역으로 복사한다. 이관한
자격 증명은 원래 issuer와 `wf-cli` client ID를 유지하므로 교체 기간 동안
갱신과 해지가 가능하다. 새 로그인부터 `imprun-cli`를 사용한다.

사용자가 `IMPRUN_CONFIG`를 명시하면 자동 설정 이전을 하지 않는다. 새 CLI에서
인증 상태를 확인한 다음 기존 `wf` 실행 파일을 `PATH`에서 제거한다.

## 배포와 호환성

- 공개 CI는 GitHub-hosted runner만 사용하며 조직 ARC와 비밀 값에 의존하지 않는다.
- GoReleaser는 OS/아키텍처별 휴대형 아카이브와 WinGet이 바로 설치할 수 있는
  Windows amd64/arm64 단일 `imprun.exe` 자산을 함께 만든다.
- 체크섬과 키 없는 Cosign 서명을 검증한 자산만 설치 경로에 사용한다.
- Windows 패키지 ID는 `Imprun.CLI`다.
- Cloud와 Identity가 새 클라이언트를 수용하고 실제 Cell E2E가 통과하기 전에는
  안정 버전이나 Winget 패키지를 배포하지 않는다.

## 완료 기준

1. Windows amd64와 Linux arm64에서 동일한 명령 계약을 검증한다.
2. Device Authorization, 토큰 갱신·해지, 계정 전환을 검증한다.
3. Cloud 테넌트 권한을 거쳐 실제 Cell의 App, Release, Run, Workspace를 제어한다.
4. 기존 `wf` 설정과 자격 증명 이전을 검증한다.
5. 서명된 릴리스와 Winget 설치·업그레이드 경로를 검증한다.
