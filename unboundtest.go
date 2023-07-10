package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

const unboundConfig = "/etc/unbound/unbound.conf"

// A regexp for reasonable close-to-valid DNS names
var dnsish = regexp.MustCompile("^[A-Za-z0-9-_.]+$")

// Only one Unbound should run at once, otherwise listen port will collide
var unboundMutex sync.Mutex

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/conf", configHandler)
	http.HandleFunc("/q", queryHandler)
	http.HandleFunc("/m/", memoryHandler)
	http.ListenAndServe(":1232", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	file, err := os.Open("index.html")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	io.Copy(w, file)
	file.Close()
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(unboundConfig)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	io.Copy(w, file)
	file.Close()
}

type recorder struct {
	sync.Mutex
	archive map[string][]byte
}

func (r *recorder) store(b []byte) string {
	var id [5]byte
	rand.Read(id[:])
	idStr := base32.StdEncoding.EncodeToString(id[:])

	r.Lock()
	defer r.Unlock()
	r.archive[idStr] = b
	return idStr
}

func (r *recorder) get(idStr string) []byte {
	r.Lock()
	defer r.Unlock()
	return r.archive[idStr]
}

var memory = &recorder{
	archive: make(map[string][]byte),
}

func memoryHandler(w http.ResponseWriter, r *http.Request) {
	components := strings.Split(r.URL.Path[1:], "/")
	if len(components) < 4 {
		http.NotFound(w, r)
		return
	}
	body := memory.get(components[3])
	if body == nil {
		http.NotFound(w, r)
		return
	}
	w.Write(body)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	typ, ok := dns.StringToType[queryParams.Get("type")]
	if !ok {
		http.NotFound(w, r)
		return
	}
	qname := queryParams.Get("qname")
	if !dnsish.MatchString(qname) {
		http.NotFound(w, r)
		return
	}

	var buf = new(bytes.Buffer)
	doQuery1(r.Context(), qname, typ, buf)
	idStr := memory.store(buf.Bytes())
	http.Redirect(w, r, fmt.Sprintf("/m/%s/%s/%s", dns.TypeToString[typ], qname, idStr), http.StatusTemporaryRedirect)
}

func doQuery1(ctx context.Context, q string, typ uint16, w io.Writer) {
	fmt.Fprintf(w, "Query results for %s %s\n", dns.TypeToString[typ], q)
	unboundMutex.Lock()
	defer unboundMutex.Unlock()
	err := doQuery(ctx, q, typ, w)
	if err != nil {
		fmt.Fprintf(w, "\n\nError running query: %s\n", err)
	}
}

func doQuery(ctx context.Context, q string, typ uint16, w io.Writer) error {
	// Automatically make the query name fully-qualified.
	if !strings.HasSuffix(q, ".") {
		q = q + "."
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	cmd := exec.CommandContext(ctx, "unbound", "-d", "-c", unboundConfig)
	defer func() {
		cancel()
		cmd.Wait()
	}()
	// Unbound logs will be sent on this channel once done.
	logs := make(chan []byte)
	pipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	defer func() {
		// Kill Unbound, then finish reading off the logs.
		cancel()
		w.Write(<-logs)
		cmd.Wait()
	}()
	go func() {
		// Read Unbound's stderr logs as they come in, both to avoid blocking and to
		// ensure we show what the logs said even if the query times out.
		buf := new(bytes.Buffer)
		fmt.Fprintln(buf, "----- Unbound logs -----")
		io.Copy(buf, pipe)
		logs <- buf.Bytes()
	}()

	// Wait for Unbound to start listening
	time.Sleep(time.Second)
	m := new(dns.Msg)
	m.SetQuestion(q, typ)
	m.AuthenticatedData = true
	m.SetEdns0(4096, true)
	c := new(dns.Client)
	// The default timeout in the dns package is 2 seconds, which is too short for
	// some domains. Increase to 30 seconds, limited by the context if applicable.
	// Also retry on timeouts.
	c.Timeout = time.Second * 30
	for {
		in, _, err := c.ExchangeContext(ctx, m, "127.0.0.1:1053")
		if err != nil {
			return err
		}
		if err == nil {
			fmt.Fprintf(w, "\nResponse:\n%s\n", in)
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			continue
		}
	}
}
