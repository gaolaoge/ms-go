package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	msLog "github.com/gaolaoge/ms-go/log"
)

/* orm 的本质就是拼接 sql  */

type MsDb struct {
	db     *sql.DB
	Logger *msLog.Logger
	Prefix string
}

func (db *MsDb) SetMaxIdleConns(n int) *MsDb {
	db.db.SetMaxIdleConns(n)
	return db
}

func (db *MsDb) SetMaxOpenConns(n int) *MsDb {
	db.db.SetMaxOpenConns(n)
	return db
}

func (db *MsDb) SetConnMaxLifetime(t time.Duration) *MsDb {
	db.db.SetConnMaxLifetime(t)
	return db
}

func (db *MsDb) SetConnMaxIdleTime(t time.Duration) *MsDb {
	db.db.SetConnMaxIdleTime(t)
	return db
}

func (db *MsDb) New() *MsSession {
	return &MsSession{
		db: db,
	}
}

func (db MsDb) Close() error {
	return db.db.Close()
}

type MsSession struct {
	db          *MsDb
	tableName   string
	fieldName   []string
	placeHolder []string
	values      []any
}

func (s *MsSession) Table(name string) *MsSession {
	s.tableName = name
	return s
}

func (s *MsSession) Insert(data any) (int64, int64, error) {
	// 这里保证每个 session 都是一个独立的不受影响的操作
	s.fieldNames(data)
	query := fmt.Sprintf(
		"insert into %s (%s) values (%s)",
		s.tableName,
		strings.Join(s.fieldName, ""),
		strings.Join(s.placeHolder, ""),
	)
	s.db.Logger.Info(query)

	stmt, err := s.db.db.Prepare(query)
	if err != nil {
		return -1, -1, err
	}

	r, err := stmt.Exec(s.values...)
	if err != nil {
		return -1, -1, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return -1, -1, err
	}

	affected, err := r.RowsAffected()
	if err != nil {
		return -1, -1, err
	}

	return id, affected, nil
}

func (s *MsSession) fieldNames(data any) {
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)

	if t.Kind() != reflect.Pointer {
		panic(errors.New("data must be pointer"))
	}

	tVar := t.Elem()
	vVar := v.Elem()

	for i := 0; i < tVar.NumField(); i++ {
		fieldName := tVar.Field(i).Name
		tag := tVar.Field(i).Tag
		sqlTag := tag.Get("orm")

		if sqlTag == "" {
			sqlTag = strings.ToLower(Name(fieldName))
		} else {
			if strings.Contains(sqlTag, "auto_increment") {
				// 自增长id主键
				continue
			}
			if strings.Contains(sqlTag, ",") {
				// 多字段，tagName 在首位
				sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
			}
			id := vVar.Field(i).Interface()
			// 这里判断结构体是否存在字段 id ，若没有则自增长赋值
			if strings.ToLower(sqlTag) == "id" && IsAutoId(id) {
				continue
			}
		}
		s.fieldName = append(s.fieldName, sqlTag)
		s.placeHolder = append(s.placeHolder, "?")
		s.values = append(s.values, vVar, vVar.Field(i).Interface())
	}
}

func IsAutoId(id any) bool {
	t := reflect.TypeOf(id)
	switch t.Kind() {
	case reflect.Int64:
		if id.(int64) <= 0 {
			return true
		}
	case reflect.Int32:
		if id.(int32) <= 0 {
			return true
		}
	case reflect.Int:
		if id.(int) <= 0 {
			return true
		}
	}
	return false
}

func Name(name string) string {
	names := name[:]
	lastIndex := 0
	var sb strings.Builder
	for index, value := range names {
		if value >= 65 && value <= 90 {
			// 大写字母
			if index == 0 {
				continue
			}
			sb.WriteString(name[lastIndex:index])
			sb.WriteString("_")
			lastIndex = index
		}
	}
	if lastIndex < len(names)-1 {
		sb.WriteString(name[lastIndex:])
	}
	return sb.String()
}

func Open(dirverName, source string) (*MsDb, error) {
	db, err := sql.Open(dirverName, source)
	if err != nil {
		return nil, err
	}

	// 最大空闲连接数，若不配置默认为 2
	db.SetMaxIdleConns(5)
	// 最大连接数，若不配置默认无上限
	db.SetMaxOpenConns(100)
	// 最大存活时间
	db.SetConnMaxLifetime(time.Minute * 3)
	// 最大空闲连接存活时间
	db.SetConnMaxIdleTime(time.Minute * 1)

	msdb := &MsDb{
		db:     db,
		Logger: msLog.Default(),
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return msdb, nil
}
