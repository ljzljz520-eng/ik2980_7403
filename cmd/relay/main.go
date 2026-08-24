package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"emergencycomms/internal/codec"
	"emergencycomms/internal/model"
	"emergencycomms/internal/report"
	"emergencycomms/internal/service"
	"emergencycomms/internal/store"
)

func main() {
	dbPath := flag.String("db", "relay.db", "path to the embedded relay database")
	command := flag.String("command", "demo", "demo, register, send, receive, audit")
	channelID := flag.String("channel", "ALPHA-01", "authenticated channel number")
	batchID := flag.String("batch", "BATCH-001", "sending batch number")
	nonce := flag.String("nonce", "N-0001", "deterministic message nonce")
	body := flag.String("body", "status-green", "short report body")
	flag.Parse()

	db, err := store.Open(*dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()
	app := service.New(db)
	var output string
	switch *command {
	case "demo":
		output, err = app.Demo(*channelID, *batchID, *nonce, *body)
	case "register":
		err = app.RegisterChannel(model.ChannelRecord{Number: *channelID, Name: "Field Alpha", Secret: "ALPHA-SECRET", Active: true})
		output = "channel registered"
	case "send":
		message := model.ProtectedMessage{ChannelNumber: *channelID, BatchNumber: *batchID, Nonce: *nonce, Body: *body}
		encoded, encodeErr := codec.Encode(message, "ALPHA-SECRET")
		err = encodeErr
		if err == nil {
			err = app.AcceptOutgoing(encoded)
		}
		output = "message queued"
	case "receive":
		result, receiveErr := app.Receive(codec.WireMessage{ChannelNumber: *channelID, BatchNumber: *batchID, Nonce: *nonce, Body: *body, Tag: codec.TagFor(*channelID, *batchID, *nonce, *body, "ALPHA-SECRET")})
		err = receiveErr
		output = report.RenderReceive(result)
	case "audit":
		entries, listErr := app.Audit(*channelID)
		err = listErr
		if listErr == nil {
			bytes, marshalErr := json.MarshalIndent(entries, "", "  ")
			err = marshalErr
			output = string(bytes)
		}
	default:
		err = fmt.Errorf("unknown command %q", *command)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(output)
}
