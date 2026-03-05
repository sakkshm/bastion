package api

const (
	HealthEndpoint = "/health"

	CreateSessionEndpoint    = "/session/create"
	SessionBaseEndpoint      = "/session/{id}"
	StartSessionEndpoint     = "/start"
	GetSessionStatusEndpoint = "/status"
	SessionExecuteEndpoint   = "/exec"
	GetSessionLogsEndpoint   = "/logs"
)
