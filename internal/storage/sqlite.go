package storage

import (
	"database/sql"
	"fmt"
	"strconv"
	// . "khromalabs/keeper/internal/log"
	"regexp"
	"sort"
	"strings"
	_ "github.com/mattn/go-sqlite3"
)

type StorageSqlite struct {
	Template string;
	TemplateData map[string]interface{};
	db *sql.DB;
}

func (s *StorageSqlite) Init(template string, templateData map[string]interface{}) error {
	if err := s.checkDb(); err != nil {
		return err
	}
	s.Template = template
	s.TemplateData = templateData
	var usesTokens bool
	if err := s.checkTable(&usesTokens); err != nil {
		return err
	}
	if usesTokens {
		if err := s.checkTokensRelTable(); err != nil {
			return err
		}
		if err := s.checkTokensTable(); err != nil {
			return err
		}
	}
	return nil
}

func (s *StorageSqlite) checkDb() (error) {
	var err error
	s.db, err = sql.Open("sqlite3", conf.Path["db"])
	if err != nil {
		return fmt.Errorf("Can't open sqlite database: %v", err)
	}
	// defer db.Close()
	return nil
}

func (s *StorageSqlite) checkTable(usesTokens *bool) (error) {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?;`
	row := s.db.QueryRow(query, s.Template)
	var name string
	err := row.Scan(&name)
	switch {
	case err == sql.ErrNoRows:
		query,err := s.schema(usesTokens)
		if err != nil {
			return err
		}
		_, err = s.db.Exec(query)
		return err
	case err != nil:
		return err
	default:
		schema,err := s.schema(usesTokens)
		if err != nil {
			return err
		}
		match,err := s.tableMatchesSchema(s.Template,schema)
		if err != nil {
			return err
		}
		if !match {
			return fmt.Errorf("Table for template %s already present but not maching description", s.Template)
		}
	}
	return nil
}

func (s *StorageSqlite) tableMatchesSchema(table string, schema string) (bool,error) {
	rows, err := s.db.Query(
		"SELECT sql FROM sqlite_schema WHERE name = ?", table)
	if err != nil {
		return false,err
	}
	defer rows.Close()
	if rows.Next() {
		var presentSchema string
		err = rows.Scan(&presentSchema)
		if err != nil {
			return false,err
		}
		return presentSchema + ";" == schema,nil
	}
	return false,nil
}

func (s *StorageSqlite) schema(usesTokens *bool) (string,error) {
	var columns []string
	keys := make([]string, 0)
	for k, _ := range s.TemplateData {
		keys = append(keys, k) // avoid random map
	}
	sort.Strings(keys)
	*usesTokens = false
	for _, label := range keys {
		fields := s.TemplateData[label]
		var columnType string
		var addColumn bool
		for i,v := range fields.(map[string]interface{}) {
			addColumn = true
			if i == "type" {
				switch v {
				case "autodate":
					columnType = "DATE DEFAULT (datetime('now'))"
				case "text":
					columnType = "TEXT"
				case "float":
					columnType = "FLOAT"
				case "integer":
					columnType = "INTEGER"
				case "string":
					columnType = "VARCHAR(255)"
				case "tokens":
					if *usesTokens == true {
						return "",fmt.Errorf("Multiple token fields not allowed")
					}
					*usesTokens = true
					addColumn = false
				default:
					return "",fmt.Errorf("unsupported field type: %s", v)
				}
			}
		}
		if addColumn {
			columnDef := fmt.Sprintf("%s %s", label, columnType)
			columns = append(columns, columnDef)
		}
	}
	return fmt.Sprintf("CREATE TABLE %s (id INTEGER PRIMARY KEY, %s);", s.Template, strings.Join(columns, ", ")),nil
}

func (s *StorageSqlite) checkTokensRelTable() error {
	tokensTable := s.Template + "_tokens"
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?;`
	row := s.db.QueryRow(query, tokensTable)
	var name string
	err := row.Scan(&name)
	switch {
	case err == sql.ErrNoRows:
		_, err = s.db.Exec(s.tokensRelSchema(tokensTable))
		if err != nil {
			return err
		}
		return err
	case err != nil:
		return err
	default:
		match,err := s.tableMatchesSchema(tokensTable, s.tokensRelSchema(tokensTable))
		if err != nil {
			return err
		}
		if !match {
			return fmt.Errorf("Table for tokens relationship for template %s already present but not maching expected schema", s.Template)
		}
	}
	return err
}

func (s *StorageSqlite) tokensRelSchema(tokensTable string) string {
	query := fmt.Sprintf(`CREATE TABLE %s (
		id INTEGER PRIMARY KEY,
		idtoken INTEGER,
		id%s INTEGER,
		FOREIGN KEY (idtoken) REFERENCES _tokens_ (id)
		FOREIGN KEY (id%s) REFERENCES %s (id)
	);`, tokensTable, s.Template, s.Template, s.Template)
	return query
}

