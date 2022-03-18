package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"tw-bot/data"
)

type Database struct {
	Sqlite *sql.DB
}

func GetDataBase() *Database {
	d := new(Database)
	var db *sql.DB
	var err error
	if d.Sqlite == nil {
		str, _ := os.Getwd()
		db, err = sql.Open("sqlite3", str+"/twitter.sqlite")
		if err != nil {
			panic(err)
		}
		d.Sqlite = db
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS tweets (tid INTEGER PRIMARY KEY, author TEXT, content TEXT, tags TEXT, media_urls TEXT, urls TEXT, is_publish INTEGER)")
	if err != nil {
		panic(err)
	}

	return d
}

func (db *Database) SaveOne(t data.Tweets) (int64, error) {
	if db.IsExists(t.ID) {
		return 0, nil
	}

	result, err := db.Sqlite.Exec("INSERT INTO tweets (tid, author, content, tags, media_urls, urls, is_publish) VALUES (?, ?, ?, ?, ?, ?, ?)", t.ID, t.Author, t.Content, t.Tags, t.MediaUrls, t.Url, 0)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
func (db *Database) IsExists(tid int64) bool {
	var count int
	err := db.Sqlite.QueryRow("SELECT count(*) FROM tweets WHERE tid = ?", tid).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

func (db *Database) Search(tid int64) ([]data.Tweets, error) {
	rows, err := db.Sqlite.Query("SELECT * FROM tweets WHERE tid = ?", tid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tweets []data.Tweets
	for rows.Next() {
		var tweet data.Tweets
		err := rows.Scan(&tweet.ID, &tweet.Author, &tweet.Content)
		if err != nil {
			return nil, err
		}
		tweets = append(tweets, tweet)
	}
	return tweets, nil
}

// QueryOne query one row from table
func (db *Database) QueryOne(tid int64) (data.Tweets, error) {
	var tweet data.Tweets
	row := db.Sqlite.QueryRow("SELECT * FROM tweets WHERE tid = ? and is_publish = ?", tid, 0)
	err := row.Scan(&tweet.ID, &tweet.Author, &tweet.Content, &tweet.Tags, &tweet.MediaUrls, &tweet.Url, &tweet.IsPublish)

	if err != nil {
		return tweet, err
	}

	return tweet, nil
}

func (db *Database) Delete(tid int64) error {
	_, err := db.Sqlite.Exec("DELETE FROM tweets WHERE tid = ?", tid)
	if err != nil {
		return err
	}
	return nil
}
