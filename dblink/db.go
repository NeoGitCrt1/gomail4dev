package dblink
import "database/sql"
import _ "github.com/mattn/go-sqlite3"
import 	"fmt"

var dbholder *sql.DB  
func init() {
	fmt.Println("open sqllite:")
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		panic(err)
	}
	
	db.Exec("CREATE TABLE Messages ( id TEXT NOT NULL, [from] TEXT, [to] TEXT, receivedDate TEXT NOT NULL, subject TEXT, data BLOB, mimeParseError TEXT, sessionId TEXT, attachmentCount INTEGER NOT NULL DEFAULT 0, isUnread INTEGER NOT NULL DEFAULT 0);")
	db.SetMaxOpenConns(20)
	dbholder = db
}

func Db() *sql.DB {
	return dbholder
}