func (s *StorageSqlite) checkTokensTable() error {
	table := "_tokens_"
	tokensSchema := `CREATE TABLE ` + table + ` (
		id INTEGER PRIMARY KEY,
		token VARCHAR(255)
	);`
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name='" + table + "';"
	row := s.db.QueryRow(query)
	var name string
	err := row.Scan(&name)
	switch {
	case err == sql.ErrNoRows:
		// LogD.Println("Generating new tokens table...")
		_, err = s.db.Exec(tokensSchema)
		return err
	case err != nil:
		return err
	default:
		match,err := s.tableMatchesSchema(table, tokensSchema)
		if err != nil {
			return err
		}
		if !match {
			return fmt.Errorf(table + " table already present but not maching expected schema")
		}
	}
	return nil
}

func (s *StorageSqlite) Create(fields map[string]string) (int64,error) {
	var columns []string
	var placeholders []string
	var values []interface{}
	tokens := ""
	for k, attr := range s.TemplateData {
		if attr.(map[string]interface{})["type"] == "tokens" {
			// tratar appropiadamente los tokens
			tokens = fields[k]
			continue;
		}
		columns = append(columns, k)
		placeholders = append(placeholders, "?")
		values = append(values, fields[k])
	}
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		s.Template,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	result, err := s.db.Exec(query, values...)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if tokens != "" {
		if err = s.createTokens(tokens,id); err != nil {
			return 0, fmt.Errorf("createTokens: %s", err)
		}
	}
	return id, nil
}

func (s *StorageSqlite) createTokens(tstr string, idtemplate int64) error {
	tokens := strings.Split(tstr, ",")
	var c int64 
	var idtoken int64
	for _,t := range tokens {
		t = strings.Trim(t, " ")
		if err := s.db.QueryRow("SELECT count(*) FROM _tokens_ WHERE token = ?", t).Scan(&c); err != nil {
			return err
		}
		if c == 0 {
			res, err := s.db.Exec("INSERT INTO _tokens_ (token) VALUES (?)", t)
			if err != nil {
				return err
			}
			if idtoken, err = res.LastInsertId(); err != nil {
				return err
			}
		} else {
			if err := s.db.QueryRow("SELECT id FROM _tokens_ WHERE token = ?", t).Scan(&idtoken); err != nil {
				return err
			}
		}
		query := fmt.Sprintf("INSERT INTO %s_tokens (idtoken,id%s) VALUES (?,?)", s.Template, s.Template)
		if _, err := s.db.Exec(query, idtoken, idtemplate); err != nil {
			return err
		}
	}
	return nil
}

func (s *StorageSqlite) Read(filter string, i int) ([]interface{},int,error) {
	tokenColumn := ""
	for k,_ := range s.TemplateData {
		attr := s.TemplateData[k].(map[string]interface{})
		if attr["type"] == "tokens" {
			tokenColumn = k
			// there should be only one tokens column
			break
		}
	}
	filterColumns,filterValues,filterTokens,err := s.queryFilter(filter,tokenColumn)
	if err != nil {
		return nil,-1,err
	}
	var rows *sql.Rows
	query := "SELECT * FROM " + s.Template + s.readWhere(filterColumns, filterValues,filterTokens)
	if filterTokens != "" {
		filterValues = append(filterValues, filterTokens)
	}
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil,-1,err
	}
	rows, err = stmt.Query(filterValues...)
	if err != nil {
		return nil,-1,err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil,-1,err
	}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil,-1,err
	}
	values := make([]sql.RawBytes,len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	p := make([]interface{}, 0)
	for rows.Next() {
		row := make(map[string]string,0)
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil,-1,err
		}
		for i, value := range values {
			c := columns[i]
			if string(value) == "" {
				continue
			}
			if columnTypes[i].DatabaseTypeName() == "TEXT" {
				if conf.Miniread {
					continue
				}
			}
			row[c] = string(value)
		}
		if tokenColumn != "" {
			tokens,err := s.readTokens(string(values[0]))
			row[tokenColumn] = tokens
			if err != nil {
				return nil,-1,err
			}
		}
		p = append(p, row)
		row = nil
		i++
	}
	if err = rows.Err(); err != nil {
		return nil,-1,err
	}
	return p,i,nil
}

