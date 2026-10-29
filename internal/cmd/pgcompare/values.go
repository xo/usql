package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
)

// gen makes one SQL expression of a type.
type gen func(r *rand.Rand) string

// typeCase is a type and how to make its values.
type typeCase struct {
	name  string // the type as written in a cast, such as "int4" or "char(5)"
	edges []string
	gen   gen
	array bool // also test the array of the type
}

func lit(s, typ string) string {
	return "E'" + escape(s) + "'::" + typ
}

// escape writes s as the body of an E” string, with every rune that is not
// printable ASCII as \U escape.
func escape(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch {
		case c == '\'':
			b.WriteString(`\'`)
		case c == '\\':
			b.WriteString(`\\`)
		case c >= 0x20 && c < 0x7f:
			b.WriteRune(c)
		default:
			fmt.Fprintf(&b, `\U%08x`, c)
		}
	}
	return b.String()
}

// pools of runes that a random text draws from
var runePools = [][]rune{
	[]rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
	[]rune(" \t\n\r"),
	[]rune(`\'"|,;{}[]()<>%$#@!&*+-=_~^` + "`"),
	[]rune("日本語中文한국어"),
	[]rune("😀🎉👍🏽👨‍👩‍👧"),
	[]rune("é̃ñüßøåÆ"),
	{0x200b, 0x200d, 0xfeff, 0x00a0, 0x2028, 0x2029, 0x202e, 0x0301, 0x0336},
	{0x01, 0x02, 0x07, 0x08, 0x0b, 0x0c, 0x1b, 0x1f, 0x7f, 0x85, 0x9b},
	[]rune("ΑΒΓαβγабвгд"),
	[]rune("עברית العربية"),
}

func randText(r *rand.Rand, max int) string {
	n := r.IntN(max + 1)
	var b strings.Builder
	pool := runePools[r.IntN(len(runePools))]
	for range n {
		if r.IntN(4) == 0 {
			pool = runePools[r.IntN(len(runePools))]
		}
		b.WriteRune(pool[r.IntN(len(pool))])
	}
	return b.String()
}

func randASCII(r *rand.Rand, max int) string {
	n := r.IntN(max + 1)
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(0x20 + r.IntN(0x5f))
	}
	return string(b)
}

func randInt(r *rand.Rand, bits int) int64 {
	switch r.IntN(6) {
	case 0:
		return 0
	case 1:
		return -1 << (bits - 1)
	case 2:
		return 1<<(bits-1) - 1
	case 3:
		return int64(r.IntN(200)) - 100
	}
	v := int64(r.Uint64())
	shift := 64 - bits + r.IntN(bits)
	return v >> shift
}

func randFloat(r *rand.Rand, bits int) string {
	switch r.IntN(10) {
	case 0:
		return []string{"NaN", "Infinity", "-Infinity", "-0", "0"}[r.IntN(5)]
	case 1:
		if bits == 32 {
			return strconv.FormatFloat(float64(math.Float32frombits(r.Uint32()&0x007fffff)), 'g', -1, 32)
		}
		return strconv.FormatFloat(math.Float64frombits(r.Uint64()&0x000fffffffffffff), 'g', -1, 64)
	case 2:
		return strconv.FormatFloat(float64(r.IntN(2000000)-1000000)/100, 'g', -1, bits)
	}
	for {
		if bits == 32 {
			f := math.Float32frombits(r.Uint32())
			if !math.IsNaN(float64(f)) && !math.IsInf(float64(f), 0) {
				return strconv.FormatFloat(float64(f), 'g', -1, 32)
			}
			continue
		}
		f := math.Float64frombits(r.Uint64())
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return strconv.FormatFloat(f, 'g', -1, 64)
		}
	}
}

func randDigits(r *rand.Rand, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('0' + r.IntN(10))
	}
	return string(b)
}

func randNumeric(r *rand.Rand) string {
	switch r.IntN(8) {
	case 0:
		return []string{"NaN", "Infinity", "-Infinity", "0", "-0.000", "1e-30", "1e30"}[r.IntN(7)]
	}
	s := randDigits(r, 1+r.IntN(30))
	if r.IntN(2) == 0 {
		s += "." + randDigits(r, r.IntN(30))
	}
	if r.IntN(2) == 0 {
		s = "-" + s
	}
	return s
}

