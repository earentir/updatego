# updatego

A simple tool to install and update go to its latest version

## Install
```bash
curl -fsSL https://github.com/earentir/updatego/releases/latest/download/updatego -o updatego && chmod +x updatego
```


## Usage
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
  toolchain       Show or set GOTOOLCHAIN (auto|local)
  sync            Set GOTOOLCHAIN to the active managed Go version

Run 'updatego COMMAND --help' for more information on a command.
```

## Examples

### Install current release
```bash
[root@themis ~]# updatego install
```
#### Expected Output
```bash
Installing Go version: 1.25.0
Downloading go1.25.0.linux-amd64.tar.gz, writing to: go1.25.0.linux-amd64.tar.gz
URL: https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
Writing to: /tmp/go1.25.0.linux-amd64.tar.gz
####################################################################################################################
Extracting the new Go version...
Switching to the newly installed Go version: 1.25.0
Switched to Go version 1.25.0 successfully.
Setting up environment variables...
GOROOT set to: /usr/local/go
GOPATH set to: /root/go
```

## Configuration

After a successful `install`, updatego writes `extract_root` to `~/.config/updatego/config.json`. Commands `update`, `status`, `list`, and `switch` use that directory (default `/usr/local` when the file is missing).

`install --user` uses `$HOME/.local` as the extract root (`$HOME/.local/go` is GOROOT). GOPATH remains `$HOME/go`.

## Toolchain (`GOTOOLCHAIN`)

Commands use the managed Go at `extract_root/go/bin/go` (not necessarily the `go` on your `PATH`).

- `updatego toolchain` — print current `GOTOOLCHAIN`
- `updatego toolchain auto` / `updatego toolchain local` — run `go env -w GOTOOLCHAIN=…`
- `updatego switch VERSION` — swap the install tree, set `GOTOOLCHAIN=local`, and update `toolchain` in `./go.mod` when that file exists (including when already on `VERSION`)
- `updatego sync` — set `GOTOOLCHAIN=goX.Y.Z` to match the active managed Go (does not change `go.mod`)

`install` and `update` do not modify `GOTOOLCHAIN` or `go.mod`.

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
