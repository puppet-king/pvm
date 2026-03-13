# PVM for Windows

[Support this project](https://github.com/sponsors/hjbdev)

> [!TIP]
> Looking for the 0.x (composer) version? See the [v0 branch](https://github.com/hjbdev/pvm/tree/v0).

Removing the hassle of changing PHP versions in the CLI on Windows.

This package has a much more niche use case than nvm does. When developing on Windows and using the integrated terminal, it's quite difficult to get those terminals to _actually_ listen to PATH changes.

This utility changes that.

## Installation

`pvm` currently supports x64 Windows only.

```powershell
irm https://pvm.hjb.dev/install.ps1 | iex
```

You can still install manually by downloading the latest `pvm.exe` release, placing it in `%UserProfile%\.pvm\bin` (for example `C:\Users\Harry\.pvm\bin`), and adding that folder to your PATH.

## Commands
```
pvm list
```
Will list out all the available PHP versions you have installed

```
pvm path
```
Will tell you what to put in your Path variable.

```
pvm use 8.2.9
```
> [!NOTE]  
> Versions must have major.minor specified in the *use* command. If a .patch version is omitted, newest available patch version is chosen.

Will switch your currently active PHP version to PHP 8.2.9

```
pvm install 8.2
```
> [!NOTE]  
> The install command will automatically determine the newest minor/patch versions if they are not specified

Will install PHP 8.2 at the latest patch.

## Composer support
`pvm` now installs also composer with each php version installed.
It will install Composer latest stable release for PHP >= 7.2 and Composer latest 2.2.x LTS for PHP < 7.2.
You'll be able to invoke composer from terminal as it is intended:
```shell
composer --version
```

## Build this project

To compile this project use:
```shell
GOOS=windows GOARCH=amd64 go build -o pvm.exe
```
