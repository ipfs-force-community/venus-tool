package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/docker/go-units"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/ipfs-force-community/venus-tool/client"
	"github.com/ipfs-force-community/venus-tool/route"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/ipfs/go-cid"
	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/venus-messager/cli/tablewriter"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/types/market"
	msgTypes "github.com/filecoin-project/venus/venus-shared/types/messager"
)

var StringToStorageState = map[string]storagemarket.StorageDealStatus{}

var DefaultServerAddr = "http://localhost:12580"

func init() {
	for state, stateStr := range storagemarket.DealStates {
		StringToStorageState[stateStr] = state
	}
}

var FlagServer = &cli.StringFlag{
	Name:  "server-addr",
	Usage: "Specify the server address to connect when using cli",
	Value: "127.0.0.1:12580",
}

func getAPI(ctx *cli.Context) (service.IService, error) {
	ret := &service.IServiceStruct{}

	serverAddr := DefaultServerAddr
	if ctx.IsSet(FlagServer.Name) {
		serverAddr = "http://" + ctx.String(FlagServer.Name)
	}

	cli, err := client.New(serverAddr)
	if err != nil {
		return nil, err
	}

	cli.SetVersion("/api/v0")

	route.Provide(cli, &ret.Internal)
	return ret, nil
}

func outputWithJson(msgs []*service.MsgResp) error {
	return printJSON(msgs)
}

func outputMsgWithTable(msgs []*service.MsgResp, verbose bool) error {
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
		val := types.MustParseFIL(msg.Msg.Value.String() + "attofil").String()
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
	TipSetKey  types.TipSetKey

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

func outputStorageDeals(out io.Writer, deals []market.MinerDeal, verbose bool) error {
	sort.Slice(deals, func(i, j int) bool {
		return deals[i].CreationTime.Time().Before(deals[j].CreationTime.Time())
	})

	w := tabwriter.NewWriter(out, 2, 4, 2, ' ', 0)

	if verbose {
		_, _ = fmt.Fprintf(w, "Creation\tVerified\tProposalCid\tDealId\tState\tPieceState\tClient\tProvider\tSize\tPrice\tDuration\tTransferChannelID\tAddFundCid\tPublishCid\tMessage\n")
	} else {
		_, _ = fmt.Fprintf(w, "ProposalCid\tDealId\tState\tPieceState\tClient\tProvider\tSize\tPrice\tDuration\n")
	}

	for _, deal := range deals {
		propcid := deal.ProposalCid.String()
		if !verbose {
			propcid = "..." + propcid[len(propcid)-8:]
		}

		fil := types.FIL(types.BigMul(deal.Proposal.StoragePricePerEpoch, types.NewInt(uint64(deal.Proposal.Duration()))))

		if verbose {
			_, _ = fmt.Fprintf(w, "%s\t%t\t", deal.CreationTime.Time().Format(time.Stamp), deal.Proposal.VerifiedDeal)
		}

		_, _ = fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s", propcid, deal.DealID, storagemarket.DealStates[deal.State], deal.PieceStatus,
			deal.Proposal.Client, deal.Proposal.Provider, units.BytesSize(float64(deal.Proposal.PieceSize)), fil, deal.Proposal.Duration())
		if verbose {
			tchid := ""
			if deal.TransferChannelID != nil {
				tchid = deal.TransferChannelID.String()
			}

			addFundcid := ""
			if deal.AddFundsCid != nil {
				addFundcid = deal.AddFundsCid.String()
			}

			pubcid := ""
			if deal.PublishCid != nil {
				pubcid = deal.PublishCid.String()
			}

			_, _ = fmt.Fprintf(w, "\t%s", tchid)
			_, _ = fmt.Fprintf(w, "\t%s", addFundcid)
			_, _ = fmt.Fprintf(w, "\t%s", pubcid)
			_, _ = fmt.Fprintf(w, "\t%s", deal.Message)
		}

		_, _ = fmt.Fprintln(w)
	}

	return w.Flush()
}