func pick(r *rand.Rand, s ...string) string { return s[r.IntN(len(s))] }

func randTimestamp(r *rand.Rand, zone bool) string {
	var y int
	switch r.IntN(5) {
	case 0:
		y = 1 + r.IntN(4713)
	case 1:
		y = 1 + r.IntN(294275)
	default:
		y = 1900 + r.IntN(200)
	}
	s := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", y, 1+r.IntN(12), 1+r.IntN(28), r.IntN(24), r.IntN(60), r.IntN(60))
	switch r.IntN(3) {
	case 0:
		s += fmt.Sprintf(".%06d", r.IntN(1000000))
	case 1:
		s += fmt.Sprintf(".%d", r.IntN(1000))
	}
	if zone {
		s += pick(r, "+00", "+05:30", "-08", "+14", "-12:45", " UTC", " Asia/Jakarta", " America/New_York", "+00:00:01")
	}
	if r.IntN(6) == 0 {
		s += " BC"
	}
	return s
}

func randTime(r *rand.Rand, zone bool) string {
	s := fmt.Sprintf("%02d:%02d:%02d", r.IntN(24), r.IntN(60), r.IntN(60))
	if r.IntN(2) == 0 {
		s += fmt.Sprintf(".%06d", r.IntN(1000000))
	}
	if r.IntN(20) == 0 {
		s = "24:00:00"
	}
	if zone {
		s += pick(r, "+00", "+05:30", "-08", "+15:59", "-15:59", "+01:02:03")
	}
	return s
}

func randFloatSimple(r *rand.Rand) string {
	return strconv.FormatFloat(float64(r.IntN(2000001)-1000000)/float64(1+r.IntN(1000)), 'g', -1, 64)
}