func (s *StorageSqlite) queryFilter(filter string, tokenColumn string) ([]string,[]interface{},string,error) {
	var filterTokens []string
	var err error
	if filter != "" {
		// @TODO: Only one token is accepted in the filter by now
		// function splitFilter works as intended but parameter library (?)
		// is removing the quotes
		filterTokens = strings.Split(filter, ",")
		// filterTokens,err = s.splitFilter(filter)
		if err != nil {
			return nil,nil,"",err
		}
	} else {
		filterTokens = nil
	}
	regex := "^[A-z][A-z0-9]*:([A-z0-9%,]+|\"[\\w\\s%,]+\")$"
	rows, err := s.db.Query("PRAGMA table_info(" + s.Template + ")")
	if err != nil {
		return nil,nil,"",err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil,nil,"",err
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	fields := make(map[string]string, 0)
	outColumns := make([]string, 0)
	outValues := make([]interface{}, 0)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		fields[string(values[1])] = string(values[2])
	}
	outTokensFilter := ""
	for _,t := range filterTokens {
		t = strings.Trim(t, " ")
		match, err := regexp.MatchString(regex, t)
		if err != nil || !match {
			return nil,nil,"",fmt.Errorf("Malformed filter: " + t)
		}
		parts := strings.Split(t, ":")
		if len(parts) == 2 {
			if _, ok := fields[parts[0]]; ok {
				outColumns = append(outColumns, parts[0])
				outValues = append(outValues, parts[1])
			} else if parts[0] == tokenColumn {
				outTokensFilter = parts[1]
			}
		}
	}
	return outColumns,outValues,outTokensFilter,nil
}

func (s *StorageSqlite) readWhere(columns []string, values []interface{},tokenFilter string) string {
	var conditions []string
	for i, column := range columns {
		if strings.Contains(values[i].(string), "%") {
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", column))
		} else {
			conditions = append(conditions, fmt.Sprintf("%s = ?", column))
		}
	}
	whereClauses := make([]string,0)
	if len(conditions) > 0 {
		whereClauses = append(whereClauses,strings.Join(conditions, " AND "))
	}
	if tokenFilter != "" {
		t := s.Template
		whereClauses = append(whereClauses, t + ".id IN (SELECT " + t + "_tokens.id" + t + " FROM " + t + "_tokens LEFT JOIN _tokens_ on " + t + "_tokens.idtoken = _tokens_.id WHERE _tokens_.token = ? GROUP BY " + t +"_tokens.id" + t + ")")
	}
	if len(whereClauses) > 0 {
		return " WHERE " + strings.Join(whereClauses, " OR ")
	}
	return ""
}

func (s *StorageSqlite) UpdateOrDelete(fields map[string]string, opUpdate bool) error {
	var columns []string
	var values []interface{}
	tokens := ""
	var err error
	for k, attr := range s.TemplateData {
		if attr.(map[string]interface{})["type"] == "tokens" {
			tokens = fields[k]
			continue;
		}
		columns = append(columns, k + "=?")
		values = append(values, fields[k])
	}
	idstr, ok := fields["id"]
	if !ok {
		return fmt.Errorf("Missing id field")
	}
	id, err := strconv.ParseInt(idstr, 10, 64) 
	if err != nil {
		return fmt.Errorf("Invalid id field")
	}
	if id <= 0 {
		return fmt.Errorf("Invalid id field value: %d", id)
	}
	var query string
	if opUpdate {
		query = fmt.Sprintf(
			"UPDATE %s SET %s WHERE id=%d",
			s.Template,
			strings.Join(columns, ", "),
			id,
		)
	} else {
		query = fmt.Sprintf(
			"DELETE FROM %s WHERE id=%d",
			s.Template,
			id,
		)
	}
	_, err = s.db.Exec(query, values...)
	if err != nil {
		return err
	}
	if err := s.deleteTokens(id); err != nil {
		return err
	}
	if tokens != "" {
		if opUpdate {
			s.createTokens(tokens,id)
		}
	}
	return nil
}

func (s *StorageSqlite) readTokens(id string) (string,error) {
	query := "select _tokens_.token from %s_tokens left join _tokens_ on %s_tokens.idtoken = _tokens_.id where id%s = ?"
	stmt, err := s.db.Prepare(fmt.Sprintf(query, s.Template, s.Template, s.Template))
	if err != nil {
		return "",err
	}
	rows, err := stmt.Query(id)
	if err != nil {
		return "",err
	}
	tokens := make([]string,0)
	var token string
	for rows.Next() {
		if err = rows.Scan(&token); err != nil {
			return "",err
		}
		tokens = append(tokens, token)
	}
	return strings.Join(tokens, ", "),nil
}

func (s *StorageSqlite) splitFilter(input string) ([]string, error) {
    res := []string{}
    var beg int
    var inString bool

    for i := 0; i < len(input); i++ {
        if input[i] == ',' && !inString {
            res = append(res, input[beg:i])
            beg = i+1
        } else if input[i] == '"' {
            if !inString {
                inString = true
            } else if i > 0 && input[i-1] != '\\' {
                inString = false
            }
        }
    }
    return append(res, input[beg:]),nil
}

func (s *StorageSqlite) deleteTokens(id int64) error {
	q := fmt.Sprintf("DELETE FROM %s_tokens WHERE id%s = ?", s.Template, s.Template)
	if _, err := s.db.Exec(q, &id); err != nil {
		return err
	}
	// @TODO: Find orphan tags and delete them
	return nil
}
