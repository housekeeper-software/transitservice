package opensips

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"jingxi.cn/transitservice/conf"
	"strings"
	"time"
)

type SubService struct {
	DB         *sql.DB
	serverConf *conf.ServerConfig
}

func NewSubService(conf *conf.ServerConfig) *SubService {
	return &SubService{
		DB:         nil,
		serverConf: conf,
	}
}

func InitDatabase(conf *conf.ServerConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", conf.Mysql.Url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(conf.Mysql.MaxOpenConns)
	db.SetMaxIdleConns(conf.Mysql.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(conf.Mysql.ConnMaxLifeTime) * time.Second)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	logrus.Infof("Connected to DB %s successfully", conf.Mysql.Url)
	return db, nil
}

func (s *SubService) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}

func (s *SubService) EnsureDatabase() error {
	if s.DB == nil {
		db, err := InitDatabase(s.serverConf)
		if err != nil {
			logrus.Errorf("Open Mysql(%s) error : %+v", s.serverConf.Mysql.Url, err)
			return err
		}
		s.DB = db //connect success!
	}
	return nil
}

func (s *SubService) AddUser(user *User) error {
	err := s.EnsureDatabase()
	if err != nil {
		return err
	}

	var query strings.Builder
	_, err = fmt.Fprintf(&query, "insert into %s(username,domain,password,email_address,ha1,ha1b,rpid) VALUES (?,?,?,?,?,?,?)",
		s.serverConf.Mysql.Table)
	if err != nil {
		logrus.Errorf("Build SQL string(%s) error %+v when Add User(%s)", query.String(), err, user.Username)
		return err
	}

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := s.DB.PrepareContext(ctx, query.String())
	if err != nil {
		logrus.Errorf("Preparing SQL statement(%s) error %+v when Add User(%+v)", query.String(), err, user)
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, user.Username, user.Domain, user.Password, user.EmailAddress, user.Ha1, user.Ha1b, user.Rpid)
	if err != nil {
		logrus.Errorf("Exec SQL statement(%s) error %+v when Add User(%+v)", query.String(), err, user)
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		logrus.Errorf("Finding rows affected error %+v when add User(%+v)", err, user)
		return err
	}
	logrus.Infof("%d rows affected when insert User (%+v)", rows, user)

	id, err := res.LastInsertId()
	if err != nil {
		logrus.Errorf("Fetching last id error: %+v when add User(%+v)", err, user)
		return err
	}
	logrus.Infof("The last inserted row id:%d when add User(%+v)", id, user)
	return nil
}

func (s *SubService) UpdateUser(user *User) error {
	err := s.EnsureDatabase()
	if err != nil {
		return err
	}
	var query strings.Builder
	_, err = fmt.Fprintf(&query, "UPDATE %s set username=?, domain=?, password=?, email_address=?, ha1=?, ha1b=? where username='%s'",
		s.serverConf.Mysql.Table,
		user.Username)
	if err != nil {
		logrus.Errorf("Build SQL string(%s) error %+v when Update User(%+v)", query.String(), err, user)
		return err
	}

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := s.DB.PrepareContext(ctx, query.String())
	if err != nil {
		logrus.Errorf("Preparing SQL statement(%s) error %+v when Update User(%+v)", query.String(), err, user)
		return err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, user.Username, user.Domain, user.Password, user.EmailAddress, user.Ha1, user.Ha1b)
	if err != nil {
		logrus.Errorf("Exec SQL statement(%s) error %+v when Update User(%+v)", query.String(), err, user)
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		logrus.Errorf("Finding rows affected error %+v when Update User(%+v)", err, user)
		return err
	}
	logrus.Infof("%d rows affected when Update User (%+v)", rows, user)
	return nil
}

func (s *SubService) DeleteUser(username string) error {
	err := s.EnsureDatabase()
	if err != nil {
		return err
	}
	var query strings.Builder
	_, err = fmt.Fprintf(&query, "DELETE from %s where username=?",
		s.serverConf.Mysql.Table)
	if err != nil {
		logrus.Errorf("Build SQL string(%s) error %+v when Delete User(%s)", query.String(), err, username)
		return err
	}

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := s.DB.PrepareContext(ctx, query.String())
	if err != nil {
		logrus.Errorf("Preparing SQL statement(%s) error %+v when Delete User(%s)", query.String(), err, username)
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, username)
	if err != nil {
		logrus.Errorf("Exec SQL statement(%s) error %+v when Delete User(%s)", query.String(), err, username)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		logrus.Errorf("Finding rows affected error %+v when Delete User(%s)", err, username)
		return err
	}
	logrus.Infof("%d rows affected when Delete User (%s)", rows, username)
	return nil
}

//bool: false is database error,so we return error to client, true is db operation ok, but not found row
func (s *SubService) GetUser(username string) (*User, error, bool) {
	err := s.EnsureDatabase()
	if err != nil {
		return nil, err, false
	}
	var query strings.Builder
	_, err = fmt.Fprintf(&query,
		"select username,domain,password,email_address,ha1,ha1b from %s where username = '%s'",
		s.serverConf.Mysql.Table, username)
	if err != nil {
		logrus.Errorf("Build SQL string(%s) error %+v when Select User(%s)", query.String(), err, username)
		return nil, err, false
	}

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := s.DB.PrepareContext(ctx, query.String())
	if err != nil {
		logrus.Errorf("Preparing SQL statement(%s) error %+v when Select User(%s)", query.String(), err, username)
		return nil, err, false
	}
	defer stmt.Close()

	var user User
	row := stmt.QueryRowContext(ctx)
	if err := row.Scan(&user.Username, &user.Domain, &user.Password, &user.EmailAddress, &user.Ha1, &user.Ha1b); err != nil {
		logrus.Errorf("Error %+v when ROW Scan SQL statement(%s)", err, username)
		return nil, err, true //user not found
	}
	logrus.Infof("select User(%s) success: %+v", username, user)
	return &user, nil, true //user existed
}
