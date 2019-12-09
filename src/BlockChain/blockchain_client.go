package BlockChain

import (
    "errors"
    "log"
    "net"
)

const (
    HASH_SIZE = 64
)

type Recver interface {
    Callback(buf []byte)
}

type BlockChainClient struct {
    conn net.Conn
    quit chan bool
    RecvHandler Recver
}

func (s *BlockChainClient)Run(address string) {
    var err error
    if s.conn, err = net.Dial("tcp", address); err != nil {
        log.Fatal("Unable to connect")
    }

    log.Print("started BlockChain Client")
    s.quit = make(chan bool)
    go func() {
        buf := make([]byte, HASH_SIZE)
        for {
            select {
            case <- s.quit:
                return
            default:
                if err := s.Recv(buf); err != nil {
                    log.Print(err.Error())
                    return
                }
            }
        }
    }()
}

func (s *BlockChainClient)Stop() {
    log.Print("stopped BlockChain Client")
    s.quit <- true
    s.conn.Close()
}

func (s *BlockChainClient)Send(buf []byte) error {
    count := 0
    for count < len(buf) {
        byteSent, err := s.conn.Write(buf[count:])
        if err != nil {
            return err
        }
        count += byteSent
    }
    return nil
}

func (s *BlockChainClient)Recv(buf []byte) error {
    count := 0
    for count < len(buf) {
        byteSent, err := s.conn.Read(buf[count:])
        if err != nil {
            return err
        }
        if byteSent == 0 {
            return errors.New("Connection closed")
        }
        count += byteSent
    }
    s.RecvHandler.Callback(buf)
    return nil
}
