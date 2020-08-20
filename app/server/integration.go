package server

import (
	"github.com/criscola/raindrop-todoist/database"
	"github.com/criscola/raindrop-todoist/logging"
	"github.com/criscola/raindrop-todoist/raindrop"
	"github.com/criscola/raindrop-todoist/todoist"
	"github.com/jackc/pgx/v4"
	vip "github.com/spf13/viper"
)

var conn *pgx.Conn
var globalLogger *logging.StandardLogger

func Start() {
	// Init logger
	globalLogger = logging.New()

	// Connect to DB
	// Try connecting to the db with 2 sec sleep between retries for a maximum of 10 times
	db, err := database.New(vip.GetString("POSTGRES_URL"), globalLogger)
	if err != nil {
		globalLogger.Fatal().Err(err).Msg("Failure setting up database connection")
	}

	// Initialize clients
	raindropClient := raindrop.New(vip.GetString("RAINDROP_TOKEN"),
		vip.GetString("POSTPONED_LABEL_NAME"),
		globalLogger)

	todoistClient := todoist.New(vip.GetString("TODOIST_TOKEN"), globalLogger)

	// Phase 1:
	// Pull every record from database
	storedBookmarks, err := db.GetAllBookmarksWithTodo()
	if err != nil {
		globalLogger.Fatal().Err(err).Msg("Failure getting bookmarks from database")
	}

	// Get data of to-be-read bookmarks except the ones already in db
	postponedReadings, err := raindropClient.GetPostponedReadings(extractIds(storedBookmarks))
	if err != nil {
		globalLogger.Fatal().Err(err).Msg("Failure getting postponed readings from Raindrop API")
	}

	// Foreach bookmark in bookmarks:

	for _, pr := range postponedReadings {
		// Create todoist task of bookmark and get taskId
		var taskId int64
		// Phase 2:
		if len(postponedReadings) != 0 {
			taskId, err = todoistClient.NewReadingTask(pr.Title, pr.Url, pr.Domain)
			if err != nil {
				globalLogger.Fatal().Err(err).Msg("Failure getting postponed readings from Raindrop API")
			}
		}
		// Create entry in db with bookmark.id and taskId
		err := db.InsertBookmarkWithTodo(pr.BookmarkId, taskId)
		if err != nil {
			globalLogger.Fatal().Err(err).Msg("Failure getting postponed readings from Raindrop API")
		}
	}
	// Phase 2:
	// Get new data from sync api of todoist
	/*
	data, err := todoistClient.GetCompletedReadings()
	// Foreach task in data:
	for _, taskId := range readingIds {
		// If task.id is contained in db:
		if db.IsTaskIdPresent(taskId) {
			// Remove label from bookmark in Raindrop
			err := raindropClient.RemoveLabelFromTask(taskId)
			// Remove entry from database
			err = db.RemoveRecordByTaskId(taskId)
		}
	}

	 */
			// Get bookmarkId where task.id == query.taskId
			// Remove label from bookmark in Raindrop
			// Remove entry from database
}

func extractIds(records []database.BookmarkWithTodo) []int64 {
	var ids []int64
	for _, record := range records {
		ids = append(ids, record.BookmarkId)
	}
	return ids
}
