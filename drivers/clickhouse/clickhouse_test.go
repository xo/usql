package clickhouse_test

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	dt "github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	"github.com/xo/usql/drivers/clickhouse"
	"github.com/xo/usql/drivers/metadata"
)

type Database struct {
	BuildArgs  []dc.BuildArg
	RunOptions *dt.RunOptions
	Exec       []string
	Driver     string
	URL        string
	DockerPort string
	Resource   *dt.Resource
	DB         *sql.DB
	Opts       []metadata.ReaderOption
	Reader     metadata.BasicReader
}

var dbName string = "clickhouse"

var db = Database{
	BuildArgs: []dc.BuildArg{
		{Name: "BASE_IMAGE", Value: "yandex/clickhouse-server:21.7.4.18"},
		{Name: "SCHEMA_URL", Value: "https://raw.githubusercontent.com/jgryko5/usql/clickhouse_driver/drivers/clickhouse/testdata/clickhouse.sql"},
		{Name: "TARGET", Value: "/docker-entrypoint-initdb.d/"},
	},
	RunOptions: &dt.RunOptions{
		Name: "usql-clickhouse-server",
	},
	Driver:     dbName,
	URL:        "clickhouse://127.0.0.1:%s",
	DockerPort: "9000/tcp",
}

func TestMain(m *testing.M) {
	cleanup := true
	flag.BoolVar(&cleanup, "cleanup", true, "delete containers when finished")
	flag.Parse()
	pool, err := dt.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	var ok bool
	db.Resource, ok = pool.ContainerByName(db.RunOptions.Name)
	if !ok {
		buildOpts := &dt.BuildOptions{
			ContextDir: "../testdata/docker",
			BuildArgs:  db.BuildArgs,
		}
		db.Resource, err = pool.BuildAndRunWithBuildOptions(buildOpts, db.RunOptions)
		if err != nil {
			log.Fatal("Could not start resource: ", err)
		}
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		hostPort := db.Resource.GetPort(db.DockerPort)
		var err error
		db.DB, err = sql.Open(db.Driver, fmt.Sprintf(db.URL, hostPort))
		if err != nil {
			return err
		}
		return db.DB.Ping()
	}); err != nil {
		log.Fatal("Timed out waiting for db: ", err)
	}
	db.Reader = clickhouse.NewMetadataReader(db.DB).(metadata.BasicReader)

	if len(db.Exec) != 0 {
		exitCode, err := db.Resource.Exec(db.Exec, dt.ExecOptions{
			StdIn:  os.Stdin,
			StdOut: os.Stdout,
			StdErr: os.Stderr,
			TTY:    true,
		})
		if err != nil || exitCode != 0 {
			log.Fatal("Could not load schema: ", err)
		}
	}
	code := m.Run()
	// You can't defer this because os.Exit doesn't care for defer
	if cleanup {
		if err := pool.Purge(db.Resource); err != nil {
			log.Fatal("Could not purge resource: ", err)
		}
	}
	os.Exit(code)
}

func TestSchemas(t *testing.T) {
	expected := "default, system, tutorial"
	r := db.Reader

	result, err := r.Schemas(metadata.Filter{WithSystem: true})
	if err != nil {
		log.Fatalf("Could not read %s schemas: %v", dbName, err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Schema)
	}
	actual := strings.Join(names, ", ")
	if actual != expected {
		t.Errorf("Wrong %s schema names, expected:\n  %v\ngot:\n  %v", dbName, expected, names)
	}

}

func TestTables(t *testing.T) {
	schema := "tutorial"
	expected := "hits_v1, visits_v1"
	r := db.Reader

	result, err := r.Tables(metadata.Filter{Schema: schema, Types: []string{"BASE TABLE", "TABLE", "VIEW"}})
	if err != nil {
		log.Fatalf("Could not read %s tables: %v", dbName, err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Name)
	}
	actual := strings.Join(names, ", ")
	if actual != expected {
		t.Errorf("Wrong %s table names, expected:\n  '%v'\ngot:\n  '%v'", dbName, expected, names)
	}

}

