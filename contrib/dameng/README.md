# Dameng DM8

Dameng publishes no DM8 image to a public container registry. The container
archive it does publish is years behind, so `Containerfile` builds a current DM8
from Dameng's own Ubuntu 22 installer instead.

`podman-run.sh` builds the image itself, because it runs `podman build` for any
target directory that holds a `Containerfile`. No manual download step is
needed.

## Start it

    ../podman-run.sh dameng

The first build downloads an 820 MB installer and takes several minutes. The
finished image is about 2.9 GB. Later runs reuse it.

## Connect

    usql dameng://SYSDBA:P4ssw0rd@localhost:5236/

The `dm` and `dm8` schemes are aliases of `dameng` and reach the same driver.

## Build arguments

Dameng removes old releases, so the default `DM_URL` stops working over time.
When it does, find the current file on the download page at
https://eco.dameng.com/download/ and build with the new one:

    podman build \
      --build-arg DM_URL=https://download.dameng.com/eco/adapter/DM8/<YYYYMM>/<file>.zip \
      --build-arg DM_ISO=<file>.iso \
      --build-arg DM_SHA256=<sha256 of the iso> \
      -t localhost/dm8 .

`DM_PASSWORD`, `DM_PORT`, `DM_NAME` and `DM_INSTANCE` are also build arguments.
The password must be 8 to 48 characters and must contain an uppercase letter, a
lowercase letter and a digit. The installer rejects anything else, which is why
the default is `P4ssw0rd` rather than Dameng's usual `SYSDBA001`.

## How the build works

The installer ships as an ISO inside a zip. The build unzips it, checks the
ISO against its published SHA256, and reads `DMInstall.bin` out of the ISO with
`bsdtar`. `bsdtar` reads ISO 9660 directly, so the build needs no loop mount and
no privileged mode.

The install runs in silent mode from a generated configuration file, with
`INIT_DB` set to `Y` so the installer also creates the instance. The database
uses UTF-8 and 16K pages. A second stage copies the finished `/opt/dmdbms` into
a clean Ubuntu image, which keeps the installer and the build tools out of the
result.

Three details cost time when this was first written, and they are kept here so
the next person does not repeat them. `SYSAUDITOR_PWD` is required even though
the template shows it empty. `CREATE_DB_SERVICE` and `STARTUP_DB_SERVICE` are
children of `DATABASE`, not of `DB_PARAMS`, and the installer rejects the file
if they sit in the wrong place. The installer also refuses to write into an
install directory that already exists with content, so the build must let it
create `/opt/dmdbms` itself.

## Architecture

Dameng publishes x86-64 installers only. On an arm64 host the container runs
under emulation, if it runs at all.

## License

DM8 is proprietary software, and its license restricts redistribution. This
build downloads the installer from Dameng at build time and produces a local
image, so nothing is redistributed. Do not push the built image to a public
registry, and do not use the repackaged DM8 images that appear on public
registries, because they redistribute the product against its license and some
carry authorization certificates that expired years ago.
