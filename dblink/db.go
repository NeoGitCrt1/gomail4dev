package dblink

import (
	"database/sql"
	"flag"

	_ "github.com/mattn/go-sqlite3"
)

var Db *sql.DB

var DbData string

func init() {
	flag.StringVar(&DbData, "data_path", "./data.db", "data file path")
}
func InitDb() {
	db, err := sql.Open("sqlite3", DbData)
	if err != nil {
		panic(err)
	}

	db.Exec("CREATE TABLE Message ( id TEXT NOT NULL, [from] TEXT, [to] TEXT, receivedDate TEXT NOT NULL, subject TEXT, data BLOB, mimeParseError TEXT, sessionId TEXT, attachmentCount INTEGER NOT NULL DEFAULT 0, isUnread INTEGER NOT NULL DEFAULT 1);")
	// migrate data from Rnoowd.Smtp4dev
	db.Exec("insert into Message ( id , [from] , [to] , receivedDate , subject , data , mimeParseError , sessionId , attachmentCount , isUnread ) SELECT Id, [From], [To], ReceivedDate, Subject, Data, MimeParseError, SessionId, AttachmentCount, IsUnread FROM Messages;")
	db.Exec(`ALTER TABLE Messages RENAME TO Messages_arch;`)
	db.SetMaxOpenConns(5)
	Db = db
}
