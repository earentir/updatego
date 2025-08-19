# updatego

A simple tool to install and update go to its latest version

## Install
```bash
curl -fsSL https://github.com/earentir/updatego/releases/latest/download/updatego -o updatego && chmod +x updatego
```


## Usage/Examples
```
$ ./updatego

Usage: updatego [OPTIONS] COMMAND [arg...]

A simple golang version manager

Options:
  -v, --version   Show the version and exit
      --verbose   Enable verbose output

Commands:
  install         Install Go
  status          Check Go installation status
  latest          Print the latest Go version available
  update          Update Go to the latest version
  list            List all local Go versions
  switch          Switch to a specific Go version

Run 'updatego COMMAND --help' for more information on a command.
```


## Dependancies & Documentation
[![Go Mod](https://img.shields.io/github/go-mod/go-version/earentir/updatego)]()

[![Go Reference](https://pkg.go.dev/badge/github.com/earentir/updatego.svg)](https://pkg.go.dev/github.com/earentir/updatego)

[![Dependancies](https://img.shields.io/librariesio/github/earentir/updatego)]()

## Contributing

Contributions are always welcome!
All contributions are required to follow the https://google.github.io/styleguide/go/

All code contributed must include its tests in (_test) and have a minimum of 80% coverage

## Vulnerability Reporting

Please report any security vulnerabilities to the project using issues or directly to the owner.

## Code of Conduct
 This project follows the go project code of conduct, please refer to https://go.dev/conduct for more details

## Roadmap
- [x] Check paths
- [x] Install go
- [ ] make changes in bashrc

## Authors

- [@earentir](https://www.github.com/earentir)

## License

I will always follow the Linux Kernel License as primary, if you require any other OPEN license please let me know and I will try to accomodate it.

[![License](https://img.shields.io/github/license/earentir/gitearelease)](https://opensource.org/license/gpl-2-0)
