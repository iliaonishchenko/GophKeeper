package repository

import "database/sql"

type baseRepo struct {
	db         *sql.DB
	classifier ErrorClassifier
}
