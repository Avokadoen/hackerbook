package app

import (
	db "gitlab.com/avokadoen/softsecoblig2/lib/database"
)

type Server struct {
	Port     string
	Database db.DbState
}
