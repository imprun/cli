#!/usr/bin/env python3

from __future__ import annotations

import hashlib
import io
import os
import platform
import shutil
import subprocess
import tarfile
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
VERSION = "0.0.0-test"
TAG = f"v{VERSION}"


def target_platform() -> tuple[str, str]:
    system = platform.system()
    machine = platform.machine().lower()
    platforms = {"Linux": "linux", "Darwin": "darwin", "Windows": "windows"}
    architectures = {
        "x86_64": "amd64",
        "amd64": "amd64",
        "arm64": "arm64",
        "aarch64": "arm64",
    }
    if system not in platforms:
        raise RuntimeError(f"unsupported test operating system: {system}")
    if machine not in architectures:
        raise RuntimeError(f"unsupported test architecture: {machine}")
    return platforms[system], architectures[machine]


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def write_unix_asset(path: Path) -> None:
    payload = (
        "#!/bin/sh\n"
        "case \"${1:-}\" in\n"
        f"  --version) printf 'imprun {VERSION}\\n' ;;\n"
        "  *) printf 'fixture\\n' ;;\n"
        "esac\n"
    ).encode()
    info = tarfile.TarInfo("imprun")
    info.mode = 0o755
    info.size = len(payload)
    with tarfile.open(path, "w:gz") as archive:
        archive.addfile(info, io.BytesIO(payload))


def write_windows_asset(path: Path) -> None:
    command = [
        "go",
        "build",
        "-trimpath",
        "-ldflags",
        f"-X github.com/imprun/cli/internal/controlcli.Version={VERSION}",
        "-o",
        str(path),
        "./cmd/imprun",
    ]
    subprocess.run(command, cwd=ROOT, check=True)


def write_fake_cosign(directory: Path, succeeds: bool) -> None:
    if platform.system() == "Windows":
        content = (
            '@if not "%1"=="verify-blob" exit /b 1\r\n'
            f"@exit /b {0 if succeeds else 1}\r\n"
        )
        (directory / "cosign.cmd").write_text(content, encoding="ascii")
        return
    path = directory / "cosign"
    path.write_text(
        "#!/bin/sh\n"
        "test \"$1\" = verify-blob\n"
        f"exit {0 if succeeds else 1}\n",
        encoding="utf-8",
    )
    path.chmod(0o755)


def installer_command(install_dir: Path, release_root: Path, version: str) -> list[str]:
    if platform.system() == "Windows":
        return [
            "powershell",
            "-NoProfile",
            "-ExecutionPolicy",
            "Bypass",
            "-File",
            str(ROOT / "install.ps1"),
            "-Version",
            version,
            "-InstallDir",
            str(install_dir),
            "-ReleaseBaseUrl",
            release_root.as_uri(),
            "-NoModifyPath",
            "-RequireSignature",
        ]
    return [
        "sh",
        str(ROOT / "install.sh"),
        "--version",
        version,
        "--install-dir",
        str(install_dir),
        "--require-signature",
    ]


def run_installer(
    install_dir: Path,
    release_root: Path,
    fake_bin: Path,
    version: str = VERSION,
    environment_overrides: dict[str, str] | None = None,
) -> subprocess.CompletedProcess[str]:
    environment = os.environ.copy()
    environment["PATH"] = str(fake_bin) + os.pathsep + environment["PATH"]
    environment["IMPRUN_RELEASE_BASE_URL"] = release_root.as_uri()
    if environment_overrides:
        environment.update(environment_overrides)
    return subprocess.run(
        installer_command(install_dir, release_root, version),
        cwd=ROOT,
        env=environment,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )


def require_success(result: subprocess.CompletedProcess[str], context: str) -> None:
    if result.returncode != 0:
        raise AssertionError(
            f"{context} failed with exit {result.returncode}\n"
            f"stdout:\n{result.stdout}\nstderr:\n{result.stderr}"
        )


def main() -> None:
    target_os, architecture = target_platform()
    suffix = ".exe" if target_os == "windows" else ".tar.gz"
    asset = f"imprun_{VERSION}_{target_os}_{architecture}{suffix}"

    test_root = Path(tempfile.mkdtemp(prefix="imprun installer test "))
    try:
        release_root = test_root / "release"
        release_dir = release_root / TAG
        install_dir = test_root / "install path"
        fake_bin = test_root / "fake bin"
        release_dir.mkdir(parents=True)
        install_dir.mkdir()
        fake_bin.mkdir()

        asset_path = release_dir / asset
        if target_os == "windows":
            write_windows_asset(asset_path)
            installed = install_dir / "imprun.exe"
        else:
            write_unix_asset(asset_path)
            installed = install_dir / "imprun"

        (release_dir / "checksums.txt").write_text(
            f"{sha256(asset_path)}  {asset}\n", encoding="ascii"
        )
        (release_dir / "checksums.txt.sigstore.json").write_text("{}\n", encoding="ascii")
        installed.write_bytes(b"old executable")
        if target_os != "windows":
            installed.chmod(0o755)

        write_fake_cosign(fake_bin, succeeds=True)
        success = run_installer(install_dir, release_root, fake_bin)
        require_success(success, "installer success case")

        if target_os == "windows":
            installer_source = (ROOT / "install.ps1").read_text(encoding="utf-8")
            if "::OSArchitecture" in installer_source:
                raise AssertionError("Windows installer requires RuntimeInformation.OSArchitecture")
            native_architecture = {"amd64": "AMD64", "arm64": "ARM64"}[architecture]
            emulated_install_dir = test_root / "emulated process install"
            emulated = run_installer(
                emulated_install_dir,
                release_root,
                fake_bin,
                environment_overrides={
                    "PROCESSOR_ARCHITECTURE": "x86",
                    "PROCESSOR_ARCHITEW6432": native_architecture,
                },
            )
            require_success(emulated, "emulated-process architecture detection")
            missing_architecture = run_installer(
                test_root / "missing architecture install",
                release_root,
                fake_bin,
                environment_overrides={
                    "PROCESSOR_ARCHITECTURE": "",
                    "PROCESSOR_ARCHITEW6432": "",
                },
            )
            if missing_architecture.returncode == 0:
                raise AssertionError("missing Windows architecture was accepted")

        version_result = subprocess.run(
            [str(installed), "--version"],
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            check=True,
        )
        if VERSION not in version_result.stdout:
            raise AssertionError(f"installed version mismatch: {version_result.stdout}")

        invalid = run_installer(install_dir, release_root, fake_bin, "../invalid")
        if invalid.returncode == 0:
            raise AssertionError("invalid version was accepted")

        installed_checksum = sha256(installed)
        write_fake_cosign(fake_bin, succeeds=False)
        signature_failure = run_installer(install_dir, release_root, fake_bin)
        if signature_failure.returncode == 0:
            raise AssertionError("signature failure was accepted")
        if sha256(installed) != installed_checksum:
            raise AssertionError("failed upgrade replaced the installed executable")

        write_fake_cosign(fake_bin, succeeds=True)
        checksum_file = release_dir / "checksums.txt"
        checksum_file.write_text(
            checksum_file.read_text(encoding="ascii")
            + f"{sha256(asset_path)}  {asset}\n",
            encoding="ascii",
        )
        duplicate_checksum = run_installer(install_dir, release_root, fake_bin)
        if duplicate_checksum.returncode == 0:
            raise AssertionError("duplicate checksum entry was accepted")
        if sha256(installed) != installed_checksum:
            raise AssertionError("checksum failure replaced the installed executable")

        print(f"{target_os}/{architecture} installer tests passed")
    finally:
        shutil.rmtree(test_root, ignore_errors=True)


if __name__ == "__main__":
    main()
