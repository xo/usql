module github.com/xo/usql

go 1.24

toolchain go1.24.0

require (
	github.com/ClickHouse/clickhouse-go/v2 v2.34.0
	github.com/IBM/nzgo/v12 v12.0.9
	github.com/MichaelS11/go-cql-driver v0.1.1
	github.com/SAP/go-hdb v1.13.6
	github.com/VoltDB/voltdb-client-go v1.0.16
	github.com/alecthomas/chroma/v2 v2.17.2
	github.com/alexbrainman/odbc v0.0.0-20250425050419-fd97e4030c76
	github.com/aliyun/aliyun-tablestore-go-sql-driver v0.0.0-20220418015234-4d337cb3eed9
	github.com/amsokol/ignite-go-client v0.12.2
	github.com/apache/arrow/go/v17 v17.0.0
	github.com/apache/calcite-avatica-go/v5 v5.4.0
	github.com/btnguyen2k/gocosmos v1.1.0
	github.com/btnguyen2k/godynamo v1.3.0
	github.com/chaisql/chai v0.16.1-0.20240218103834-23e406360fd2
	github.com/couchbase/go_n1ql v0.0.0-20220303011133-0ed4bf93e31d
	github.com/databricks/databricks-sql-go v1.7.0
	github.com/datafuselabs/databend-go v0.7.4
	github.com/docker/docker v28.1.1+incompatible
	github.com/exasol/exasol-driver-go v1.0.13
	github.com/go-git/go-billy/v5 v5.6.2
	github.com/go-sql-driver/mysql v1.9.2
	github.com/gocql/gocql v1.7.0
	github.com/godror/godror v0.48.2
	github.com/gohxs/readline v0.0.0-20171011095936-a780388e6e7c
	github.com/google/go-cmp v0.7.0
	github.com/google/goexpect v0.0.0-20210430020637-ab937bf7fd6f
	github.com/googleapis/go-sql-spanner v1.13.1
	github.com/jackc/pgx/v5 v5.7.4
	github.com/jeandeaual/go-locale v0.0.0-20250421151639-a9d6ed1b3d45
	github.com/jmrobles/h2go v0.5.0
	github.com/kenshaw/colors v0.2.1
	github.com/kenshaw/rasterm v0.1.13
	github.com/lib/pq v1.10.9
	github.com/marcboeker/go-duckdb/v2 v2.2.1
	github.com/mattn/go-adodb v0.0.1
	github.com/mattn/go-isatty v0.0.20
	github.com/mattn/go-sqlite3 v1.14.28
	github.com/microsoft/go-mssqldb v1.8.1
	github.com/mithrandie/csvq v1.18.1
	github.com/mithrandie/csvq-driver v1.7.0
	github.com/nakagami/firebirdsql v0.9.15
	github.com/ory/dockertest/v3 v3.12.0
	github.com/prestodb/presto-go-client v0.0.0-20240426182841-905ac40a1783
	github.com/proullon/ramsql v0.1.4
	github.com/sclgo/impala-go v1.2.0
	github.com/sijms/go-ora/v2 v2.8.24
	github.com/snowflakedb/gosnowflake v1.14.0
	github.com/spf13/cobra v1.9.1
	github.com/spf13/pflag v1.0.6
	github.com/spf13/viper v1.20.1
	github.com/thda/tds v0.1.7
	github.com/trinodb/trino-go-client v0.323.0
	github.com/uber/athenadriver v1.1.15
	github.com/vertica/vertica-sql-go v1.3.3
	github.com/xo/dburl v0.23.7
	github.com/xo/tblfmt v0.15.1
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e
	github.com/ydb-platform/ydb-go-sdk/v3 v3.108.2
	github.com/yookoala/realpath v1.0.0
	github.com/ziutek/mymysql v1.5.4
	golang.org/x/exp v0.0.0-20250506013437-ce4c2cf36ca6
	gorm.io/driver/bigquery v1.2.0
	modernc.org/ql v1.4.13
	modernc.org/sqlite v1.37.0
	sqlflow.org/gohive v0.0.0-20240730014249-8960223660e2
	sqlflow.org/gomaxcompute v0.0.0-20210805062559-c14ae028b44c
)

