package dblink

import (
	"database/sql"
	"flag"
	"strconv"
	"time"

	"github.com/NeoGitCrt1/go-snowflake"
	_ "github.com/mattn/go-sqlite3"
)

var Db *sql.DB

var dbData string

func init() {
	flag.StringVar(&dbData, "data_path", "./data.db", "data file path")
}
func InitDb() {
	db, err := sql.Open("sqlite3", dbData)
	if err != nil {
		panic(err)
	}

	db.Exec("CREATE TABLE Message ( id TEXT NOT NULL, [from] TEXT, [to] TEXT, receivedDate TEXT NOT NULL, subject TEXT, data BLOB, mimeParseError TEXT, sessionId TEXT, attachmentCount INTEGER NOT NULL DEFAULT 0, isUnread INTEGER NOT NULL DEFAULT 1);")
	// migrate data from Rnoowd.Smtp4dev
	x, err := db.Query("SELECT [From], [To], ReceivedDate, Subject, Data, MimeParseError, SessionId, AttachmentCount, IsUnread FROM Messages order by ReceivedDate;")

	if err == nil {
		var from string
		var to string
		var recv string
		var subj string
		var data []byte
		var pErr string
		var seId string
		var aCnt int
		var unread int
		stmt, e := db.Prepare("INSERT INTO Message ( id, [from], [to], subject,receivedDate, data, isUnread, mimeParseError, attachmentCount ) values (?,?,?,?,?,?,?,?,?)")
		snowflake.SetStartTime(time.Date(2000, 1,1, 1,1,1,1,time.Now().Location()))
		for x.Next() && e == nil {
			x.Scan(&from, &to, &recv, &subj, &data, &pErr, &seId, &aCnt, &unread)
			stmt.Exec(strconv.FormatUint(snowflake.ID(), 10),
				from, to, subj,
				recv,
				data, unread, pErr,
				aCnt,
			)

		}
		stmt.Close()

		db.Exec(`ALTER TABLE Messages RENAME TO Messages_arch;`)
		// db.Exec("insert into Message ( id , [from] , [to] , receivedDate , subject , data , mimeParseError , sessionId , attachmentCount , isUnread ) SELECT Id, [From], [To], ReceivedDate, Subject, Data, MimeParseError, SessionId, AttachmentCount, IsUnread FROM Messages;")
	}
	snowflake.SetStartTime(time.Date(2021,1,1,1,1,1,1,time.Now().Location()))
	db.SetMaxOpenConns(5)
	Db = db
}