func randJSON(r *rand.Rand, depth int) string {
	k := r.IntN(8)
	if depth > 3 {
		k = r.IntN(5)
	}
	switch k {
	case 0:
		return "null"
	case 1:
		return pick(r, "true", "false")
	case 2:
		return pick(r, "0", "-1", "1.5e300", "123456789012345678901234567890", "-0.0000001", "3.14")
	case 3, 4:
		return strconvQuote(randText(r, 12))
	case 5:
		n := r.IntN(4)
		parts := make([]string, n)
		for i := range parts {
			parts[i] = randJSON(r, depth+1)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		n := r.IntN(4)
		parts := make([]string, n)
		for i := range parts {
			parts[i] = strconvQuote(randText(r, 6)) + ": " + randJSON(r, depth+1)
		}
		return "{" + strings.Join(parts, ", ") + "}"
	}
}

// strconvQuote writes a JSON string.
func strconvQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, c := range s {
		switch {
		case c == '"':
			b.WriteString(`\"`)
		case c == '\\':
			b.WriteString(`\\`)
		case c < 0x20:
			fmt.Fprintf(&b, `\u%04x`, c)
		default:
			b.WriteRune(c)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func randIP(r *rand.Rand, v6 bool) (string, int) {
	if !v6 {
		return fmt.Sprintf("%d.%d.%d.%d", r.IntN(256), r.IntN(256), r.IntN(256), r.IntN(256)), 32
	}
	parts := make([]string, 8)
	for i := range parts {
		if r.IntN(3) == 0 {
			parts[i] = "0"
		} else {
			parts[i] = strconv.FormatUint(uint64(r.IntN(65536)), 16)
		}
	}
	return strings.Join(parts, ":"), 128
}

func randHex(r *rand.Rand, n int, sep string, group int) string {
	var b strings.Builder
	for i := range n {
		if i > 0 && group > 0 && i%group == 0 {
			b.WriteString(sep)
		}
		fmt.Fprintf(&b, "%x", r.IntN(16))
	}
	return b.String()
}

func randBits(r *rand.Rand, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('0' + r.IntN(2))
	}
	return string(b)
}

func randPoint(r *rand.Rand) string {
	return "(" + randFloatSimple(r) + "," + randFloatSimple(r) + ")"
}

func randPoints(r *rand.Rand, min int) string {
	n := min + r.IntN(4)
	parts := make([]string, n)
	for i := range parts {
		parts[i] = randPoint(r)
	}
	return strings.Join(parts, ",")
}

func randInterval(r *rand.Rand) string {
	switch r.IntN(8) {
	case 0:
		return pick(r, "'infinity'::interval", "'-infinity'::interval", "'0'::interval", "'178000000 years'::interval", "'-178000000 years'::interval")
	}
	return fmt.Sprintf("make_interval(%d, %d, %d, %d, %d, %d, %s)",
		r.IntN(200)-100, r.IntN(30)-15, r.IntN(10)-5, r.IntN(100)-50, r.IntN(100)-50, r.IntN(200)-100,
		strconv.FormatFloat(float64(r.IntN(200000000)-100000000)/1e6, 'f', 6, 64))
}

func randRange(r *rand.Rand, fn string, bound func() string) string {
	if r.IntN(8) == 0 {
		return "'empty'::" + fn
	}
	a, b := bound(), bound()
	bounds := pick(r, "[)", "(]", "[]", "()")
	lo, hi := "least("+a+","+b+")", "greatest("+a+","+b+")"
	switch r.IntN(6) {
	case 0:
		lo = "NULL"
	case 1:
		hi = "NULL"
	}
	return fmt.Sprintf("%s(%s, %s, '%s')", fn, lo, hi, bounds)
}

func ts(r *rand.Rand) string {
	return "'" + fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", 1900+r.IntN(200), 1+r.IntN(12), 1+r.IntN(28), r.IntN(24), r.IntN(60), r.IntN(60)) + "'::timestamp"
}
func tstz(r *rand.Rand) string {
	return "'" + fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d+0%d", 1900+r.IntN(200), 1+r.IntN(12), 1+r.IntN(28), r.IntN(24), r.IntN(60), r.IntN(60), r.IntN(9)) + "'::timestamptz"
}
func date(r *rand.Rand) string {
	return fmt.Sprintf("('2000-01-01'::date + %d)", r.IntN(200000)-100000)
}

func catalog(table, typ string, k int) string {
	return fmt.Sprintf("(select oid from %s order by oid offset %d limit 1)::%s", table, k, typ)
}

var cases = []typeCase{
	{name: "bool", array: true, edges: []string{"true", "false"}, gen: func(r *rand.Rand) string { return pick(r, "true", "false", "'yes'::bool", "'0'::bool") }},
	{name: "int2", array: true, gen: func(r *rand.Rand) string { return strconv.FormatInt(randInt(r, 16), 10) + "::int2" }},
	{name: "int4", array: true, gen: func(r *rand.Rand) string { return "(" + strconv.FormatInt(randInt(r, 32), 10) + ")::int4" }},
	{name: "int8", array: true, gen: func(r *rand.Rand) string { return "'" + strconv.FormatInt(randInt(r, 64), 10) + "'::int8" }},
	{name: "float4", array: true, gen: func(r *rand.Rand) string { return "'" + randFloat(r, 32) + "'::float4" }},
	{name: "float8", array: true, gen: func(r *rand.Rand) string { return "'" + randFloat(r, 64) + "'::float8" }},
	{name: "numeric", array: true, gen: func(r *rand.Rand) string { return "'" + randNumeric(r) + "'::numeric" }},
	{name: "numeric(10,3)", gen: func(r *rand.Rand) string {
		return fmt.Sprintf("'%d.%s'::numeric(10,3)", r.IntN(2000000)-1000000, randDigits(r, r.IntN(6)))
	}},
	{name: "money", array: true, gen: func(r *rand.Rand) string {
		return fmt.Sprintf("'%d.%02d'::money", r.Int64N(184467440737095)-92233720368547, r.IntN(100))
	}},
	{name: "text", array: true, edges: []string{"''::text", lit("NULL", "text"), lit(" lead and trail ", "text"), lit("a\nb", "text"), lit("tab\there", "text"), lit(`back\slash`, "text"), lit("{}", "text"), lit(`"`, "text")}, gen: func(r *rand.Rand) string { return lit(randText(r, 30), "text") }},
	{name: "varchar", array: true, gen: func(r *rand.Rand) string { return lit(randText(r, 20), "varchar") }},
	{name: "varchar(5)", gen: func(r *rand.Rand) string { return lit(randText(r, 8), "varchar(5)") }},
	{name: "bpchar", array: true, gen: func(r *rand.Rand) string { return lit(randText(r, 10), "bpchar") }},
	{name: "char(8)", gen: func(r *rand.Rand) string { return lit(randText(r, 8), "char(8)") }},
	{name: `"char"`, array: true, gen: func(r *rand.Rand) string { return lit(randASCII(r, 1), `"char"`) }},
	{name: "name", array: true, gen: func(r *rand.Rand) string { return lit(randText(r, 40), "name") }},
	{name: "bytea", array: true, edges: []string{`'\x'::bytea`, `'\x00'::bytea`, `'\xdeadbeef'::bytea`, `'hello'::bytea`}, gen: func(r *rand.Rand) string {
		b := make([]byte, r.IntN(24))
		for i := range b {
			b[i] = byte(r.IntN(256))
		}
		return fmt.Sprintf(`'\x%x'::bytea`, b)
	}},
	{name: "bit(8)", array: true, gen: func(r *rand.Rand) string { return "B'" + randBits(r, 8) + "'::bit(8)" }},
	{name: "varbit", array: true, gen: func(r *rand.Rand) string { return "B'" + randBits(r, r.IntN(40)) + "'::varbit" }},
	{name: "uuid", array: true, gen: func(r *rand.Rand) string {
		h := randHex(r, 32, "", 0)
		return "'" + h[:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:] + "'::uuid"
	}},
	{name: "date", array: true, edges: []string{"'infinity'::date", "'-infinity'::date", "'4713-01-01 BC'::date", "'5874897-12-31'::date"}, gen: date},
	{name: "timestamp", array: true, edges: []string{"'infinity'::timestamp", "'-infinity'::timestamp", "'294276-12-31 23:59:59.999999'::timestamp", "'4713-01-01 00:00:00 BC'::timestamp"}, gen: func(r *rand.Rand) string { return "'" + randTimestamp(r, false) + "'::timestamp" }},
	{name: "timestamp(0)", gen: func(r *rand.Rand) string { return "'" + randTimestamp(r, false) + "'::timestamp(0)" }},
	{name: "timestamptz", array: true, edges: []string{"'infinity'::timestamptz", "'-infinity'::timestamptz"}, gen: func(r *rand.Rand) string { return "'" + randTimestamp(r, true) + "'::timestamptz" }},
	{name: "time", array: true, gen: func(r *rand.Rand) string { return "'" + randTime(r, false) + "'::time" }},
	{name: "timetz", array: true, gen: func(r *rand.Rand) string { return "'" + randTime(r, true) + "'::timetz" }},
	{name: "interval", array: true, gen: randInterval},
	{name: "inet", array: true, gen: func(r *rand.Rand) string {
		ip, bits := randIP(r, r.IntN(2) == 0)
		if r.IntN(2) == 0 {
			return "'" + ip + "'::inet"
		}
		return fmt.Sprintf("'%s/%d'::inet", ip, r.IntN(bits+1))
	}},
	{name: "cidr", array: true, gen: func(r *rand.Rand) string {
		ip, bits := randIP(r, r.IntN(2) == 0)
		return fmt.Sprintf("network('%s/%d'::inet)", ip, r.IntN(bits+1))
	}},
	{name: "macaddr", array: true, gen: func(r *rand.Rand) string { return "'" + randHex(r, 12, ":", 2) + "'::macaddr" }},
	{name: "macaddr8", array: true, gen: func(r *rand.Rand) string { return "'" + randHex(r, 16, ":", 2) + "'::macaddr8" }},
	{name: "point", array: true, gen: func(r *rand.Rand) string { return "'" + randPoint(r) + "'::point" }},
	{name: "line", array: true, gen: func(r *rand.Rand) string {
		return fmt.Sprintf("'{%s,%d,%s}'::line", randFloatSimple(r), 1+r.IntN(100), randFloatSimple(r))
	}},
	{name: "lseg", array: true, gen: func(r *rand.Rand) string { return "'[" + randPoint(r) + "," + randPoint(r) + "]'::lseg" }},
	{name: "box", array: true, gen: func(r *rand.Rand) string { return "'" + randPoint(r) + "," + randPoint(r) + "'::box" }},
	{name: "path", array: true, gen: func(r *rand.Rand) string {
		if r.IntN(2) == 0 {
			return "'[" + randPoints(r, 1) + "]'::path"
		}
		return "'(" + randPoints(r, 1) + ")'::path"
	}},
	{name: "polygon", array: true, gen: func(r *rand.Rand) string { return "'(" + randPoints(r, 1) + ")'::polygon" }},
	{name: "circle", array: true, gen: func(r *rand.Rand) string {
		return fmt.Sprintf("'<%s,%d>'::circle", randPoint(r), r.IntN(1000))
	}},
	{name: "json", array: true, gen: func(r *rand.Rand) string { return lit(randJSON(r, 0), "json") }},
	{name: "jsonb", array: true, gen: func(r *rand.Rand) string { return lit(randJSON(r, 0), "jsonb") }},
	{name: "jsonpath", array: true, gen: func(r *rand.Rand) string {
		return lit(pick(r, "$", "$.a", "$.a[*] ? (@ > 1)", `$."key with space"`, "strict $.a.b[1 to 3]", `$ ? (@ like_regex "^a.*" flag "i")`, "$.datetime()"), "jsonpath")
	}},
	{name: "xml", array: true, gen: func(r *rand.Rand) string {
		return "xmlelement(name a, xmlattributes(" + lit(randText(r, 6), "text") + " as b), " + lit(randText(r, 12), "text") + ")"
	}},
	{name: "tsvector", array: true, gen: func(r *rand.Rand) string {
		if r.IntN(2) == 0 {
			return "to_tsvector('english', " + lit(randText(r, 30), "text") + ")"
		}
		return lit(pick(r, "a:1A fat:2B,4C cat:5D", "'with space' 'quote''s'", "a b c", ""), "tsvector")
	}},
	{name: "tsquery", array: true, gen: func(r *rand.Rand) string {
		return lit(pick(r, "a & b", "(a | b) & !c", "fat:ab & cat", "'super':*", "a <-> b", "a <2> b"), "tsquery")
	}},
	{name: "oid", array: true, gen: func(r *rand.Rand) string { return fmt.Sprintf("'%d'::oid", r.Uint32()) }},
	{name: "xid", array: true, gen: func(r *rand.Rand) string { return fmt.Sprintf("'%d'::xid", r.Uint32()) }},
	{name: "xid8", array: true, gen: func(r *rand.Rand) string { return fmt.Sprintf("'%d'::xid8", r.Uint64()>>1) }},
	{name: "cid", array: true, gen: func(r *rand.Rand) string { return fmt.Sprintf("'%d'::cid", r.Uint32()) }},
	{name: "tid", array: true, gen: func(r *rand.Rand) string { return fmt.Sprintf("'(%d,%d)'::tid", r.Uint32(), r.IntN(65536)) }},
	{name: "pg_lsn", array: true, gen: func(r *rand.Rand) string { return fmt.Sprintf("'%X/%X'::pg_lsn", r.Uint32(), r.Uint32()) }},
	{name: "pg_snapshot", array: true, gen: func(r *rand.Rand) string {
		a := 10 + r.IntN(100)
		return fmt.Sprintf("'%d:%d:%d,%d'::pg_snapshot", a, a+10, a+1, a+5)
	}},
	{name: "txid_snapshot", array: true, gen: func(r *rand.Rand) string {
		a := 10 + r.IntN(100)
		return fmt.Sprintf("'%d:%d:'::txid_snapshot", a, a+10)
	}},
	{name: "aclitem", array: true, gen: func(r *rand.Rand) string {
		return fmt.Sprintf("'%s=%s/postgres'::aclitem", pick(r, "postgres", "", "pg_monitor"), pick(r, "r", "arwdDxt", "r*w*", "U", "C*"))
	}},
	{name: "refcursor", array: true, gen: func(r *rand.Rand) string { return lit(randText(r, 10), "refcursor") }},
	{name: "int2vector", gen: func(r *rand.Rand) string {
		return fmt.Sprintf("'%d %d %d'::int2vector", r.IntN(100), r.IntN(100), r.IntN(100))
	}},
	{name: "oidvector", gen: func(r *rand.Rand) string { return fmt.Sprintf("'%d %d'::oidvector", r.Uint32(), r.Uint32()) }},
	{name: "regclass", array: true, gen: func(r *rand.Rand) string { return catalog("pg_class", "regclass", r.IntN(300)) }},
	{name: "regtype", array: true, gen: func(r *rand.Rand) string { return catalog("pg_type", "regtype", r.IntN(300)) }},
	{name: "regproc", array: true, gen: func(r *rand.Rand) string { return catalog("pg_proc", "regproc", r.IntN(2000)) }},
	{name: "regprocedure", array: true, gen: func(r *rand.Rand) string { return catalog("pg_proc", "regprocedure", r.IntN(2000)) }},
	{name: "regoper", array: true, gen: func(r *rand.Rand) string { return catalog("pg_operator", "regoper", r.IntN(500)) }},
	{name: "regoperator", array: true, gen: func(r *rand.Rand) string { return catalog("pg_operator", "regoperator", r.IntN(500)) }},
	{name: "regnamespace", array: true, gen: func(r *rand.Rand) string { return catalog("pg_namespace", "regnamespace", r.IntN(3)) }},
	{name: "regrole", array: true, gen: func(r *rand.Rand) string { return catalog("pg_authid", "regrole", r.IntN(5)) }},
	{name: "regconfig", array: true, gen: func(r *rand.Rand) string { return catalog("pg_ts_config", "regconfig", r.IntN(20)) }},
	{name: "regdictionary", array: true, gen: func(r *rand.Rand) string { return catalog("pg_ts_dict", "regdictionary", r.IntN(20)) }},
	{name: "regcollation", array: true, gen: func(r *rand.Rand) string { return catalog("pg_collation", "regcollation", r.IntN(3)) }},
	{name: "pg_node_tree", gen: func(r *rand.Rand) string {
		return fmt.Sprintf("(select ev_action from pg_rewrite order by oid offset %d limit 1)", r.IntN(50))
	}},
	{name: "int4range", array: true, gen: func(r *rand.Rand) string {
		return randRange(r, "int4range", func() string { return strconv.FormatInt(randInt(r, 31), 10) })
	}},
	{name: "int8range", array: true, gen: func(r *rand.Rand) string {
		return randRange(r, "int8range", func() string { return "'" + strconv.FormatInt(randInt(r, 63), 10) + "'::int8" })
	}},
	{name: "numrange", array: true, gen: func(r *rand.Rand) string {
		return randRange(r, "numrange", func() string { return "'" + randDigits(r, 1+r.IntN(8)) + "." + randDigits(r, r.IntN(5)) + "'::numeric" })
	}},
	{name: "tsrange", array: true, gen: func(r *rand.Rand) string { return randRange(r, "tsrange", func() string { return ts(r) }) }},
	{name: "tstzrange", array: true, gen: func(r *rand.Rand) string { return randRange(r, "tstzrange", func() string { return tstz(r) }) }},
	{name: "daterange", array: true, gen: func(r *rand.Rand) string { return randRange(r, "daterange", func() string { return date(r) }) }},
	{name: "int4multirange", array: true, gen: func(r *rand.Rand) string {
		b := func() string { return strconv.Itoa(r.IntN(2000) - 1000) }
		return fmt.Sprintf("int4multirange(%s, %s)", randRange(r, "int4range", b), randRange(r, "int4range", b))
	}},
	{name: "nummultirange", array: true, gen: func(r *rand.Rand) string {
		b := func() string { return "'" + randDigits(r, 3) + "." + randDigits(r, 2) + "'::numeric" }
		return fmt.Sprintf("nummultirange(%s)", randRange(r, "numrange", b))
	}},
	{name: "datemultirange", array: true, gen: func(r *rand.Rand) string {
		return fmt.Sprintf("datemultirange(%s, %s)", randRange(r, "daterange", func() string { return date(r) }), randRange(r, "daterange", func() string { return date(r) }))
	}},
	{name: "tsmultirange", array: true, gen: func(r *rand.Rand) string {
		return fmt.Sprintf("tsmultirange(%s)", randRange(r, "tsrange", func() string { return ts(r) }))
	}},
	{name: "tstzmultirange", array: true, gen: func(r *rand.Rand) string {
		return fmt.Sprintf("tstzmultirange(%s)", randRange(r, "tstzrange", func() string { return tstz(r) }))
	}},
	{name: "int8multirange", array: true, gen: func(r *rand.Rand) string {
		b := func() string { return "'" + strconv.FormatInt(randInt(r, 63), 10) + "'::int8" }
		return fmt.Sprintf("int8multirange(%s)", randRange(r, "int8range", b))
	}},
}
