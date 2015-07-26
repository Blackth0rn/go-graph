package main

import (
	"log"

	_ "github.com/mattn/go-sqlite3" //no name needed as it implements the database/sql interface
	"github.com/ugorji/go/codec"
)

type Link struct {
	Start_state string `codec:"start_state"`
	Action      string `codec:"action"`
	End_state   string `codec:"end_state"`
}

func (l *Link) Decode(p []byte, ctx *Context) error {
	var dec *codec.Decoder = codec.NewDecoderBytes(p, ctx.mh)
	err := dec.Decode(l)
	return err
}

func (l *Link) Encode(output []byte, ctx *Context) error {
	var enc *codec.Encoder = codec.NewEncoderBytes(&output, ctx.mh)
	err := enc.Encode(l)
	return err
}

type Message interface {
	Decode([]byte, *Context) error
	Encode([]byte, *Context) error
}

func decodeMessage(p []byte, m Message, ctx *Context) error {
	return m.Decode(p, ctx)
}

func encodeMessage(p []byte, m Message, ctx *Context) error {
	return m.Encode(p, ctx)
}

func addLink(payload []byte, ctx *Context) ([]byte, error) {
	var err error
	var output []byte

	m := new(Link)

	// decode msgpack here
	if err = decodeMessage(payload, m, ctx); err != nil {
		log.Println("Failed to decode data:", string(payload), err)
	}
	_, err = ctx.db.Exec("INSERT INTO links VALUES (?, ?, ?)", m.Start_state, m.Action, m.End_state)
	if err != nil {
		log.Println("Failed to write data to db:", m, err)
	}

	// encode msgpack here
	if err = encodeMessage(output, m, ctx); err != nil {
		log.Println("Failed to encode data:", m, err)
	}

	return output, err
}

func sendList(payload []byte, ctx *Context) ([]byte, error) {
	var err error
	var output []byte
	return output, err
}