require (
	cel.dev/expr v0.24.0 // indirect
	cloud.google.com/go v0.121.0 // indirect
	cloud.google.com/go/auth v0.16.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/bigquery v1.67.0 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	cloud.google.com/go/iam v1.5.2 // indirect
	cloud.google.com/go/longrunning v0.6.7 // indirect
	cloud.google.com/go/monitoring v1.24.2 // indirect
	cloud.google.com/go/spanner v1.81.0 // indirect
	dario.cat/mergo v1.0.1 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.18.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.9.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.6.1 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.4.2 // indirect
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/ClickHouse/ch-go v0.65.1 // indirect
	github.com/DATA-DOG/go-sqlmock v1.5.2 // indirect
	github.com/DataDog/zstd v1.5.7 // indirect
	github.com/GoogleCloudPlatform/grpc-gcp-go/grpcgcp v1.5.2 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.27.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/aliyun/aliyun-tablestore-go-sdk v1.7.17 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/apache/arrow-go/v18 v18.3.0 // indirect
	github.com/apache/arrow/go/v12 v12.0.1 // indirect
	github.com/apache/arrow/go/v15 v15.0.2 // indirect
	github.com/apache/thrift v0.21.0 // indirect
	github.com/avast/retry-go v3.0.0+incompatible // indirect
	github.com/aws/aws-sdk-go v1.55.7 // indirect
	github.com/aws/aws-sdk-go-v2 v1.36.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.67 // indirect
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.19.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.75 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.43.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.10.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.79.3 // indirect
	github.com/aws/smithy-go v1.22.3 // indirect
	github.com/beltran/gohive v1.8.0 // indirect
	github.com/beltran/gosasl v1.0.0 // indirect
	github.com/beltran/gssapi v0.0.0-20200324152954-d86554db4bab // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitfield/gotestdox v0.2.2 // indirect
	github.com/btnguyen2k/consu/checksum v1.1.1 // indirect
	github.com/btnguyen2k/consu/g18 v0.1.0 // indirect
	github.com/btnguyen2k/consu/gjrc v0.2.2 // indirect
	github.com/btnguyen2k/consu/olaf v0.1.3 // indirect
	github.com/btnguyen2k/consu/reddo v0.1.9 // indirect
	github.com/btnguyen2k/consu/semita v0.1.5 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20250501225837-2ac532fd4443 // indirect
	github.com/cockroachdb/errors v1.12.0 // indirect
	github.com/cockroachdb/fifo v0.0.0-20240816210425-c5d0cb0b6fc0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20241215232642-bb51bb14a506 // indirect
	github.com/cockroachdb/pebble v1.1.5 // indirect
	github.com/cockroachdb/redact v1.1.6 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20250429170803-42689b6311bb // indirect
	github.com/containerd/continuity v0.4.5 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/coreos/go-oidc/v3 v3.14.1 // indirect
	github.com/couchbase/go-couchbase v0.1.1 // indirect
	github.com/couchbase/gomemcached v0.3.3 // indirect
	github.com/couchbase/goutils v0.1.2 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/dnephin/pflag v1.0.7 // indirect
	github.com/docker/cli v28.0.4+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/duckdb/duckdb-go-bindings v0.1.14 // indirect
	github.com/duckdb/duckdb-go-bindings/darwin-amd64 v0.1.9 // indirect
	github.com/duckdb/duckdb-go-bindings/darwin-arm64 v0.1.9 // indirect
	github.com/duckdb/duckdb-go-bindings/linux-amd64 v0.1.9 // indirect
	github.com/duckdb/duckdb-go-bindings/linux-arm64 v0.1.9 // indirect
	github.com/duckdb/duckdb-go-bindings/windows-amd64 v0.1.9 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/dvsekhvalnov/jose2go v1.8.0 // indirect
	github.com/edsrzf/mmap-go v1.2.0 // indirect
	github.com/elastic/go-sysinfo v1.15.3 // indirect
	github.com/elastic/go-windows v1.0.2 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.32.4 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/exasol/error-reporting-go v0.2.0 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.9 // indirect
	github.com/getsentry/sentry-go v0.32.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-jose/go-jose/v4 v4.1.0 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/go-zookeeper/zk v1.0.4 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/godror/knownpb v0.2.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/golang-module/carbon/v2 v2.4.1 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/mock v1.7.0-rc.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/flatbuffers/go v0.0.0-20230110200425-62e4d2e5b215 // indirect
	github.com/google/goterm v0.0.0-20200907032337-555d40f16ae2 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/icholy/digest v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/goidentity/v6 v6.0.1 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jedib0t/go-pretty/v6 v6.6.7 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/klauspost/asmfmt v1.3.2 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/marcboeker/go-duckdb/arrowmapping v0.0.7 // indirect
	github.com/marcboeker/go-duckdb/mapping v0.0.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mattn/go-sixel v0.0.5 // indirect
	github.com/minio/asm2plan9s v0.0.0-20200509001527-cdd76441f9d8 // indirect
	github.com/minio/c2goasm v0.0.0-20190812172519-36a3d3bbc4f3 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mithrandie/go-file/v2 v2.1.0 // indirect
	github.com/mithrandie/go-text v1.6.0 // indirect
	github.com/mithrandie/ternary v1.1.1 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/go-archive v0.1.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/murfffi/gorich v0.2.0 // indirect
	github.com/nakagami/chacha20 v0.1.0 // indirect
	github.com/nathan-fiscaletti/consolesize-go v0.0.0-20220204101620-317176b6684d // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opencontainers/runc v1.2.6 // indirect
	github.com/paulmach/orb v0.11.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.63.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	github.com/sagikazarmark/locafero v0.9.0 // indirect
	github.com/samber/lo v1.50.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/soniakeys/quant v1.0.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/afero v1.14.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spiffe/go-spiffe/v2 v2.5.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/twmb/murmur3 v1.1.8 // indirect
	github.com/uber-go/tally v3.5.10+incompatible // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xwb1989/sqlparser v0.0.0-20180606152119-120387863bf2 // indirect
	github.com/ydb-platform/ydb-go-genproto v0.0.0-20241112172322-ea1f63298f77 // indirect
	github.com/zeebo/errs v1.4.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	gitlab.com/nyarla/go-crypt v0.0.0-20160106005555-d9a5dc2b789b // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.35.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.60.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/api v0.232.0 // indirect
	google.golang.org/genproto v0.0.0-20250505200425-f936aa4a68b2 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250505200425-f936aa4a68b2 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250505200425-f936aa4a68b2 // indirect
	google.golang.org/grpc v1.72.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/jcmturner/aescts.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/dnsutils.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/gokrb5.v6 v6.1.1 // indirect
	gopkg.in/jcmturner/rpc.v1 v1.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/gotestsum v1.12.2 // indirect
	howett.net/plist v1.0.1 // indirect
	modernc.org/b v1.1.0 // indirect
	modernc.org/db v1.0.14 // indirect
	modernc.org/file v1.0.10 // indirect
	modernc.org/fileutil v1.3.1 // indirect
	modernc.org/golex v1.1.0 // indirect
	modernc.org/internal v1.1.1 // indirect
	modernc.org/libc v1.65.3 // indirect
	modernc.org/lldb v1.0.8 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.10.0 // indirect
	modernc.org/sortutil v1.2.1 // indirect
	modernc.org/strutil v1.2.1 // indirect
	modernc.org/zappy v1.1.0 // indirect
)
