module github.com/xo/usql

require (
	cloud.google.com/go v0.78.0 // indirect
	cloud.google.com/go/bigquery v1.16.0 // indirect
	cloud.google.com/go/spanner v1.15.0 // indirect
	github.com/ClickHouse/clickhouse-go v1.4.3
	github.com/DATA-DOG/go-sqlmock v1.5.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/MichaelS11/go-cql-driver v0.1.1
	github.com/Microsoft/hcsshim v0.8.15 // indirect
	github.com/SAP/go-hdb v0.102.7
	github.com/VoltDB/voltdb-client-go v1.0.1
	github.com/alecthomas/chroma v0.8.2
	github.com/alecthomas/kingpin v2.2.6+incompatible
	github.com/alecthomas/repr v0.0.0-20181024024818-d37bc2a10ba1 // indirect
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15 // indirect
	github.com/alexbrainman/odbc v0.0.0-20200426075526-f0492dfa1575
	github.com/amsokol/ignite-go-client v0.12.2
	github.com/apache/arrow/go/arrow v0.0.0-20210305224846-906331c47338 // indirect
	github.com/apache/calcite-avatica-go/v5 v5.0.0
	github.com/apache/thrift v0.14.1 // indirect
	github.com/aws/aws-sdk-go v1.37.25 // indirect
	github.com/beltran/gohive v1.4.0 // indirect
	github.com/beltran/gosasl v0.0.0-20210215125809-4fa075701386 // indirect
	github.com/bippio/go-impala v2.1.0+incompatible
	github.com/btnguyen2k/consu/semita v0.1.5 // indirect
	github.com/btnguyen2k/gocosmos v0.1.3
	github.com/couchbase/go-couchbase v0.0.0-20210301172442-553722772724 // indirect
	github.com/couchbase/go_n1ql v0.0.0-20160215142504-6cf4e348b127
	github.com/couchbase/gomemcached v0.1.2 // indirect
	github.com/couchbase/goutils v0.0.0-20210118111533-e33d3ffb5401 // indirect
	github.com/denisenkom/go-mssqldb v0.9.0
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/docker/docker v20.10.3+incompatible
	github.com/fatih/color v1.10.0 // indirect
	github.com/genjidb/genji v0.10.1
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/go-openapi/errors v0.20.0 // indirect
	github.com/go-openapi/strfmt v0.20.0 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-zookeeper/zk v1.0.2 // indirect
	github.com/gocql/gocql v0.0.0-20210303210847-f18e0979d243
	github.com/godror/godror v0.24.2
	github.com/gohxs/readline v0.0.0-20171011095936-a780388e6e7c
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/flatbuffers v1.12.0 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/jackc/pgconn v1.8.0
	github.com/jackc/pgproto3/v2 v2.0.7 // indirect
	github.com/jackc/pgx/v4 v4.10.1
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jmrobles/h2go v0.5.0
	github.com/lib/pq v1.9.0
	github.com/magefile/mage v1.11.0 // indirect
	github.com/mattn/go-adodb v0.0.1
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/go-runewidth v0.0.10 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/nakagami/firebirdsql v0.9.0
	github.com/ory/dockertest/v3 v3.6.3
	github.com/pierrec/lz4 v2.2.6+incompatible // indirect
	github.com/pkg/browser v0.0.0-20210115035449-ce105d075bb4 // indirect
	github.com/prestodb/presto-go-client v0.0.0-20201204133205-8958eb37e584
	github.com/rakyll/go-sql-driver-spanner v0.0.0-20200507191418-c013a6449778
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sijms/go-ora v0.0.0-20210301201657-604392efe05f
	github.com/sirupsen/logrus v1.8.0
	github.com/snowflakedb/gosnowflake v1.4.1
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/thda/tds v0.1.7
	github.com/trinodb/trino-go-client v0.300.0
	github.com/uber/athenadriver v1.1.12
	github.com/urfave/cli v1.22.5 // indirect
	github.com/vertica/vertica-sql-go v1.1.1
	github.com/vmihailenco/msgpack/v5 v5.2.0 // indirect
	github.com/xo/dburl v0.3.0
	github.com/xo/tblfmt v0.1.0
	github.com/xo/terminfo v0.0.0-20210125001918-ca9a967f8778
	github.com/xo/xoutil v0.0.0-20171112033149-46189f4026a5
	github.com/zaf/temp v0.0.0-20170209143821-94e385923345
	github.com/ziutek/mymysql v1.5.4
	go.mongodb.org/mongo-driver v1.4.6 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	golang.org/x/oauth2 v0.0.0-20210220000619-9bb904979d93 // indirect
	golang.org/x/sys v0.0.0-20210305230114-8fe3ee5dd75b // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/genproto v0.0.0-20210303154014-9728d6b83eeb // indirect
	google.golang.org/grpc v1.36.0 // indirect
	gorm.io/driver/bigquery v1.0.16
	modernc.org/b v1.0.1 // indirect
	modernc.org/db v1.0.1 // indirect
	modernc.org/file v1.0.2 // indirect
	modernc.org/golex v1.0.1 // indirect
	modernc.org/lldb v1.0.1 // indirect
	modernc.org/ql v1.3.1
	modernc.org/sqlite v1.8.8
	modernc.org/zappy v1.0.2 // indirect
	sqlflow.org/gohive v0.0.0-20200521083454-ed52ee669b84
	sqlflow.org/gomaxcompute v0.0.0-20200410041603-30fa752b7593
)

go 1.16
