# 32 ビット RenkuOS 用 Go

[English](README.md) · [Italiano](README.it.md) · 日本語 · [한국어](README.ko.md)

32 ビット x86 の RenkuOS (Haiku、`x86_gcc2`) の上でそのまま動く Go ツールチェーン
です。`go`、`gofmt`、標準ライブラリ、コンパイラ、リンカ、アセンブラ、vet が
入っています。Go 1.26.5 で、[korli の Haiku 用 Go ツリー](https://github.com/korli/go)
に実験的な `haiku/386` ポートを重ねたものです。Go 本家と HaikuPorts が Haiku を
サポートするのは x86_64 だけで、このパッケージは非公式です。

## pkgman でインストール

```sh
pkgman add-repo https://pkgman.rainygirl.com/x86_gcc2
pkgman install golang_x86
```

`pkgman add-repo` が `Operation not supported` で失敗する場合、そのイメージの
ネットワークキットに TLS がありません。アドレスの `https://` を `http://` に
変えてください。

パッケージは `go` と `gofmt` をインストールします。`GOROOT` は
`/boot/system/develop/lib/x86/go` です。`/boot/system/packages` で約 60 MB を
使います。

## ファイルからインストール

```sh
pkgman install ./golang_x86-1.26.5-3-x86_gcc2.hpkg
```

## 使い方

```sh
go version                    # go version go1.26.5 haiku/386
go build ./...
```

- cgo はありません。`CGO_ENABLED` は 0 で、有効にできません。
- `GOTOOLCHAIN` は `local` です。ダウンロードできる `haiku/386` ツールチェーンは
  ないので、より新しい Go を要求する `go.mod` は取得せずに失敗します。
- `golang.org/x/sys/unix` を import するモジュールはビルドできません。公開されて
  いるモジュールには、どのアーキテクチャの Haiku 用ファイルもありません。
- 最初のビルドでは標準ライブラリをビルドキャッシュ (`~/.cache/go-build`) に
  コンパイルします。ブートボリュームが小さい場合は `GOCACHE`、`GOPATH`、
  `TMPDIR` を別のボリュームに向けてください。
- `net` パッケージは Haiku の hosts ファイル
  `/boot/system/settings/network/hosts` を読みます (Haiku には `/etc/hosts` が
  ありません)。そのため `localhost` を引けます。リビジョン 1 と 2 は DNS
  サーバーに問い合わせていました。

## アンインストール

```sh
pkgman uninstall golang_x86
```

## パッケージのビルド

Linux か macOS で、ブートストラップ用の Go と Haiku の `package` ツールが必要です。

```sh
./build.sh                    # dist/golang_x86-<version>-3-x86_gcc2.hpkg を作成
```

`tests/run.sh` は、インストール後に 32 ビット機で実行するスモークテストです。
パッチの内容、テストしたこと、していないことは
[docs/NOTES.md](docs/NOTES.md) (英語) にあります。`patches/` が haiku/386 の唯一の
パッチ一式で、[renkuos-gotools](https://github.com/rainygirl/renkuos-gotools) の
32 ビット版ツールもこのツールチェインでビルドしています。

## AI の利用について

このポートの一部は AI コーディングツール (Anthropic の Claude) の支援を受けて開発しました。
すべてのコードは作者がレビューし、テストしています。
