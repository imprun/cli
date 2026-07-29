# Imprun CLI

`imprun`은 Imprun Cloud와 연결된 Windforce Cell에서 앱, 릴리스, 실행과 워크스페이스를 관리하는 명령줄 도구입니다.

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
