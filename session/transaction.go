package session

import "GeeORM/logger"

// Begin  得到 *sql.Tx 对象，赋值给 s.tx
func (s *Session) Begin() (err error) {
	logger.Info("transaction begin")
	if s.tx, err = s.db.Begin(); err != nil {
		logger.Error(err)
		return
	}
	return
}

func (s *Session) Commit() (err error) {
	logger.Info("transaction commit")
	if err = s.tx.Commit(); err != nil {
		logger.Error(err)
	}
	return
}

func (s *Session) Rollback() (err error) {
	logger.Info("transaction rollback")
	if err = s.tx.Rollback(); err != nil {
		logger.Error(err)
	}
	return
}
