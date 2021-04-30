package dblink

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var dbholder *sql.DB

func init() {
	fmt.Println("open sqllite:")
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		panic(err)
	}

	db.Exec("CREATE TABLE Message ( id TEXT NOT NULL, [from] TEXT, [to] TEXT, receivedDate TEXT NOT NULL, subject TEXT, data BLOB, mimeParseError TEXT, sessionId TEXT, attachmentCount INTEGER NOT NULL DEFAULT 0, isUnread INTEGER NOT NULL DEFAULT 0);")
	db.Exec("insert into Message ( id , [from] , [to] , receivedDate , subject , data , mimeParseError , sessionId , attachmentCount , isUnread ) SELECT Id, [From], [To], ReceivedDate, Subject, Data, MimeParseError, SessionId, AttachmentCount, IsUnread FROM Messages;")
	db.Exec("drop table Messages;")
	
	db.SetMaxOpenConns(20)
	dbholder = db
}

func Db() *sql.DB {
	return dbholder
}