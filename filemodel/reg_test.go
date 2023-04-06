package filemodel

import (
	"regexp"
	"testing"
)

func TestDSNPrefix(t *testing.T) {
	pat := regexp.MustCompile(`^([\d\w]+)://(.*)`)
	dsn := "mysql://dddsssff"
	matched := pat.FindStringSubmatch(dsn)
	if matched[1] != "mysql" {
		t.Fatal("regex failed")
	}
	dsn = "sqlite3://dddsssff"
	matched = pat.FindStringSubmatch(dsn)
	if matched[1] != "sqlite3" {
		t.Fatal("regex failed")
	}

}
