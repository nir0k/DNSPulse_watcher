package sqldb

import (
	"database/sql"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)


func ConnectToDB(dbPath string) error {
    var err error
    AppDB, err = sql.Open("sqlite3", dbPath)
    return err
}

func CheckConnectDB() bool {
    err := AppDB.Ping()
    if err != nil {
        ConnectToDB(DBName)
    }
    return err == nil
}



func isUniqueViolationError(err error) bool {
    // Adjust the error checking based on the specific error message or type returned by your database driver
    // The below is an example and may need modification based on your SQLite driver
    return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func convertToCamelCase(snakeStr string) string {
    // Implement conversion logic here
    // This is an example logic, adjust as needed based on your actual label naming convention
    parts := strings.Split(snakeStr, "_")
    for i, part := range parts {
        parts[i] = strings.Title(part)
    }
    return strings.Join(parts, "")
}