func TestFunctions(t *testing.T) {
	schema := "tutorial"
	expected := "BIT_AND, BIT_OR, BIT_XOR, CAST, CHARACTER_LENGTH, CHAR_LENGTH, COVAR_POP, COVAR_SAMP, CRC32, CRC32IEEE, CRC64, DATABASE, DATE, DAY, DAYOFMONTH, DAYOFWEEK, DAYOFYEAR, FQDN, FROM_BASE64, FROM_UNIXTIME, HOUR, INET6_ATON, INET6_NTOA, INET_ATON, INET_NTOA, IPv4CIDRToRange, IPv4NumToString, IPv4NumToStringClassC, IPv4StringToNum, IPv4ToIPv6, IPv6CIDRToRange, IPv6NumToString, IPv6StringToNum, JSONExtract, JSONExtractArrayRaw, JSONExtractBool, JSONExtractFloat, JSONExtractInt, JSONExtractKeysAndValues, JSONExtractKeysAndValuesRaw, JSONExtractRaw, JSONExtractString, JSONExtractUInt, JSONHas, JSONKey, JSONLength, JSONType, MACNumToString, MACStringToNum, MACStringToOUI, MD5, MINUTE, MONTH, QUARTER, SECOND, SHA1, SHA224, SHA256, STDDEV_POP, STDDEV_SAMP, TO_BASE64, URLHash, URLHierarchy, URLPathHierarchy, UUIDNumToString, UUIDStringToNum, VAR_POP, VAR_SAMP, YEAR, __bitBoolMaskAnd, __bitBoolMaskOr, __bitSwapLastTwo, __bitWrapperFunc, __getScalar, abs, accurateCast, accurateCastOrNull, acos, acosh, addDays, addHours, addMinutes, addMonths, addQuarters, addSeconds, addWeeks, addYears, addressToLine, addressToSymbol, aes_decrypt_mysql, aes_encrypt_mysql, aggThrow, alphaTokens, and, any, anyHeavy, anyLast, appendTrailingCharIfAbsent, argMax, argMin, array, arrayAUC, arrayAll, arrayAvg, arrayCompact, arrayConcat, arrayCount, arrayCumSum, arrayCumSumNonNegative, arrayDifference, arrayDistinct, arrayElement, arrayEnumerate, arrayEnumerateDense, arrayEnumerateDenseRanked, arrayEnumerateUniq, arrayEnumerateUniqRanked, arrayExists, arrayFill, arrayFilter, arrayFirst, arrayFirstIndex, arrayFlatten, arrayIntersect, arrayJoin, arrayMap, arrayMax, arrayMin, arrayPopBack, arrayPopFront, arrayProduct, arrayPushBack, arrayPushFront, arrayReduce, arrayReduceInRanges, arrayResize, arrayReverse, arrayReverseFill, arrayReverseSort, arrayReverseSplit, arraySlice, arraySort, arraySplit, arrayStringConcat, arraySum, arrayUniq, arrayWithConstant, arrayZip, asin, asinh, assumeNotNull, atan, atan2, atanh, avg, avgWeighted, bar, base64Decode, base64Encode, basename, bayesAB, bitAnd, bitCount, bitHammingDistance, bitNot, bitOr, bitPositionsToArray, bitRotateLeft, bitRotateRight, bitShiftLeft, bitShiftRight, bitTest, bitTestAll, bitTestAny, bitXor, bitmapAnd, bitmapAndCardinality, bitmapAndnot, bitmapAndnotCardinality, bitmapBuild, bitmapCardinality, bitmapContains, bitmapHasAll, bitmapHasAny, bitmapMax, bitmapMin, bitmapOr, bitmapOrCardinality, bitmapSubsetInRange, bitmapSubsetLimit, bitmapToArray, bitmapTransform, bitmapXor, bitmapXorCardinality, bitmaskToArray, bitmaskToList, blockNumber, blockSerializedSize, blockSize, boundingRatio, buildId, byteSize, caseWithExpr, caseWithExpression, caseWithoutExpr, caseWithoutExpression, categoricalInformationValue, cbrt, ceil, ceiling, char, cityHash64, coalesce, concat, concatAssumeInjective, connectionId, connection_id, convertCharset, corr, corrStable, cos, cosh, count, countDigits, countEqual, countMatches, countMatchesCaseInsensitive, countSubstrings, countSubstringsCaseInsensitive, countSubstringsCaseInsensitiveUTF8, covarPop, covarPopStable, covarSamp, covarSampStable, currentDatabase, currentUser, cutFragment, cutIPv6, cutQueryString, cutQueryStringAndFragment, cutToFirstSignificantSubdomain, cutToFirstSignificantSubdomainCustom, cutToFirstSignificantSubdomainCustomWithWWW, cutToFirstSignificantSubdomainWithWWW, cutURLParameter, cutWWW, dateDiff, dateName, dateTrunc, date_trunc, decodeURLComponent, decodeXMLComponent, decrypt, defaultValueOfArgumentType, defaultValueOfTypeName, deltaSum, deltaSumTimestamp, demangle, dense_rank, dictGet, dictGetChildren, dictGetDate, dictGetDateOrDefault, dictGetDateTime, dictGetDateTimeOrDefault, dictGetDescendants, dictGetFloat32, dictGetFloat32OrDefault, dictGetFloat64, dictGetFloat64OrDefault, dictGetHierarchy, dictGetInt16, dictGetInt16OrDefault, dictGetInt32, dictGetInt32OrDefault, dictGetInt64, dictGetInt64OrDefault, dictGetInt8, dictGetInt8OrDefault, dictGetOrDefault, dictGetOrNull, dictGetString, dictGetStringOrDefault, dictGetUInt16, dictGetUInt16OrDefault, dictGetUInt32, dictGetUInt32OrDefault, dictGetUInt64, dictGetUInt64OrDefault, dictGetUInt8, dictGetUInt8OrDefault, dictGetUUID, dictGetUUIDOrDefault, dictHas, dictIsIn, divide, domain, domainWithoutWWW, dumpColumnStructure, e, empty, emptyArrayDate, emptyArrayDateTime, emptyArrayFloat32, emptyArrayFloat64, emptyArrayInt16, emptyArrayInt32, emptyArrayInt64, emptyArrayInt8, emptyArrayString, emptyArrayToSingle, emptyArrayUInt16, emptyArrayUInt32, emptyArrayUInt64, emptyArrayUInt8, encodeXMLComponent, encrypt, endsWith, entropy, equals, erf, erfc, errorCodeToName, evalMLMethod, exp, exp10, exp2, extract, extractAll, extractAllGroups, extractAllGroupsHorizontal, extractAllGroupsVertical, extractGroups, extractTextFromHTML, extractURLParameter, extractURLParameterNames, extractURLParameters, farmFingerprint64, farmHash64, file, filesystemAvailable, filesystemCapacity, filesystemFree, finalizeAggregation, firstSignificantSubdomain, firstSignificantSubdomainCustom, first_value, flatten, floor, format, formatDateTime, formatReadableQuantity, formatReadableSize, formatReadableTimeDelta, formatRow, formatRowNoNewline, fragment, fromModifiedJulianDay, fromModifiedJulianDayOrNull, fromUnixTimestamp, fromUnixTimestamp64Micro, fromUnixTimestamp64Milli, fromUnixTimestamp64Nano, fullHostName, fuzzBits, gccMurmurHash, gcd, generateUUIDv4, geoDistance, geoToH3, geohashDecode, geohashEncode, geohashesInBox, getMacro, getSetting, getSizeOfEnumType, globalIn, globalInIgnoreSet, globalNotIn, globalNotInIgnoreSet, globalNotNullIn, globalNotNullInIgnoreSet, globalNullIn, globalNullInIgnoreSet, globalVariable, greatCircleAngle, greatCircleDistance, greater, greaterOrEquals, greatest, groupArray, groupArrayInsertAt, groupArrayMovingAvg, groupArrayMovingSum, groupArraySample, groupBitAnd, groupBitOr, groupBitXor, groupBitmap, groupBitmapAnd, groupBitmapOr, groupBitmapXor, groupUniqArray, h3EdgeAngle, h3EdgeLengthM, h3GetBaseCell, h3GetResolution, h3HexAreaM2, h3IndexesAreNeighbors, h3IsValid, h3ToChildren, h3ToParent, h3ToString, h3kRing, halfMD5, has, hasAll, hasAny, hasColumnInTable, hasSubstr, hasThreadFuzzer, hasToken, hasTokenCaseInsensitive, hex, histogram, hiveHash, hostName, hostname, hypot, identity, if, ifNotFinite, ifNull, ignore, ilike, in, inIgnoreSet, indexHint, indexOf, initializeAggregation, intDiv, intDivOrZero, intExp10, intExp2, intHash32, intHash64, intervalLengthSum, isConstant, isDecimalOverflow, isFinite, isIPAddressInRange, isIPv4String, isIPv6String, isInfinite, isNaN, isNotNull, isNull, isValidJSON, isValidUTF8, isZeroOrNull, javaHash, javaHashUTF16LE, joinGet, joinGetOrNull, jumpConsistentHash, kurtPop, kurtSamp, lagInFrame, last_value, lcase, lcm, leadInFrame, least, length, lengthUTF8, less, lessOrEquals, lgamma, like, ln, locate, log, log10, log1p, log2, logTrace, lowCardinalityIndices, lowCardinalityKeys, lower, lowerUTF8, mannWhitneyUTest, map, mapAdd, mapContains, mapKeys, mapPopulateSeries, mapSubtract, mapValues, match, materialize, max, maxIntersections, maxIntersectionsPosition, maxMap, median, medianBFloat16, medianDeterministic, medianExact, medianExactHigh, medianExactLow, medianExactWeighted, medianTDigest, medianTDigestWeighted, medianTiming, medianTimingWeighted, metroHash64, mid, min, minMap, minus, mod, modelEvaluate, modulo, moduloLegacy, moduloOrZero, multiFuzzyMatchAllIndices, multiFuzzyMatchAny, multiFuzzyMatchAnyIndex, multiIf, multiMatchAllIndices, multiMatchAny, multiMatchAnyIndex, multiSearchAllPositions, multiSearchAllPositionsCaseInsensitive, multiSearchAllPositionsCaseInsensitiveUTF8, multiSearchAllPositionsUTF8, multiSearchAny, multiSearchAnyCaseInsensitive, multiSearchAnyCaseInsensitiveUTF8, multiSearchAnyUTF8, multiSearchFirstIndex, multiSearchFirstIndexCaseInsensitive, multiSearchFirstIndexCaseInsensitiveUTF8, multiSearchFirstIndexUTF8, multiSearchFirstPosition, multiSearchFirstPositionCaseInsensitive, multiSearchFirstPositionCaseInsensitiveUTF8, multiSearchFirstPositionUTF8, multiply, murmurHash2_32, murmurHash2_64, murmurHash3_128, murmurHash3_32, murmurHash3_64, negate, neighbor, netloc, ngramDistance, ngramDistanceCaseInsensitive, ngramDistanceCaseInsensitiveUTF8, ngramDistanceUTF8, ngramMinHash, ngramMinHashArg, ngramMinHashArgCaseInsensitive, ngramMinHashArgCaseInsensitiveUTF8, ngramMinHashArgUTF8, ngramMinHashCaseInsensitive, ngramMinHashCaseInsensitiveUTF8, ngramMinHashUTF8, ngramSearch, ngramSearchCaseInsensitive, ngramSearchCaseInsensitiveUTF8, ngramSearchUTF8, ngramSimHash, ngramSimHashCaseInsensitive, ngramSimHashCaseInsensitiveUTF8, ngramSimHashUTF8, normalizeQuery, normalizeQueryKeepNames, normalizedQueryHash, normalizedQueryHashKeepNames, not, notEmpty, notEquals, notILike, notIn, notInIgnoreSet, notLike, notNullIn, notNullInIgnoreSet, now, now64, nullIf, nullIn, nullInIgnoreSet, or, parseDateTime32BestEffort, parseDateTime32BestEffortOrNull, parseDateTime32BestEffortOrZero, parseDateTime64BestEffort, parseDateTime64BestEffortOrNull, parseDateTime64BestEffortOrZero, parseDateTimeBestEffort, parseDateTimeBestEffortOrNull, parseDateTimeBestEffortOrZero, parseDateTimeBestEffortUS, parseDateTimeBestEffortUSOrNull, parseDateTimeBestEffortUSOrZero, partitionId, path, pathFull, pi, plus, pointInEllipses, pointInPolygon, polygonAreaCartesian, polygonAreaSpherical, polygonConvexHullCartesian, polygonPerimeterCartesian, polygonPerimeterSpherical, polygonsDistanceCartesian, polygonsDistanceSpherical, polygonsEqualsCartesian, polygonsIntersectionCartesian, polygonsIntersectionSpherical, polygonsSymDifferenceCartesian, polygonsSymDifferenceSpherical, polygonsUnionCartesian, polygonsUnionSpherical, polygonsWithinCartesian, polygonsWithinSpherical, port, position, positionCaseInsensitive, positionCaseInsensitiveUTF8, positionUTF8, pow, power, protocol, quantile, quantileBFloat16, quantileDeterministic, quantileExact, quantileExactExclusive, quantileExactHigh, quantileExactInclusive, quantileExactLow, quantileExactWeighted, quantileTDigest, quantileTDigestWeighted, quantileTiming, quantileTimingWeighted, quantiles, quantilesBFloat16, quantilesDeterministic, quantilesExact, quantilesExactExclusive, quantilesExactHigh, quantilesExactInclusive, quantilesExactLow, quantilesExactWeighted, quantilesTDigest, quantilesTDigestWeighted, quantilesTiming, quantilesTimingWeighted, queryString, queryStringAndFragment, rand, rand32, rand64, randConstant, randomFixedString, randomPrintableASCII, randomString, randomStringUTF8, range, rank, rankCorr, readWktMultiPolygon, readWktPoint, readWktPolygon, readWktRing, regexpQuoteMeta, regionHierarchy, regionIn, regionToArea, regionToCity, regionToContinent, regionToCountry, regionToDistrict, regionToName, regionToPopulation, regionToTopContinent, reinterpret, reinterpretAsDate, reinterpretAsDateTime, reinterpretAsFixedString, reinterpretAsFloat32, reinterpretAsFloat64, reinterpretAsInt128, reinterpretAsInt16, reinterpretAsInt256, reinterpretAsInt32, reinterpretAsInt64, reinterpretAsInt8, reinterpretAsString, reinterpretAsUInt128, reinterpretAsUInt16, reinterpretAsUInt256, reinterpretAsUInt32, reinterpretAsUInt64, reinterpretAsUInt8, reinterpretAsUUID, repeat, replace, replaceAll, replaceOne, replaceRegexpAll, replaceRegexpOne, replicate, retention, reverse, reverseUTF8, round, roundAge, roundBankers, roundDown, roundDuration, roundToExp2, rowNumberInAllBlocks, rowNumberInBlock, row_number, runningAccumulate, runningConcurrency, runningDifference, runningDifferenceStartingWithFirstValue, sequenceCount, sequenceMatch, sequenceNextNode, sigmoid, sign, simpleJSONExtractBool, simpleJSONExtractFloat, simpleJSONExtractInt, simpleJSONExtractRaw, simpleJSONExtractString, simpleJSONExtractUInt, simpleJSONHas, simpleLinearRegression, sin, sinh, sipHash128, sipHash64, skewPop, skewSamp, sleep, sleepEachRow, splitByChar, splitByRegexp, splitByString, sqrt, startsWith, stddevPop, stddevPopStable, stddevSamp, stddevSampStable, stochasticLinearRegression, stochasticLogisticRegression, stringToH3, studentTTest, substr, substring, substringUTF8, subtractDays, subtractHours, subtractMinutes, subtractMonths, subtractQuarters, subtractSeconds, subtractWeeks, subtractYears, sum, sumCount, sumKahan, sumMap, sumMapFiltered, sumMapFilteredWithOverflow, sumMapWithOverflow, sumWithOverflow, svg, tan, tanh, tcpPort, tgamma, throwIf, tid, timeSlot, timeSlots, timeZone, timeZoneOf, timeZoneOffset, timezone, timezoneOf, timezoneOffset, toColumnTypeName, toDate, toDateOrNull, toDateOrZero, toDateTime, toDateTime32, toDateTime64, toDateTime64OrNull, toDateTime64OrZero, toDateTimeOrNull, toDateTimeOrZero, toDayOfMonth, toDayOfWeek, toDayOfYear, toDecimal128, toDecimal128OrNull, toDecimal128OrZero, toDecimal256, toDecimal256OrNull, toDecimal256OrZero, toDecimal32, toDecimal32OrNull, toDecimal32OrZero, toDecimal64, toDecimal64OrNull, toDecimal64OrZero, toFixedString, toFloat32, toFloat32OrNull, toFloat32OrZero, toFloat64, toFloat64OrNull, toFloat64OrZero, toHour, toIPv4, toIPv6, toISOWeek, toISOYear, toInt128, toInt128OrNull, toInt128OrZero, toInt16, toInt16OrNull, toInt16OrZero, toInt256, toInt256OrNull, toInt256OrZero, toInt32, toInt32OrNull, toInt32OrZero, toInt64, toInt64OrNull, toInt64OrZero, toInt8, toInt8OrNull, toInt8OrZero, toIntervalDay, toIntervalHour, toIntervalMinute, toIntervalMonth, toIntervalQuarter, toIntervalSecond, toIntervalWeek, toIntervalYear, toJSONString, toLowCardinality, toMinute, toModifiedJulianDay, toModifiedJulianDayOrNull, toMonday, toMonth, toNullable, toQuarter, toRelativeDayNum, toRelativeHourNum, toRelativeMinuteNum, toRelativeMonthNum, toRelativeQuarterNum, toRelativeSecondNum, toRelativeWeekNum, toRelativeYearNum, toSecond, toStartOfDay, toStartOfFifteenMinutes, toStartOfFiveMinute, toStartOfHour, toStartOfISOYear, toStartOfInterval, toStartOfMinute, toStartOfMonth, toStartOfQuarter, toStartOfSecond, toStartOfTenMinutes, toStartOfWeek, toStartOfYear, toString, toStringCutToZero, toTime, toTimeZone, toTimezone, toTypeName, toUInt128, toUInt128OrNull, toUInt128OrZero, toUInt16, toUInt16OrNull, toUInt16OrZero, toUInt256, toUInt256OrNull, toUInt256OrZero, toUInt32, toUInt32OrNull, toUInt32OrZero, toUInt64, toUInt64OrNull, toUInt64OrZero, toUInt8, toUInt8OrNull, toUInt8OrZero, toUUID, toUUIDOrNull, toUUIDOrZero, toUnixTimestamp, toUnixTimestamp64Micro, toUnixTimestamp64Milli, toUnixTimestamp64Nano, toValidUTF8, toWeek, toYYYYMM, toYYYYMMDD, toYYYYMMDDhhmmss, toYear, toYearWeek, today, topK, topKWeighted, topLevelDomain, transform, trimBoth, trimLeft, trimRight, trunc, truncate, tryBase64Decode, tuple, tupleElement, tupleHammingDistance, ucase, unhex, uniq, uniqCombined, uniqCombined64, uniqExact, uniqHLL12, uniqTheta, uniqUpTo, upper, upperUTF8, uptime, user, validateNestedArraySizes, varPop, varPopStable, varSamp, varSampStable, version, visibleWidth, visitParamExtractBool, visitParamExtractFloat, visitParamExtractInt, visitParamExtractRaw, visitParamExtractString, visitParamExtractUInt, visitParamHas, week, welchTTest, windowFunnel, wkt, wordShingleMinHash, wordShingleMinHashArg, wordShingleMinHashArgCaseInsensitive, wordShingleMinHashArgCaseInsensitiveUTF8, wordShingleMinHashArgUTF8, wordShingleMinHashCaseInsensitive, wordShingleMinHashCaseInsensitiveUTF8, wordShingleMinHashUTF8, wordShingleSimHash, wordShingleSimHashCaseInsensitive, wordShingleSimHashCaseInsensitiveUTF8, wordShingleSimHashUTF8, xor, xxHash32, xxHash64, yandexConsistentHash, yearweek, yesterday"
	r := clickhouse.NewMetadataReader(db.DB).(metadata.FunctionReader)

	result, err := r.Functions(metadata.Filter{Schema: schema})
	if err != nil {
		log.Fatalf("Could not read %s functions: %v", dbName, err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Name)
	}
	actual := strings.Join(names, ", ")
	if actual != expected {
		t.Errorf("Wrong %s function names, expected:\n  %v\ngot:\n  %v", dbName, expected, names)
	}
}

