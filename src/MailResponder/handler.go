package MailResponder

import (
    "database/sql"
    "log"
    "sync"
)

const (
    BLOCKS_MAX_WAIT = 64
)

type Sender interface {
    Send([]byte) error
}

type BlockChainDelegate struct {
    emailQueue chan string
    sendLock sync.Mutex
    blockChainClient Sender
    SqlBaseTo *sql.DB
}

func (d *BlockChainDelegate)Create(delegate Sender) {
    d.blockChainClient = delegate
    d.emailQueue = make(chan string, BLOCKS_MAX_WAIT)
}

func (d *BlockChainDelegate)BeforeSend(email string, buf []byte) error {
    d.sendLock.Lock()
    defer d.sendLock.Unlock()
    if err := d.blockChainClient.Send(buf); err != nil {
        log.Fatal(err.Error())
        return err
    }
    d.emailQueue <- email
    log.Print("sent to blockchain from email: ", email, " ", string(buf))
    return nil
}

func (d *BlockChainDelegate)Callback(buf []byte) {
    email := <- d.emailQueue
    log.Print("recv from blockchain to email: ", email, " ", string(buf))
    sqlStatement := `UPDATE users SET hash = $2 WHERE email = $1`
    if _, err := d.SqlBaseTo.Exec(sqlStatement, email, string(buf)); err != nil {
        log.Print(err.Error())
    }
}
