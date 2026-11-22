# Go per RenkuOS a 32 bit

[English](README.md) · Italiano · [日本語](README.ja.md) · [한국어](README.ko.md)

La toolchain di Go per RenkuOS (Haiku, `x86_gcc2`) su x86 a 32 bit, che gira
sulla macchina stessa: `go`, `gofmt`, la libreria standard e compilatore,
linker, assemblatore e vet. Go 1.26.5, dall'[albero Go per Haiku di korli](https://github.com/korli/go)
con sopra un port sperimentale `haiku/386`. Go ufficiale e HaikuPorts
supportano Haiku solo su x86_64; questo pacchetto non è ufficiale.

## Installazione con pkgman

```sh
pkgman add-repo https://pkgman.rainygirl.com/x86_gcc2
pkgman install golang_x86
```

Se `pkgman add-repo` fallisce con `Operation not supported`, il network kit di
quell'immagine non ha TLS; nell'indirizzo usa `http://` al posto di `https://`.

Il pacchetto installa `go` e `gofmt`; `GOROOT` è
`/boot/system/develop/lib/x86/go`. Occupa circa 60 MB in
`/boot/system/packages`.

## Installazione da file

```sh
pkgman install ./golang_x86-1.26.5-2-x86_gcc2.hpkg
```

## Uso

```sh
go version                    # go version go1.26.5 haiku/386
go build ./...
```

- Niente cgo: `CGO_ENABLED` vale 0 e non si può attivare.
- `GOTOOLCHAIN` è `local`. Non esiste una toolchain `haiku/386` da scaricare,
  quindi un `go.mod` che chiede un Go più recente fallisce invece di scaricarlo.
- I moduli che importano `golang.org/x/sys/unix` non si compilano: il modulo
  pubblicato non ha file per Haiku, per nessuna architettura.
- La prima compilazione costruisce la libreria standard nella cache
  (`~/.cache/go-build`). Se il volume di avvio è piccolo, punta `GOCACHE`,
  `GOPATH` e `TMPDIR` a un altro volume.

## Disinstallazione

```sh
pkgman uninstall golang_x86
```

## Compilare il pacchetto

Su Linux o macOS, con un Go da cui fare il bootstrap e lo strumento `package`
di Haiku:

```sh
./build.sh                    # scrive dist/golang_x86-<version>-2-x86_gcc2.hpkg
```

`tests/run.sh` è lo smoke test, da eseguire sulla macchina a 32 bit dopo
l'installazione. Cosa cambiano le patch, cosa è stato provato e cosa no è in
[docs/NOTES.md](docs/NOTES.md) (in inglese). `patches/` è l'unica serie di patch per
haiku/386: anche gli strumenti a 32 bit di
[renkuos-gotools](https://github.com/rainygirl/renkuos-gotools) sono compilati
con questo toolchain.

## Uso dell'IA

Parti di questo port sono state sviluppate con l'aiuto di strumenti di programmazione basati sull'IA (Claude di Anthropic). Tutto il codice è stato rivisto e testato dall'autore.
