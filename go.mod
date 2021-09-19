module github.com/xo/usql

go 1.17

require (
	cloud.google.com/go/bigquery v1.22.0 // indirect
	cloud.google.com/go/spanner v1.25.0 // indirect
	github.com/ClickHouse/clickhouse-go v1.4.8
	github.com/DATA-DOG/go-sqlmock v1.5.0 // indirect
	github.com/IBM/nzgo v0.0.0-20210406171630-186d127e2795
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/MichaelS11/go-cql-driver v0.1.1
	github.com/SAP/go-hdb v0.105.3
	github.com/VoltDB/voltdb-client-go v1.0.1
	github.com/alecthomas/chroma v0.9.2
	github.com/alecthomas/kingpin v2.2.6+incompatible
	github.com/alecthomas/repr v0.0.0-20181024024818-d37bc2a10ba1 // indirect
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15 // indirect
	github.com/alexbrainman/odbc v0.0.0-20210605012845-39f8520b0d5f
	github.com/amsokol/ignite-go-client v0.12.2
	github.com/apache/arrow/go/arrow v0.0.0-20210901201644-e9eeff1c9297 // indirect
	github.com/apache/calcite-avatica-go/v5 v5.0.0
	github.com/apache/thrift v0.14.2 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/aws/aws-sdk-go v1.40.45 // indirect
	github.com/beltran/gohive v1.5.1 // indirect
	github.com/beltran/gosasl v0.0.0-20210629234946-b41ac5bb612a // indirect
	github.com/bippio/go-impala v2.1.0+incompatible
	github.com/btnguyen2k/consu/semita v0.1.5 // indirect
	github.com/btnguyen2k/gocosmos v0.1.4
	github.com/cloudspannerecosystem/go-sql-spanner v0.0.0-20210917063549-4e7fa9746975
	github.com/couchbase/go-couchbase v0.1.0 // indirect
	github.com/couchbase/go_n1ql v0.0.0-20160215142504-6cf4e348b127
	github.com/couchbase/gomemcached v0.1.3 // indirect
	github.com/couchbase/goutils v0.1.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/creack/pty v1.1.15 // indirect
	github.com/denisenkom/go-mssqldb v0.10.0
	github.com/docker/docker v20.10.6+incompatible
	github.com/exasol/exasol-driver-go v0.0.0-20210823115457-bc7df4ac05ae
	github.com/fatih/color v1.12.0 // indirect
	github.com/genjidb/genji v0.13.0
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/go-openapi/errors v0.20.1 // indirect
	github.com/go-openapi/strfmt v0.20.2 // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/go-zookeeper/zk v1.0.2 // indirect
	github.com/gocql/gocql v0.0.0-20210817081954-bc256bbb90de
	github.com/godror/godror v0.25.3
	github.com/gohxs/readline v0.0.0-20171011095936-a780388e6e7c
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/flatbuffers v2.0.0+incompatible // indirect
	github.com/google/goexpect v0.0.0-20210430020637-ab937bf7fd6f
	github.com/googleapis/gax-go/v2 v2.1.1 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/jackc/pgconn v1.10.0
	github.com/jackc/pgx/v4 v4.13.0
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jmrobles/h2go v0.5.0
	github.com/lib/pq v1.10.3
	github.com/mattn/go-adodb v0.0.1
	github.com/mattn/go-isatty v0.0.14
	github.com/mattn/go-runewidth v0.0.13
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.4.2 // indirect
	github.com/mithrandie/csvq v1.15.2
	github.com/mithrandie/csvq-driver v1.4.3
	github.com/mithrandie/go-text v1.4.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/nakagami/firebirdsql v0.9.2
	github.com/ory/dockertest/v3 v3.6.3
	github.com/pierrec/lz4 v2.2.6+incompatible // indirect
	github.com/pkg/browser v0.0.0-20210706143420-7d21f8c997e2 // indirect
	github.com/prestodb/presto-go-client v0.0.0-20201204133205-8958eb37e584
	github.com/sijms/go-ora/v2 v2.2.9
	github.com/sirupsen/logrus v1.8.1
	github.com/snowflakedb/gosnowflake v1.6.1
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/thda/tds v0.1.7
	github.com/trinodb/trino-go-client v0.300.0
	github.com/twmb/murmur3 v1.1.6 // indirect
	github.com/uber-go/tally v3.4.2+incompatible // indirect
	github.com/uber/athenadriver v1.1.13
	github.com/urfave/cli v1.22.5 // indirect
	github.com/vertica/vertica-sql-go v1.1.1
	github.com/vmihailenco/msgpack/v5 v5.3.4 // indirect
	github.com/xo/dburl v0.9.0
	github.com/xo/tblfmt v0.7.5
	github.com/xo/terminfo v0.0.0-20210125001918-ca9a967f8778
	github.com/yookoala/realpath v1.0.0
	github.com/zaf/temp v0.0.0-20170209143821-94e385923345
	github.com/ziutek/mymysql v1.5.4
	go.etcd.io/bbolt v1.3.6 // indirect
	go.mongodb.org/mongo-driver v1.7.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.1 // indirect
	golang.org/x/crypto v0.0.0-20210915214749-c084706c2272 // indirect
	golang.org/x/mod v0.5.0 // indirect
	golang.org/x/net v0.0.0-20210917221730-978cfadd31cf // indirect
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f // indirect
	golang.org/x/sys v0.0.0-20210917161153-d61c044b1678 // indirect
	golang.org/x/term v0.0.0-20210916214954-140adaaadfaf // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20210917145530-b395a37504d4 // indirect
	gorm.io/driver/bigquery v1.0.16
	modernc.org/ccgo/v3 v3.11.3 // indirect
	modernc.org/libc v1.11.6 // indirect
	modernc.org/ql v1.4.0
	modernc.org/sqlite v1.13.1
	sqlflow.org/gohive v0.0.0-20200521083454-ed52ee669b84
	sqlflow.org/gomaxcompute v0.0.0-20210805062559-c14ae028b44c
)

