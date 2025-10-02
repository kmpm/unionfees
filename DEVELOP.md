# Utveckla Unionfees

Du behöver
- En version av [go](https://go.dev) >= 1.24
- Verktyget `task` som du hittar på https://taskfile.dev
- Verktyget `goreleaser` från https://goreleaser.com/
- Verktyget `reuse` från https://reuse.software


## Utveckling
Utveckla koden som vanligt, i en egen git-gren men före commit så kör
`task pre-commit`.
Detta kör en hel uppsättning andra uppgifter definierade i `Taskfile.yml` som ...
- formaterar koden (`go mod tidy` och `go fmt`)
- analyserar koden (bl.a. staticcheck)
- kontrollerar att man kan förstå vilken licens som använts på alla filer (reuse)
- kontrollerar att goreleaser kan bygga projektet.

### Fillayout
I mappen [docs](./docs) så hittar du det mesta du behöver för att förstå strukturen.


## Release
När all kod som ska vara med i en release finns i grenen `main` på github
så skapa en git-tagg med formatet `vX.Y.Z` som är lämpliga nästa nummer.

Så fort denna tagg finns i github så kommer ett arbetsflöde automatiskt att köras som skapa och laddar upp filer till aktuell release samt skapar
docker image på ghcr.io


## Länkar
- https://go.dev
- https://taskfile.dev
- https://goreleaser.com/
- https://reuse.software
- https://picocss.com/docs

