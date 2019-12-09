package main

import (
    "errors"
    "net/url"
    "database/sql"
    "sync"
)

var sql_lock sync.Mutex

func validate(params url.Values) error {
    if v, found := params["password"]; !found || len(v[0]) == 0 {
        return errors.New("No Password")
    }

    if v, found := params["email"]; !found || len(v[0]) == 0 {
        return errors.New("No Email")
    }
    return nil
}

func CheckAuth(params url.Values, db *sql.DB, public_key *string) error {
    if err := validate(params); err != nil {
        return err
    }

    var hash sql.NullString
    sql_lock.Lock()
    defer sql_lock.Unlock()
    sqlStatement := `SELECT hash, public_key FROM users WHERE email=$1 AND password=$2`
    err := db.QueryRow(sqlStatement, params["email"][0], params["password"][0]).Scan(&hash, public_key)
    if err != nil {
        return err
    }
    if hash.Valid {
        return errors.New("Already Voted")
    }
    return nil
}

func Register(params url.Values, db *sql.DB) error {
    if err := validate(params); err != nil {
        return err
    }

    if v, found := params["public_key"]; !found || len(v[0]) == 0 {
        return errors.New("No Public Key")
    }

    var tmp string
    sql_lock.Lock()
    defer sql_lock.Unlock()
    sqlStatement := `SELECT hash FROM users WHERE email=$1`
    if err := db.QueryRow(sqlStatement, params["email"][0]).Scan(&tmp); err != sql.ErrNoRows {
        return errors.New("Already exists")
    }
    sqlStatement = `INSERT INTO users (email, password, public_key) VALUES ($1, $2, $3)`
    if _, err := db.Exec(sqlStatement, params["email"][0], params["password"][0], params["public_key"][0]); err != nil {
        return err
    }
    return nil
}
