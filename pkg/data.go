package pkg

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"tw-bot/entity"
)

var (
	db  *sql.DB
	err error
)

func init() {

	str, _ := os.Getwd()
	db, err = sql.Open("sqlite3", str+"/twitter.sqlite")

	if err != nil {
		panic(err)
	}
	//defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS tweets (tid INTEGER PRIMARY KEY, author TEXT, content TEXT, tags TEXT, media_urls TEXT, urls TEXT)")
	if err != nil {
		return
	}
}

func GetDB() *sql.DB {

	if db == nil {
		str, _ := os.Getwd()
		db, err = sql.Open("sqlite3", str+"/twitter.sqlite")

		if err != nil {
			panic(err)
		}
	}
	return db
}

func SaveToDB(tid int64, author, content, tags, mediaUrls, urls string) (int64, error) {
	db := GetDB()

	if IsExists(tid) {
		return 0, nil
	}

	result, err := db.Exec("INSERT INTO tweets (tid, author, content, tags, media_urls, urls) VALUES (?, ?, ?, ?, ?, ?)", tid, author, content, tags, mediaUrls, urls)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
func IsExists(tid int64) bool {
	db := GetDB()
	var count int
	err := db.QueryRow("SELECT count(*) FROM tweets WHERE tid = ?", tid).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

func Search(tid int64) ([]entity.BotTweet, error) {
	db := GetDB()
	rows, err := db.Query("SELECT * FROM tweets WHERE tid = ?", tid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tweets []entity.BotTweet
	for rows.Next() {
		var tweet entity.BotTweet
		err := rows.Scan(&tweet.ID, &tweet.Author, &tweet.Content)
		if err != nil {
			return nil, err
		}
		tweets = append(tweets, tweet)
	}
	return tweets, nil
}

func Delete(tid int64) error {
	db := GetDB()
	_, err := db.Exec("DELETE FROM tweets WHERE tid = ?", tid)
	if err != nil {
		return err
	}
	return nil
}
