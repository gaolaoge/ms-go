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

func (db *MsDb) New(data interface{}) *MsSession {
	m := &MsSession{
		db: db,
	}

	if data != nil {
		t := reflect.TypeOf(data)
		if t.Kind() != reflect.Pointer {
			panic(errors.New("data must be pointer"))
		}
		tVar := t.Elem()
		if m.tableName == "" {
			m.tableName = m.db.Prefix + strings.ToLower(Name(tVar.Name()))
		}
	}

	return m
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
	updateParam strings.Builder
	whereParam  strings.Builder
	whereValues []any
}

func (s *MsSession) Table(name string) *MsSession {
	s.tableName = name
	return s
}

func (s *MsSession) Insert(data any) (int64, int64, error) {
	// insert into table (xxx,xxx) values (?,?)
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

func (s MsSession) InsertBatch(data []any) (int64, int64, error) {
	if len(data) == 0 {
		return -1, -1, errors.New("no data insert")
	}
	s.fieldNames(data[0])
	query := fmt.Sprintf("insert into %s (%s) values ", s.tableName, strings.Join(s.fieldName, ","))

	var sb strings.Builder
	sb.WriteString(query)
	for index, _ := range data {
		sb.WriteString("(")
		sb.WriteString(strings.Join(s.placeHolder, ","))
		sb.WriteString(")")
		if index < len(data)-1 {
			sb.WriteString(",")
		}
	}
	s.batchValues(data)
	s.db.Logger.Info(sb.String())

	stmt, err := s.db.db.Prepare(sb.String())
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

func (s *MsSession) batchValues(data []any) {
	s.values = make([]any, 0)

	for _, val := range data {
		t := reflect.TypeOf(val)
		v := reflect.ValueOf(val)
		if t.Kind() != reflect.Pointer {
			panic("data mast be pointer")
		}

		tVar := t.Elem()
		vVar := v.Elem()
		for i := 0; i < tVar.NumField(); i++ {
			fieldName := t.Field(i).Name
			tag := tVar.Field(i).Tag
			sqlTag := tag.Get("orm")
			if sqlTag == "" {
				sqlTag = strings.ToLower(Name(fieldName))
			} else {
				if strings.Contains(sqlTag, "auto_increment") {
					// 主键id 自增长
					continue
				}
			}
			id := vVar.Field(i).Interface()
			if strings.ToLower(sqlTag) == "id" && IsAutoId(id) {
				continue
			}
			s.values = append(s.values, vVar.Field(i).Interface())
		}
	}
}

func (s *MsSession) Update(data ...any) (int64, int64, error) {
	// Update("age", 1) or Update(&user) 都支持
	if len(data) == 0 || len(data) >= 2 {
		return -1, -1, errors.New("param not valid")
	}

	single := false
	if len(data) == 1 {
		single = true
	}

	// update table set age=?,name=? where id=?
	if !single {
		if s.updateParam.String() != "" {
			s.updateParam.WriteString(",")
		}
		s.updateParam.WriteString(data[0].(string))
		s.updateParam.WriteString("=?")
		s.values = append(s.values, data[1])
	} else {
		updateData := data[0]
		t := reflect.TypeOf(updateData)
		v := reflect.ValueOf(updateData)

		if t.Kind() != reflect.Pointer {
			panic(errors.New("updateData mast be pointer"))
		}

		tVar := t.Elem()
		vVar := v.Elem()
		if s.tableName == "" {
			s.tableName = s.db.Prefix + strings.ToLower(Name(tVar.Name()))
		}

		for i := 0; i < tVar.NumField(); i++ {
			fieldName := tVar.Field(i).Name
			tag := tVar.Field(i).Tag
			sqlTag := tag.Get("orm")

			if sqlTag == "" {
				sqlTag = strings.ToLower(Name(fieldName))
			} else {
				if strings.Contains(sqlTag, ",") {
					sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
				}
			}

			id := vVar.Field(i).Interface()
			if strings.ToLower(sqlTag) == "id" && IsAutoId(id) {
				continue
			}
			s.updateParam.WriteString(sqlTag)
			s.updateParam.WriteString("=?")
			s.values = append(s.values, vVar.Field(i).Interface())
		}
	}

	query := fmt.Sprintf("UPDATE %s SET %s", s.tableName, s.updateParam.String())

	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())

	s.db.Logger.Info(sb.String())

	stmt, err := s.db.db.Prepare(sb.String())
	if err != nil {
		return -1, -1, err
	}

	s.values = append(s.values, s.whereValues...)

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

func (s *MsSession) UpdateParam(field string, value interface{}) *MsSession {
	if s.whereParam.String() != "" {
		s.whereParam.WriteString(" , ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString("=?")
	s.whereValues = append(s.values, value)
	return s
}

func (s *MsSession) UpdateMap(data map[string]any) *MsSession {
	for k, v := range data {
		if s.whereParam.String() != "" {
			s.whereParam.WriteString(" , ")
		}
		s.whereParam.WriteString(k)
		s.whereParam.WriteString("=?")
		s.whereValues = append(s.values, v)
	}
	return s
}

func (s *MsSession) Where(field string, value interface{}) *MsSession {
	// id=1
	if s.whereParam.String() == "" {
		s.whereParam.WriteString("where ")
	} else {
		s.whereParam.WriteString(" and ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString("=?")
	s.whereValues = append(s.whereValues, value)
	return s
}

func (s *MsSession) Or(field string, value interface{}) *MsSession {
	// id=1
	if s.whereParam.String() == "" {
		s.whereParam.WriteString("where ")
	} else {
		s.whereParam.WriteString(" or ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString("=?")
	s.whereValues = append(s.whereValues, value)
	return s
}

func (s *MsSession) SelectOne(data interface{}, fields ...string) error {
	// select * from table where id=1000
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		panic(errors.New("data must be pointer"))
	}

	fieldStr := "*"
	if len(fields) > 0 {
		fieldStr = strings.Join(fields, ",")
	}

	query := fmt.Sprintf("select %s from %s ", fieldStr, s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())

	s.db.Logger.Info(sb.String())

	stmt, err := s.db.db.Prepare(sb.String())
	if err != nil {
		return err
	}
	rows, err := stmt.Query(s.whereValues...)
	if err != nil {
		return err
	}
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	fieldScan := make([]interface{}, len(columns))
	for i := range fieldScan {
		fieldScan[i] = &values[i]
	}

	if rows.Next() {
		err := rows.Scan(fieldScan...)
		if err != nil {
			return err
		}

		tVar := t.Elem()
		vVar := reflect.ValueOf(data).Elem()
		for i := 0; i < tVar.NumField(); i++ {
			fieldName := tVar.Field(i).Name
			tag := tVar.Field(i).Tag
			sqlTag := tag.Get("orm")

			if sqlTag == "" {
				sqlTag = strings.ToLower(Name(fieldName))
			} else {
				if strings.Contains(sqlTag, ",") {
					sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
				}
			}

			for j, colName := range columns {
				if sqlTag == colName {
					target := values[j]
					targetValue := reflect.ValueOf(target)
					fieldType := tVar.Field(j).Type
					result := reflect.ValueOf(targetValue.Interface()).Convert(fieldType)
					vVar.Field(i).Set(result)
				}
			}

		}
	}

	return nil
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
