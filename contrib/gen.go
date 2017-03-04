// +build ignore

package main

// this is an unfinished piece of code that originally was going to be used to
// parse the postgres docs to extract the "query prefixes" for the various
// types, it is left here in case this approach is followed in the future.

import (
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

const (
	pgurl = `https://www.postgresql.org/docs/current/static/sql-commands.html`
)

var entryRE = regexp.MustCompile(`<dt><a[a-z \n\-=\."]*?>([A-Z \n]+)`)

var cleanRE = regexp.MustCompile(`[ \n]+`)

func main() {
	res, err := http.Get(pgurl)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	matches := entryRE.FindAllStringSubmatch(string(buf), -1)
	for _, m := range matches {
		s := cleanRE.ReplaceAllString(m[1], " ")
		log.Printf(">>> %s", s)
	}
}
