package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/ipfs-force-community/venus-tool/client"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/ipfs/go-cid"
	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/venus-messager/cli/tablewriter"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	msgTypes "github.com/filecoin-project/venus/venus-shared/types/messager"
)

var FlagServer = &cli.StringFlag{
	Name:  "server-addr",
	Usage: "Specify the server address to connect when using cli",
	Value: "127.0.0.1:12580",
}

func getAPI(ctx *cli.Context) (*client.Client, error) {

	serverAddr := "http://localhost:12580"
	if ctx.IsSet(FlagServer.Name) {
		serverAddr = "http://" + ctx.String(FlagServer.Name)
	}

	cli, err := client.New(serverAddr)
	if err != nil {
		return nil, err
	}

	cli.SetVersion("/api/v0")
	return cli, nil
}

func outputWithJson(msgs []*service.MsgResp) error {
	return printJSON(msgs)
}

func outputWithTable(msgs []*service.MsgResp, verbose bool) error {
	var tw = tablewriter.New(
		tablewriter.Col("ID"),
		tablewriter.Col("To"),
		tablewriter.Col("From"),
		tablewriter.Col("Nonce"),
		tablewriter.Col("Value"),
		tablewriter.Col("GasLimit"),
		tablewriter.Col("GasFeeCap"),
		tablewriter.Col("GasPremium"),
		tablewriter.Col("Method"),
		tablewriter.Col("State"),
		tablewriter.Col("ExitCode"),
		tablewriter.Col("CreateAt"),
	)

	for _, msgT := range msgs {
		msg := transformMessage(msgT)
		val := venusTypes.MustParseFIL(msg.Msg.Value.String() + "attofil").String()
		row := map[string]interface{}{
			"ID":         msg.ID,
			"To":         msg.Msg.To,
			"From":       msg.Msg.From,
			"Nonce":      msg.Msg.Nonce,
			"Value":      val,
			"GasLimit":   msg.Msg.GasLimit,
			"GasFeeCap":  msg.Msg.GasFeeCap,
			"GasPremium": msg.Msg.GasPremium,
			"Method":     msg.Msg.Method,
			"State":      msg.State,
			"ErrorMsg":   msgT.ErrorMsg,
			"CreateAt":   msg.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if !verbose {
			if from := msg.Msg.From.String(); len(from) > 9 {
				row["From"] = from[:9] + "..."
			}
			if to := msg.Msg.To.String(); len(to) > 9 {
				row["To"] = to[:9] + "..."
			}
			if len(msg.ID) > 36 {
				row["ID"] = msg.ID[:36] + "..."
			}
			if len(val) > 6 {
				row["Value"] = val[:6] + "..."
			}
		}
		if msg.Receipt != nil {
			row["ExitCode"] = msg.Receipt.ExitCode
		}
		tw.Write(row)
	}

	buf := new(bytes.Buffer)
	if err := tw.Flush(buf); err != nil {
		return err
	}
	fmt.Println(buf)
	return nil
}

type msgTmp struct {
	Version    uint64
	To         address.Address
	From       address.Address
	Nonce      uint64
	Value      abi.TokenAmount
	GasLimit   int64
	GasFeeCap  abi.TokenAmount
	GasPremium abi.TokenAmount
	Method     string
	Params     []byte
}

type receipt struct {
	ExitCode exitcode.ExitCode
	Return   string
	GasUsed  int64
}

type message struct {
	ID string

	UnsignedCid *cid.Cid
	SignedCid   *cid.Cid
	Msg         msgTmp
	Signature   *crypto.Signature

	Height     int64
	Confidence int64
	Receipt    *receipt
	TipSetKey  venusTypes.TipSetKey

	Meta *msgTypes.SendSpec

	WalletName string

	State string

	UpdatedAt time.Time
	CreatedAt time.Time
}

func transformMessage(msg *service.MsgResp) *message {
	if msg == nil {
		return nil
	}

	m := &message{
		ID:          msg.ID,
		UnsignedCid: msg.UnsignedCid,
		SignedCid:   msg.SignedCid,
		Signature:   msg.Signature,
		Height:      msg.Height,
		Confidence:  msg.Confidence,
		TipSetKey:   msg.TipSetKey,
		Meta:        msg.Meta,
		WalletName:  msg.WalletName,
		State:       msg.State.String(),

		UpdatedAt: msg.UpdatedAt,
		CreatedAt: msg.CreatedAt,
	}
	if msg.Receipt != nil {
		m.Receipt = &receipt{
			ExitCode: msg.Receipt.ExitCode,
			Return:   string(msg.Receipt.Return),
			GasUsed:  msg.Receipt.GasUsed,
		}
	}

	m.Msg = msgTmp{
		Version:    msg.Version,
		To:         msg.To,
		From:       msg.From,
		Nonce:      msg.Nonce,
		Value:      msg.Value,
		GasLimit:   msg.GasLimit,
		GasFeeCap:  msg.GasFeeCap,
		GasPremium: msg.GasPremium,
		Method:     fmt.Sprint(msg.Message.Method),
		Params:     msg.Params,
	}
	if msg.MethodName != "" {
		m.Msg.Method = msg.MethodName
	}

	return m
}

func printJSON(v interface{}) error {
	bytes, err := json.MarshalIndent(v, " ", "\t")
	if err != nil {
		return err
	}
	fmt.Println(string(bytes))
	return nil
}