func TestColumns(t *testing.T) {
	schema := "tutorial"
	table := "hits_v1"
	expected := "AdvEngineID, Age, BrowserCountry, BrowserLanguage, CLID, ClientEventTime, ClientIP, ClientIP6, ClientTimeZone, CodeVersion, ConnectTiming, CookieEnable, CounterClass, CounterID, DNSTiming, DOMCompleteTiming, DOMContentLoadedTiming, DOMInteractiveTiming, DontCountHits, EventDate, EventTime, FUniqID, FetchTiming, FirstPaintTiming, FlashMajor, FlashMinor, FlashMinor2, FromTag, GeneralInterests, GoalsReached, GoodEvent, HID, HTTPError, HasGCLID, HistoryLength, HitColor, IPNetworkID, Income, Interests, IsArtifical, IsDownload, IsEvent, IsLink, IsMobile, IsNotBounce, IsOldCounter, IsParameter, IsRobot, IslandID, JavaEnable, JavascriptEnable, LoadEventEndTiming, LoadEventStartTiming, MobilePhone, MobilePhoneModel, NSToDOMContentLoadedTiming, NetMajor, NetMinor, OS, OpenerName, OpenstatAdID, OpenstatCampaignID, OpenstatServiceName, OpenstatSourceID, PageCharset, ParamCurrency, ParamCurrencyID, ParamOrderID, ParamPrice, Params, ParsedParams.Key1, ParsedParams.Key2, ParsedParams.Key3, ParsedParams.Key4, ParsedParams.Key5, ParsedParams.ValueDouble, RedirectCount, RedirectTiming, Referer, RefererCategories, RefererDomain, RefererHash, RefererRegions, Refresh, RegionID, RemoteIP, RemoteIP6, RequestNum, RequestTry, ResolutionDepth, ResolutionHeight, ResolutionWidth, ResponseEndTiming, ResponseStartTiming, Robotness, SearchEngineID, SearchPhrase, SendTiming, Sex, ShareService, ShareTitle, ShareURL, SilverlightVersion1, SilverlightVersion2, SilverlightVersion3, SilverlightVersion4, SocialAction, SocialNetwork, SocialSourceNetworkID, SocialSourcePage, Title, TraficSourceID, URL, URLCategories, URLDomain, URLHash, URLRegions, UTCEventTime, UTMCampaign, UTMContent, UTMMedium, UTMSource, UTMTerm, UserAgent, UserAgentMajor, UserAgentMinor, UserID, WatchID, WindowClientHeight, WindowClientWidth, WindowName, WithHash, YCLID"
	r := db.Reader

	result, err := r.Columns(metadata.Filter{Schema: schema, Parent: table})
	if err != nil {
		log.Fatalf("Could not read %s columns: %v", dbName, err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Name)
	}
	actual := strings.Join(names, ", ")
	if actual != expected {
		t.Errorf("Wrong %s column names, expected:\n  %v, got:\n  %v", dbName, expected, names)
	}
}
