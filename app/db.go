package app

import (
	"database/sql"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ungerik/go-rss"
)

var db *sql.DB

func (s *Settings) GetDatabase() *sql.DB {
	if db != nil {
		return db
	}

	log.Print("mysql DSN ", s.Db.Dsn, "?", s.Db.Opt)
	if handle, err := sql.Open("sqlite3", s.Db.Dsn+"?"+s.Db.Opt); err != nil {
		s.Fatal(err)
	} else {
		db = handle
		s.initDb(db)
	}
	return db
}

func (s *Settings) CloseDatabase() {
	if db != nil {
		db.Close()
	}
}

func (s *Settings) initDb(db *sql.DB) {
	sqlStmt := "CREATE TABLE IF NOT EXISTS `feed` (" +
		"`ts` TEXT NOT NULL," +
		"`sent` TEXT NULL," +
		"`url` TEXT PRIMARY KEY," +
		"`title` TEXT NOT NULL," +
		"`category` TEXT NOT NULL," +
		"`teaser` TEXT NOT NULL" +
		")"
	if _, err := db.Exec(sqlStmt); err != nil {
		s.Fatal(err)
	}

	owa := time.Now().AddDate(0, 0, -7)
	sdel := "DELETE FROM `feed` WHERE `ts`<?"
	if _, err := db.Exec(sdel, owa.Format(time.RFC3339)); err != nil {
		s.Fatal(err)
	}
}

func (s *Settings) StoreItem(item *rss.Item) bool {
	db := s.GetDatabase()
	categories := strings.Join(item.Category, ";")
	sql := "INSERT OR IGNORE INTO `feed` (`ts`,`url`,`title`,`teaser`,`category`) VALUES (?,?,?,?,?)"
	if rslt, err := db.Exec(sql, item.PubDate.MustFormat(time.RFC3339), item.Link, item.Title, item.Description, categories); err != nil {
		s.Log(err)
	} else if ra, err := rslt.RowsAffected(); err != nil {
		s.Log(err)
	} else {
		return ra > 0
	}
	return false
}

func (s *Settings) GetUnsent() *rss.Item {
	db := s.GetDatabase()
	sql := "SELECT `ts`,`url`,`title`,`category`,`teaser` FROM `feed` WHERE `sent` IS NULL ORDER BY `ts` ASC LIMIT 1"
	row := db.QueryRow(sql)
	item := rss.Item{}
	categories := ""
	if err := row.Scan(&item.PubDate, &item.Link, &item.Title, &categories, &item.Description); err != nil {
		s.Log(err)
		return nil
	}
	item.Category = strings.Split(categories, ";")
	return &item
}

func (s *Settings) MarkSent(item *rss.Item) {
	db := s.GetDatabase()
	sql := "UPDATE `feed` SET `sent`=datetime() WHERE `url`=?"
	if _, err := db.Exec(sql, item.Link); err != nil {
		s.Log(err)
	}
}

func (s *Settings) MarkError(item *rss.Item, err error) {
	db := s.GetDatabase()
	sql := "UPDATE `feed` SET `sent`=? WHERE `url`=?"
	if _, err := db.Exec(sql, err, item.Link); err != nil {
		s.Log(err)
	}
}
