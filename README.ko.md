# 32비트 RenkuOS용 Go

[English](README.md) · [Italiano](README.it.md) · [日本語](README.ja.md) · 한국어

32비트 x86 RenkuOS(Haiku, `x86_gcc2`)에서 직접 실행되는 Go 툴체인입니다.
`go`, `gofmt`, 표준 라이브러리, 컴파일러, 링커, 어셈블러, vet이 들어 있습니다.
Go 1.26.5이며, [korli의 Haiku용 Go 트리](https://github.com/korli/go) 위에
실험적인 `haiku/386` 포트를 얹은 것입니다. Go 공식 배포와 HaikuPorts는
Haiku를 x86_64에서만 지원하며, 이 패키지는 비공식입니다.

## pkgman으로 설치

```sh
pkgman add-repo https://pkgman.rainygirl.com/x86_gcc2
pkgman install golang_x86
```

`pkgman add-repo`가 `Operation not supported`로 실패하면 그 이미지의 네트워크
키트에 TLS가 없는 것입니다. 주소의 `https://`를 `http://`로 바꿔 쓰세요.

패키지는 `go`와 `gofmt`를 설치하며 `GOROOT`는
`/boot/system/develop/lib/x86/go`입니다. `/boot/system/packages`에서 약 60 MB를
차지합니다.

## 파일로 설치

```sh
pkgman install ./golang_x86-1.26.5-2-x86_gcc2.hpkg
```

## 사용

```sh
go version                    # go version go1.26.5 haiku/386
go build ./...
```

- cgo는 없습니다. `CGO_ENABLED`는 0이고 켤 수 없습니다.
- `GOTOOLCHAIN`은 `local`입니다. 내려받을 `haiku/386` 툴체인이 없으므로,
  더 새로운 Go를 요구하는 `go.mod`는 내려받지 않고 실패합니다.
- `golang.org/x/sys/unix`를 import하는 모듈은 빌드되지 않습니다. 공개된
  모듈에는 어떤 아키텍처용 Haiku 파일도 없습니다.
- 첫 빌드는 표준 라이브러리를 빌드 캐시(`~/.cache/go-build`)에 컴파일합니다.
  부트 볼륨이 작으면 `GOCACHE`, `GOPATH`, `TMPDIR`을 다른 볼륨으로 지정하세요.

## 제거

```sh
pkgman uninstall golang_x86
```

## 패키지 빌드

Linux나 macOS에서, 부트스트랩용 Go와 Haiku의 `package` 도구가 있어야 합니다.

```sh
./build.sh                    # dist/golang_x86-<version>-2-x86_gcc2.hpkg 생성
```

`tests/run.sh`는 설치 후 32비트 기기에서 돌리는 스모크 테스트입니다. 패치가
무엇을 바꾸는지, 무엇을 시험했고 무엇을 하지 않았는지는
[docs/NOTES.md](docs/NOTES.md)(영문)에 있습니다. `patches/`가 haiku/386용 패치의
전부이며, [renkuos-gotools](https://github.com/rainygirl/renkuos-gotools)의
32비트 도구들도 이 툴체인으로 빌드합니다.

## AI 사용 고지

이 포트의 일부는 AI 코딩 도구(Anthropic Claude)의 도움을 받아 개발했습니다.
모든 코드는 작성자가 검토하고 테스트했습니다.
