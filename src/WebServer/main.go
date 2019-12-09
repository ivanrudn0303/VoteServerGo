package main

import (
    "crypto"
    "crypto/rsa"
    "crypto/sha256"
    "crypto/x509"
    "database/sql"
    "fmt"
    "encoding/json"
    "encoding/pem"
    "BlockChain"
    "log"
    "MailResponder"
    "net/http"
    "encoding/hex"
    _ "github.com/lib/pq"
)

const (
    MSG_SIZE     = 2048
)

var conf map[string]string

type RequestHandler struct {
    sqlBase *sql.DB
    delegate MailResponder.BlockChainDelegate
    blockChainClient BlockChain.BlockChainClient
}

type serialized struct {
    Message string
    Email string
    Sign string
    PublicKey string
}

func (r *RequestHandler)Run(address string) {
    r.blockChainClient.RecvHandler = &(r.delegate)
    r.delegate.Create(&(r.blockChainClient))
    r.blockChainClient.Run(address)
}

func (h *RequestHandler)handler(w http.ResponseWriter, r *http.Request) {
    log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
    http.ServeFile(w, r, "$GOPATH/HtmlFiles" + r.URL.Path)
}

func (h *RequestHandler)handlerApiVote(w http.ResponseWriter, r *http.Request) {
    log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
    r.ParseForm()
    var pub_key string
    if  err := CheckAuth(r.PostForm, h.sqlBase, &pub_key); err != nil {
        log.Print("Vote: ", err.Error())
        http.Error(w, err.Error(), 404)
        return
    }

    bufstr := pub_key
    block, _ := pem.Decode([]byte(bufstr))
    pub2, err2 := x509.ParsePKIXPublicKey(block.Bytes)
    if err2 != nil {
        log.Fatal(err2.Error())
    }
    pub := pub2.(*rsa.PublicKey)
    if v, found := r.PostForm["message"]; !found || len(v[0]) == 0 {
        log.Print("Vote: No Message")
        http.Error(w, "No Message", 404)
        return
    }
    if v, found := r.PostForm["sign"]; !found || len(v[0]) == 0 {
        log.Print("Vote: No Sign")
        http.Error(w, "No Sign", 404)
        return
    }
    var dataToSend serialized
    dataToSend.Message = r.PostForm["message"][0]
    dataToSend.Email = r.PostForm["email"][0]
    dataToSend.Sign = r.PostForm["sign"][0]
    dataToSend.PublicKey = bufstr
    sign_coded, _ := hex.DecodeString(r.PostForm["sign"][0])

    hashed := sha256.Sum256([]byte(dataToSend.Message))

    if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], sign_coded); err != nil {
        log.Print("Vote: ", err.Error())
        return
    }
    sqlStatement := `UPDATE users SET hash = 'started' WHERE email = $1`
    if _, err := h.sqlBase.Exec(sqlStatement, r.PostForm["email"][0]); err != nil {
        log.Print("Vote: ", err.Error())
        http.Error(w, err.Error(), 404)
        return
    }

    var respData []byte
    respData, err2 = json.Marshal(dataToSend)
    if err2 != nil {
        log.Fatal(err2.Error())
    }
    log.Print("Vote: Will send ", len(respData), " bytes")
    buf := make([]byte, MSG_SIZE)
    copy(buf, respData)
    h.delegate.BeforeSend(r.PostForm["email"][0], buf[:])
}

func (h *RequestHandler)handlerApiRegister(w http.ResponseWriter, r *http.Request) {
    log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
    r.ParseForm()
    if  err := Register(r.PostForm, h.sqlBase); err != nil {
        log.Print("Vote: ", err.Error())
        http.Error(w, err.Error(), 404)
        return
    }
 }

func main() {
    conf = loadData()
    var mainHandler RequestHandler
    psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
            "password=%s dbname=%s sslmode=disable",
    conf["ip_sql"], conf["port_sql"], conf["user_sql"], conf["password_sql"], conf["dbname_sql"])

    var err error
    mainHandler.sqlBase, err = sql.Open("postgres", psqlInfo)
    if err != nil {
        panic(err)
    }
    mainHandler.delegate.SqlBaseTo = mainHandler.sqlBase
    defer mainHandler.sqlBase.Close()
    mainHandler.Run(conf["address_blockchain"])
    http.HandleFunc("/", mainHandler.handler)
    http.HandleFunc("/api/vote/", mainHandler.handlerApiVote)
    http.HandleFunc("/api/register/", mainHandler.handlerApiRegister)
    log.Fatal(http.ListenAndServe(conf["address_listen"], nil))
}
