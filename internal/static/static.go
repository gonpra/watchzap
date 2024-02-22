package static

const (
	// Internal stuff
	APP_NAME = "watchzap"
	WIPE_DB  = "PRAGMA writable_schema = 1;DELETE FROM sqlite_master;PRAGMA writable_schema = 0;VACUUM;PRAGMA integrity_check;"

	// Errors
	INTERNAL_SERVER_ERROR = "An unexpected error has occurred"
	EMPTY_FIELD           = "Mandatory field is empty"
	NO_PARSER_FOUND       = "No parser found for extension"
	INVALID_BYTES         = "Must have even byte slice"
)
