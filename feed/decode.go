package feed

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
	"golang.org/x/text/transform"
)

// ErrInvalidUTF8 is the error for a line of input that is not valid UTF-8.
var ErrInvalidUTF8 = errors.New(`not valid UTF-8: name the encoding of the script with \uset encoding, such as \uset encoding windows-1252`)

// UTF8 is the name of the encoding that input has when nothing else is named.
const UTF8 = "utf-8"

// Encoding returns the name of an encoding as the WHATWG Encoding Standard
// names it, such as windows-1252 for latin1 or shift_jis for sjis. It returns
// an error for a name that the standard does not know.
func Encoding(name string) (string, error) {
	if name == "" {
		return UTF8, nil
	}
	enc, err := htmlindex.Get(name)
	if err != nil {
		return "", fmt.Errorf("unknown encoding: %s", name)
	}
	return htmlindex.Name(enc)
}

// Decode returns a reader of r as UTF-8.
//
// A byte order mark at the start of r decides the encoding: the mark of UTF-8
// is dropped, and UTF-16 and UTF-32 are decoded. Otherwise name, a name that
// Encoding takes, is the encoding of r. Text in UTF-8, the default, is read
// as it is, and Source reports a line that is not valid UTF-8.
func Decode(r io.Reader, name string) (io.Reader, error) {
	br := bufio.NewReader(r)
	// Peek returns what there is when r holds less
	b, _ := br.Peek(4)
	var enc encoding.Encoding
	switch {
	case bytes.HasPrefix(b, []byte{0xef, 0xbb, 0xbf}):
		_, _ = br.Discard(3)
		return br, nil
	// UTF-32 first, because its little endian mark starts with UTF-16's
	case bytes.HasPrefix(b, []byte{0xff, 0xfe, 0, 0}):
		enc = utf32.UTF32(utf32.LittleEndian, utf32.ExpectBOM)
	case bytes.HasPrefix(b, []byte{0, 0, 0xfe, 0xff}):
		enc = utf32.UTF32(utf32.BigEndian, utf32.ExpectBOM)
	case bytes.HasPrefix(b, []byte{0xff, 0xfe}):
		enc = unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM)
	case bytes.HasPrefix(b, []byte{0xfe, 0xff}):
		enc = unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM)
	case name == "", strings.EqualFold(name, UTF8):
		return br, nil
	default:
		var err error
		if enc, err = htmlindex.Get(name); err != nil {
			return nil, fmt.Errorf("unknown encoding: %s", name)
		}
		if n, _ := htmlindex.Name(enc); n == UTF8 {
			return br, nil
		}
	}
	return transform.NewReader(br, enc.NewDecoder()), nil
}
