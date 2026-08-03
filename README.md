# Imprun CLI

`imprun`은 Imprun Cloud와 연결된 Windforce Cell에서 앱, 릴리스, 실행과 워크스페이스를 관리하는 명령줄 도구입니다.

## 설치

Windows PowerShell:

```powershell
irm https://github.com/imprun/cli/releases/latest/download/install.ps1 | iex
```

Linux 또는 macOS:

```shell
curl -fsSL https://github.com/imprun/cli/releases/latest/download/install.sh | sh
```

설치기는 운영체제와 아키텍처에 맞는 최신 안정 릴리스를 선택하고 SHA-256을 검증합니다.
`cosign`이 있으면 키 없는 서명도 자동 검증합니다. 스크립트를 먼저 검토하거나 특정
버전을 고정하고 서명 검증을 필수화하려면 [CLI 설치 가이드](docs/cli.md#install)를
따르세요. Winget 패키지가 게시된 뒤에는 `winget install --id Imprun.CLI --exact`도
사용할 수 있습니다.

```powershell
imprun auth login
imprun context show
imprun app list
imprun run create hello echo --input-json '{"message":"hello"}'
```

사람 로그인은 Imprun Identity의 OAuth 2.0 Device Authorization을 사용합니다. 비밀 값은 설정 파일에 기록하지 않고 Windows Credential Manager, macOS Keychain 또는 Linux Secret Service에 보관합니다.

전체 명령은 [CLI 사용 가이드](docs/cli.md), 제품 간 책임과 인증·권한 경계는
[아키텍처](docs/architecture.md)를 참고하세요.

## 개발

Go 1.23 이상과 설치기 테스트용 Python 3.13을 사용합니다.

```powershell
make fmt
make build
make test
make snapshot
```

Windows에서 `make`를 사용하지 않는 경우:

```powershell
go fmt ./...
go test ./...
go build -trimpath -o .tmp/bin/imprun.exe ./cmd/imprun
```

Apache-2.0 라이선스와 DCO 정책을 따릅니다.
