package app

import (
	"database/sql"
	"sync"
)

type App struct {
	DB                 *sql.DB
	Mu                 sync.Mutex
	ActiveAnnouncement interface{}
}
