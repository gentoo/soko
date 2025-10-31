# Go Compiler

To update the go compiler itself, it mainly consists of updating all the relevant places:

1. `.gitlab-ci.yml` - the docker version for `go-test`
2. `Dockerfile*` - the base image for the dockers
3. `README.md` - the badge link
4. `go.mod` - the go version for modules file (file header)

It is recommended to update all Go dependencies after compiler upgrade.

# Go Dependencies

This one is very simple and straightforward, and most of the times just work without issues:

```bash
go get -t -u ./...
go mod tidy
```

# Node Compiler

Just update the docker version in `Dockerfile`.

# Node Dependencies

This one is more complicated and error prune, so prepare for surprises at any point. I mostly do this through a docker run, since I don't want to compile nodejs on my development PC. For example with a docker:

```bash
docker run -it --rm -v "${PWD}:/go/src/soko" -w /go/src/soko node:24 bash
```

Then inside the docker, run this command to update:

```bash
npx npm-check-updates -u && rm package-lock.json && npm install && npx webpack
```

And this should regen `package-lock.json` (which included the full tree of dependencies).
