# jflint-go

jflint-go helps to lint a Declarative Jenkinsfile via [sending request to Jenkins server](https://www.jenkins.io/doc/book/pipeline/development/#linter).

This project is highly inspired by its original implemention [jflint](https://github.com/miyajan/jflint), but written in Go.

## Installation

For Go 1.16+

```bash
go install github.com/masakichi/jflint-go@latest
```

Or use `go get`

```bash
GO111MODULE=on go get github.com/masakichi/jflint-go
```

## Usage

### Config file

default location is `$HOME/.jflintrc`

```
{
  "jenkinsUrl": "http://jenkins.example.com",
  "username": "admin",
  "password": "p@ssword"
}
```

### Example

```
$ jflint-go deploy-xxx.Jenkinsfile
```

### Full usage

```bash
$ jflint-go -h
jflint-go helps to lint a Declarative Jenkinsfile.

This tool itself does not lint a Jenkinsfile,
but sends a request to Jenkins in the same way
as curl approach and displays the result.

Usage:
  jflint-go Jenkinsfile [flags]

Flags:
  -c, --config string        config file (default is $HOME/.jflintrc)
      --csrf-disabled        Specify when CSRF security setting is disabled on Jenkins.
  -h, --help                 help for jflint-go
  -j, --jenkins-url string   Specify Jenkins URL
  -p, --password string      Specify password/API token on Jenkins
  -u, --username string      Specify username on Jenkins
```

## License

[MIT](https://choosealicense.com/licenses/mit/)

## Acknowledgements

- [miyajan/jflint: A client tool to lint Jenkinsfile](https://github.com/miyajan/jflint)
- [Cobra.Dev](https://cobra.dev/)
