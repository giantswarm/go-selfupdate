This command line tool is a small wrapper of [`selfupdate.DetectLatest()`](https://pkg.go.dev/github.com/creativeprojects/go-selfupdate/selfupdate#DetectLatest).

Please install using `go get`.

```
$ go install github.com/giantswarm/go-selfupdate/cmd/detect-latest-release@latest
```

To know the usage, please try the command without any argument.

```
$ detect-latest-release
```

For example, following shows the latest version of [resticprofile](https://github.com/creativeprojects/resticprofile).

```
$ detect-latest-release creativeprojects/resticprofile
```