require (
	cloud.google.com/go v0.94.1 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/azure-storage-blob-go v0.14.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.17-0.20210211115548-6eac466e5fa3 // indirect
	github.com/Microsoft/hcsshim v0.8.16 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/aws/aws-sdk-go-v2 v1.9.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.5.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.6.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.14.0 // indirect
	github.com/aws/smithy-go v1.8.0 // indirect
	github.com/beltran/gssapi v0.0.0-20200324152954-d86554db4bab // indirect
	github.com/btnguyen2k/consu/gjrc v0.1.1 // indirect
	github.com/btnguyen2k/consu/olaf v0.1.3 // indirect
	github.com/btnguyen2k/consu/reddo v0.1.7 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cenkalti/backoff/v3 v3.0.0 // indirect
	github.com/cloudflare/golz4 v0.0.0-20150217214814-ef862a3cdc58 // indirect
	github.com/containerd/cgroups v0.0.0-20210114181951-8a68de567b68 // indirect
	github.com/containerd/containerd v1.5.0-beta.4 // indirect
	github.com/containerd/continuity v0.0.0-20210208174643-50096c924a4e // indirect
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964 // indirect
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible // indirect
	github.com/go-logfmt/logfmt v0.5.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/goterm v0.0.0-20200907032337-555d40f16ae2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.1.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.8.1 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jcmturner/goidentity/v6 v6.0.1 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jedib0t/go-pretty v4.3.0+incompatible // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.13.5 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-ieproxy v0.0.1 // indirect
	github.com/mithrandie/go-file/v2 v2.0.2 // indirect
	github.com/mithrandie/readline-csvq v1.1.1 // indirect
	github.com/mithrandie/ternary v1.1.0 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.4.1 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/nathan-fiscaletti/consolesize-go v0.0.0-20210105204122-a87d9f614b9d // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v1.0.0-rc93 // indirect
	github.com/pierrec/lz4/v4 v4.1.8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/unchartedsoftware/witch v0.0.0-20200617171400-4f405404126f // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xinsnake/go-http-digest-auth-client v0.6.0 // indirect
	github.com/xwb1989/sqlparser v0.0.0-20180606152119-120387863bf2 // indirect
	gitlab.com/nyarla/go-crypt v0.0.0-20160106005555-d9a5dc2b789b // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/tools v0.1.6 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/api v0.57.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/grpc v1.40.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/jcmturner/aescts.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/dnsutils.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/gokrb5.v6 v6.1.1 // indirect
	gopkg.in/jcmturner/rpc.v1 v1.1.0 // indirect
	lukechampine.com/uint128 v1.1.1 // indirect
	modernc.org/b v1.0.2 // indirect
	modernc.org/cc/v3 v3.35.0 // indirect
	modernc.org/db v1.0.3 // indirect
	modernc.org/file v1.0.3 // indirect
	modernc.org/fileutil v1.0.0 // indirect
	modernc.org/golex v1.0.1 // indirect
	modernc.org/internal v1.0.2 // indirect
	modernc.org/lldb v1.0.2 // indirect
	modernc.org/mathutil v1.4.1 // indirect
	modernc.org/memory v1.0.5 // indirect
	modernc.org/opt v0.1.1 // indirect
	modernc.org/sortutil v1.1.0 // indirect
	modernc.org/strutil v1.1.1 // indirect
	modernc.org/token v1.0.0 // indirect
	modernc.org/zappy v1.0.3 // indirect
)
