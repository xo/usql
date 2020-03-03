# Oracle Notes

1. Clone:

```sh
$ mkdir -p ~/src/oracle && cd ~/src/oracle
$ git clone https://github.com/oracle/docker-images.git
```

2. Download `LINUX.X64_193000_db_home.zip` -> `~/src/oracle/docker-images/OracleDatabase/SingleInstance/dockerfiles/19.3.0/`:

```sh
$ mv ~/Downloads/LINUX.X64_193000_db_home.zip ~/src/oracle/docker-images/OracleDatabase/SingleInstance/dockerfiles/19.3.0/
```
3. Build:

```sh
$ cd ~/src/oracle/docker-images/OracleDatabase/SingleInstance/dockerfiles/
$ ./buildDockerImage.sh -v 19.3.0 -e
```

4. Fix ["out-of-band" issue][out-of-band-issue] ([see also GitHub issue][out-of-band-github]) by adding `userland-proxy: false` to `/etc/docker/daemon.json`

[out-of-band-issue]: https://medium.com/@FranckPachot/19c-instant-client-and-docker-1566630ab20e
[out-of-band-github]: https://github.com/oracle/docker-images/issues/1352

5. Create storage directory, and change ownership:

```sh
$ mkdir -p /media/src/opt/oracle
$ sudo chown -R 54321:54321 /media/src/opt/oracle
```

6. Start:

```sh
$ cd $GOPATH/src/github.com/xo/usql/contrib
$ ./docker-run.sh oracle
```
