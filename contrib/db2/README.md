# Configuring db2

1. Install unixodbc:

```sh
$ sudo aptitude install unixodbc unixodbc-bin unixodbc-dev
```

2. Download `dsdriver` and install:

```sh
$ ls ~/Downloads/ibm_data_server_driver_package_linuxx64_v11.5.tar.gz
/home/ken/Downloads/ibm_data_server_driver_package_linuxx64_v11.5.tar.gz
$ sudo ./install-dsdriver.sh
```

3. Copy ODBC and CLI configs:

```sh
$ cat odbcinst.ini | sudo tee -a /etc/odbcinst.ini
$ sudo cp {db2cli.ini,db2dsdriver.cfg} /opt/db2/clidriver/cfg/
```

4. Run DB2 docker image:

```sh
$ ../docker-run.sh db2 -u
```

5. Verify DB2 working:

```sh
$ ./db2cli-validate.sh
```
