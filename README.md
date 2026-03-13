# PVM for Windows

[Support this project](https://github.com/sponsors/hjbdev)

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
pvm list remote
```
Will list the PHP versions available for installation.

```
pvm bin
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

```
pvm extensions list
```
Will show regular and Zend extensions for the active PHP version, including whether each extension is enabled, disabled, available in `ext`, or missing from disk.

```
pvm extensions enable curl,openssl
```
Will enable one or more extensions that already have entries in the active version's `php.ini`.

```
pvm extensions disable xdebug
```
Will disable an extension or Zend extension in the active version's `php.ini`.

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
bash ./build.sh
```

To override the embedded version for a release-style local build:
```shell
VERSION=1.2.1 bash ./build.sh
```

GitHub releases are built automatically from pushed tags and publish both `pvm.exe` and `install.ps1` as release assets.
