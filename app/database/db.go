package database

import (
	"context"
	"fmt"
	"github.com/criscola/raindrop-todoist/logging"
	"github.com/jackc/pgx/v4"
	"strconv"
	"time"
)

type Db struct {
	db *pgx.Conn
	logger *logging.StandardLogger
}

type BookmarkWithTodo struct {
	BookmarkId int64
	TaskId     int64
}

func New(dsn string, logger *logging.StandardLogger) (*Db, error) {
	var conn *pgx.Conn
	var err error

	// TODO extract retry logic
	// Connect to DB
	// Try connecting to the db with 2 sec sleep between retries for a maximum of 10 times
	for i := 0; i < 10; i++ {
		conn, err = pgx.Connect(context.Background(), dsn)
		if err != nil {
			logger.Err(err).Str("service", "Db").
				Msg("Error connecting to DB, retrying [" + strconv.Itoa(i+1) + "]...")

			time.Sleep(time.Second * 2)
		} else {
			// TODO add info log
			fmt.Println("DB successfully connected.")
			return &Db{
				conn,
				logger,
			}, nil
		}
	}
	defer conn.Close(context.Background())

	return nil, err
}

func (db *Db) GetAllBookmarksWithTodo() (records []BookmarkWithTodo, err error) {
	// Send the query to the server. The returned rows MUST be closed
	// before conn can be used again.
	rows, err := db.db.Query(context.Background(), "SELECT * FROM bookmark_with_todo")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bookmarkId int64
		var taskId int64
		err = rows.Scan(&bookmarkId, &taskId)
		if err != nil {
			return nil, err
		}
		records = append(records, BookmarkWithTodo{bookmarkId, taskId})
	}

	return records, nil
}

func (db *Db) InsertBookmarkWithTodo(bookmarkId, taskId int64) error {
	commandTag, err := db.db.Exec(context.Background(), "INSERT INTO bookmark_with_todo VALUES ($1, $2)",
		&bookmarkId, &taskId)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("error inserting the record with bookmarkId: %d taskID: %d", bookmarkId, taskId)
	}
	return nil
}