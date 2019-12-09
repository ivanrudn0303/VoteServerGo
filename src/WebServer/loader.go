package main

import (
    "encoding/json"
    "io/ioutil"
    "os"
)

func loadData() map[string]string {
    home, _ := os.UserHomeDir()
    b, err := ioutil.ReadFile(home + "/.vote_server")
    if err != nil {
        panic(err)
    }
    var res map[string]string
    if err = json.Unmarshal(b, &res); err != nil {
        panic(err)
    }
    return res
}
