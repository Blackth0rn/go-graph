package main

import (
	"database/sql"
	"flag"
	"fmt"
	"go/build"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3" //no name needed as it implements the database/sql interface
	"github.com/ugorji/go/codec"
)

type Context struct {
	db *sql.DB
	mh codec.Handle
}

func (c *Context) Close() {
	c.db.Close()
}

func NewContext(db_path string) *Context {
	db, err := sql.Open("sqlite3", db_path)
	if err != nil {
		log.Fatal("DB connection failed with path:", db_path)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Couldn't ping the DB at:", db_path)
	}
	return &Context{db, new(codec.MsgpackHandle)}
}

type contextHandler struct {
	ctx *Context
	h   handler
}

type handler func(http.ResponseWriter, *http.Request, *Context)

func (ch contextHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ch.h(w, req, ch.ctx)
}

func defaultAssetPath() string {
	p, err := build.Default.Import("github.com/Blackth0rn/go-graph/public", "", build.FindOnly)
	if err != nil {
		return "."
	}
	return p.Dir
}

func homeHandler(c http.ResponseWriter, req *http.Request, ctx *Context) {
	http.ServeFile(c, req, filepath.Join(*assets, "index.html"))
}

type Link struct {
	Start_state string `codec:"start_state"`
	Action      string `codec:"action"`
	End_state   string `codec:"end_state"`
}

func wsHandler(w http.ResponseWriter, r *http.Request, ctx *Context) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// use ws to send/receive messages
	for {
		messageType, p, err := ws.ReadMessage()
		if err != nil {
			return
		}

		// decode msgpack here
		var dec *codec.Decoder = codec.NewDecoderBytes(p, ctx.mh)
		var data Link
		err = dec.Decode(&data)

		if err != nil {
			log.Println("Failed to decode data:", string(p), err)
		}
		_, err = ctx.db.Exec("INSERT INTO links VALUES (?, ?, ?)", data.Start_state, data.Action, data.End_state)
		if err != nil {
			log.Println("Failed to write data to db:", data, err)
		}

		// encode msgpack here
		var output []byte
		var enc *codec.Encoder = codec.NewEncoderBytes(&output, ctx.mh)
		err = enc.Encode(data)

		if err = ws.WriteMessage(messageType, output); err != nil {
			fmt.Println(err)
			return
		}
	}
}

var (
	addr     *string             = flag.String("addr", ":8080", "http service address")
	assets   *string             = flag.String("assets", defaultAssetPath(), "path to assets")
	upgrader *websocket.Upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func main() {
	db_path := flag.String("dbpath", "./go-graph.db", "path to db file")
	action := flag.String("action", "", "action to take: init (wipe and initialize db), clear (clear db)")
	flag.Parse()

	// Setup a file server at the default asset path
	fs := http.FileServer(http.Dir(defaultAssetPath()))

	c := NewContext(*db_path)
	defer c.Close()

	switch *action {
	case "init":
		//Init DB
		var err error
		_, err = c.db.Exec("DROP TABLE IF EXISTS links")
		if err != nil {
			log.Fatal("Failed to drop the links table: ", err)
		}
		_, err = c.db.Exec("CREATE TABLE links (start_layout TEXT, action TEXT, end_layout TEXT, CONSTRAINT unique_link UNIQUE (start_layout, action, end_layout))")
		if err != nil {
			log.Fatal("Failed to create the links table: ", err)
		}
	case "clear":
		//Clear DB
		var err error
		_, err = c.db.Exec("DELETE FROM links")
		if err != nil {
			log.Fatal("Failed to clear the links table: ", err)
		}
	}

	// handle all /public/ locations via the file server
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	// Returns the index.html file
	http.Handle("/", contextHandler{c, homeHandler})

	// Handles all websocket connections
	http.Handle("/ws", contextHandler{c, wsHandler})

	log.Println("Starting a server on:", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
