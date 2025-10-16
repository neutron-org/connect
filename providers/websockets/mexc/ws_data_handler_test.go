package mexc_test

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"

	providertypes "github.com/skip-mev/slinky/providers/types"
	mexcpb "github.com/skip-mev/slinky/providers/websockets/mexc/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	goproto "google.golang.org/protobuf/proto"

	"github.com/skip-mev/slinky/oracle/config"
	"github.com/skip-mev/slinky/oracle/types"
	"github.com/skip-mev/slinky/providers/base/websocket/handlers"
	"github.com/skip-mev/slinky/providers/websockets/mexc"
)

var (
	btcusdt = types.DefaultProviderTicker{
		OffChainTicker: "BTCUSDT",
	}
	ethusdt = types.DefaultProviderTicker{
		OffChainTicker: "ETHUSDT",
	}
	atomusdc = types.DefaultProviderTicker{
		OffChainTicker: "ATOMUSDC",
	}
	logger = zap.NewExample()
)

func TestTest(t *testing.T) {
	testCases := []struct {
		name string
		msg  func() string
	}{
		{
			name: "unknown message - old format",
			msg: func() string {
				return `{"id":0,"code":0,"msg":"UNKNOWN"}`
			},
		},
		{
			name: "unsupported market price update",
			msg: func() string {
				msg := mexcpb.PushDataV3ApiWrapper{
					Channel: "spot@public.miniTicker.v3.api.pb",
					Body: &mexcpb.PushDataV3ApiWrapper_PublicMiniTicker{
						PublicMiniTicker: &mexcpb.PublicMiniTickerV3Api{
							Symbol:             "MEMCOIN",
							Price:              "10.00",
							Rate:               "",
							ZonedRate:          "",
							High:               "",
							Low:                "",
							Volume:             "",
							Quantity:           "",
							LastCloseRate:      "",
							LastCloseZonedRate: "",
							LastCloseHigh:      "",
							LastCloseLow:       "",
						},
					},
				}

				bz, err := goproto.Marshal(&msg)

				if err != nil {
					panic("kekw")
				}

				str := base64.StdEncoding.EncodeToString(bz)

				return str
			},
		},
		{
			name: "price update from incorrect channel",
			msg: func() string {
				msg := mexcpb.PushDataV3ApiWrapper{
					Channel: "futures@public.miniTicker.v3.api.pb",
					Body: &mexcpb.PushDataV3ApiWrapper_PublicMiniTicker{
						PublicMiniTicker: &mexcpb.PublicMiniTickerV3Api{
							Symbol:             "BTCUSDT",
							Price:              "10000.00",
							Rate:               "",
							ZonedRate:          "",
							High:               "",
							Low:                "",
							Volume:             "",
							Quantity:           "",
							LastCloseRate:      "",
							LastCloseZonedRate: "",
							LastCloseHigh:      "",
							LastCloseLow:       "",
						},
					},
					Symbol:     nil,
					SymbolId:   nil,
					CreateTime: nil,
					SendTime:   nil,
				}

				bz, err := goproto.Marshal(&msg)

				if err != nil {
					panic("kekw")
				}

				str := base64.StdEncoding.EncodeToString(bz)

				return str
			},
		},
		{
			name: "price update with invalid price",
			msg: func() string {
				msg := mexcpb.PushDataV3ApiWrapper{
					Channel: "spot@public.miniTicker.v3.api.pb",
					Body: &mexcpb.PushDataV3ApiWrapper_PublicMiniTicker{
						PublicMiniTicker: &mexcpb.PublicMiniTickerV3Api{
							Symbol:             "BTCUSDT",
							Price:              "$10,000.00",
							Rate:               "",
							ZonedRate:          "",
							High:               "",
							Low:                "",
							Volume:             "",
							Quantity:           "",
							LastCloseRate:      "",
							LastCloseZonedRate: "",
							LastCloseHigh:      "",
							LastCloseLow:       "",
						},
					},
					Symbol:     nil,
					SymbolId:   nil,
					CreateTime: nil,
					SendTime:   nil,
				}

				bz, err := goproto.Marshal(&msg)

				if err != nil {
					panic("kekw")
				}

				str := base64.StdEncoding.EncodeToString(bz)

				return str
			},
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\nname: %s Message: %s\n", tc.name, tc.msg())
	}
}

func TestHandleMessage(t *testing.T) {
	testCases := []struct {
		name          string
		msg           func() []byte
		resp          types.PriceResponse
		updateMessage func() []handlers.WebsocketEncodedMessage
		expErr        bool
	}{
		{
			name: "pong message",
			msg: func() []byte {
				return []byte(`{"id":0,"code":0,"msg":"PONG"}`)
			},
			resp: types.PriceResponse{},
			updateMessage: func() []handlers.WebsocketEncodedMessage {
				return nil
			},
			expErr: false,
		},
		{
			name: "subscription message",
			msg: func() []byte {
				return []byte(`{"id":0,"code":0,"msg":"spot@public.miniTicker.v3.api.pb@BTCUSDT@UTC+8"}`)
			},
			resp: types.PriceResponse{},
			updateMessage: func() []handlers.WebsocketEncodedMessage {
				return nil
			},
			expErr: false,
		},
		{
			name: "unknown message",
			msg: func() []byte {
				return []byte(`{"id":0,"code":0,"msg":"UNKNOWN"}`)
			},
			resp: types.PriceResponse{},
			updateMessage: func() []handlers.WebsocketEncodedMessage {
				return nil
			},
			expErr: true,
		},
		{
			name: "unsupported market price update",
			msg: func() []byte {
				msg := "CiBzcG90QHB1YmxpYy5taW5pVGlja2VyLnYzLmFwaS5wYqoTEAoHTUVNQ09JThIFMTAuMDA="
				decoded, err := base64.StdEncoding.DecodeString(msg)

				if err != nil {
					panic(err)
				}
				return decoded
			},
			resp: types.PriceResponse{},
			updateMessage: func() []handlers.WebsocketEncodedMessage {
				return nil
			},
			expErr: true,
		},
		{
			name: "price update from incorrect channel",
			msg: func() []byte {
				msg := "CjFmdXR1cmVzQHB1YmxpYy5taW5pVGlja2VyLnYzLmFwaS5wYkBCVENVU0RUQFVUQys4GgdCVENVU0RUMLulvZqUM6oTfAoHQlRDVVNEVBIJMTE1ODg3LjUxGgYwLjAwNjciBjAuMDA2NyoJMTE2NjU1LjM5MgkxMTQ4NzIuODM6Czg1Njc3MjA1Ny41Qg03MzkwLjYxNzM4MDM4SgYwLjAwNjdSBjAuMDA2N1oJMTE2NjU1LjM5YgkxMTQ4NzIuODM="
				decoded, err := base64.StdEncoding.DecodeString(msg)

				if err != nil {
					panic(err)
				}
				return decoded
			},
			resp: types.PriceResponse{
				UnResolved: types.UnResolvedPrices{
					btcusdt: providertypes.UnresolvedResult{
						ErrorWithCode: providertypes.NewErrorWithCode(fmt.Errorf("invalid channel"), providertypes.ErrorWebSocketGeneral),
					},
				},
			},
			updateMessage: func() []handlers.WebsocketEncodedMessage {
				return nil
			},
			expErr: true,
		},
		{
			name: "price update with invalid price",
			msg: func() []byte {
				msg := "CiBzcG90QHB1YmxpYy5taW5pVGlja2VyLnYzLmFwaS5wYqoTFQoHQlRDVVNEVBIKJDEwLDAwMC4wMA=="
				decoded, err := base64.StdEncoding.DecodeString(msg)

				if err != nil {
					panic(err)
				}
				return decoded
			},
			resp: types.PriceResponse{
				UnResolved: types.UnResolvedPrices{
					btcusdt: providertypes.UnresolvedResult{
						ErrorWithCode: providertypes.NewErrorWithCode(fmt.Errorf("invalid price"), providertypes.ErrorWebSocketGeneral),
					},
				},
			},
			updateMessage: func() []handlers.WebsocketEncodedMessage {
				return nil
			},
			expErr: true,
		},
		{
			name: "valid price update 1",
			msg: func() []byte {
				msg := "Ci5zcG90QHB1YmxpYy5taW5pVGlja2VyLnYzLmFwaS5wYkBCVENVU0RUQFVUQys4GgdCVENVU0RUMLulvZqUM6oTfAoHQlRDVVNEVBIJMTE1ODg3LjUxGgYwLjAwNjciBjAuMDA2NyoJMTE2NjU1LjM5MgkxMTQ4NzIuODM6Czg1Njc3MjA1Ny41Qg03MzkwLjYxNzM4MDM4SgYwLjAwNjdSBjAuMDA2N1oJMTE2NjU1LjM5YgkxMTQ4NzIuODM="
				decoded, err := base64.StdEncoding.DecodeString(msg)

				if err != nil {
					panic(err)
				}
				return decoded
			},
			resp: types.PriceResponse{
				Resolved: types.ResolvedPrices{
					btcusdt: {
						Value: big.NewFloat(115887.50),
					},
				},
			},
			updateMessage: func() []handlers.WebsocketEncodedMessage {
				return nil
			},
			expErr: false,
		},
		{
			name: "valid price update 2",
			msg: func() []byte {
				msg := "Ci5zcG90QHB1YmxpYy5taW5pVGlja2VyLnYzLmFwaS5wYkBFVEhVU0RUQFVUQys4GgdFVEhVU0RUMLelvZqUM6oTbwoHRVRIVVNEVBIHNDcxNC4wNxoFMC4wMzgiBTAuMDM4Kgc0NzY2LjU5Mgc0NTA4LjY0Og0xMzUzNzcwNjA2LjQ3QgwyOTAyNDcuNDA4MzdKBTAuMDM4UgUwLjAzOFoHNDc2Ni41OWIHNDUwOC42NA=="
				decoded, err := base64.StdEncoding.DecodeString(msg)

				if err != nil {
					panic(err)
				}
				return decoded
			},
			resp: types.PriceResponse{
				Resolved: types.ResolvedPrices{
					ethusdt: {
						Value: big.NewFloat(4714.06),
					},
				},
			},
			updateMessage: func() []handlers.WebsocketEncodedMessage {
				return nil
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wsHandler, err := mexc.NewWebSocketDataHandler(logger, mexc.DefaultWebSocketConfig)
			require.NoError(t, err)

			// Update the cache since it is assumed that CreateMessages is executed before anything else.
			_, err = wsHandler.CreateMessages([]types.ProviderTicker{btcusdt, ethusdt, atomusdc})
			require.NoError(t, err)

			resp, updateMsg, err := wsHandler.HandleMessage(tc.msg())
			if tc.expErr {
				require.Error(t, err)

				require.Equal(t, len(tc.resp.UnResolved), len(resp.UnResolved))
				for cp := range tc.resp.UnResolved {
					require.Contains(t, resp.UnResolved, cp)
					require.Error(t, resp.UnResolved[cp])
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.updateMessage(), updateMsg)

			require.Equal(t, len(tc.resp.Resolved), len(resp.Resolved))
			require.Equal(t, len(tc.resp.UnResolved), len(resp.UnResolved))

			for cp, result := range tc.resp.Resolved {
				require.Contains(t, resp.Resolved, cp)
				require.Equal(t,
					result.Value.SetPrec(18).SetMode(big.ToNearestEven),
					resp.Resolved[cp].Value.SetPrec(18).SetMode(big.ToNearestEven),
				)
			}

			for cp := range tc.resp.UnResolved {
				require.Contains(t, resp.UnResolved, cp)
				require.Error(t, resp.UnResolved[cp])
			}
		})
	}
}

func TestCreateMessages(t *testing.T) {
	batchCfg := mexc.DefaultWebSocketConfig
	batchCfg.MaxSubscriptionsPerBatch = 2

	testCases := []struct {
		name        string
		cps         []types.ProviderTicker
		cfg         config.WebSocketConfig
		expected    func() []handlers.WebsocketEncodedMessage
		expectedErr bool
	}{
		{
			name: "single currency pair",
			cps: []types.ProviderTicker{
				btcusdt,
			},
			cfg: mexc.DefaultWebSocketConfig,
			expected: func() []handlers.WebsocketEncodedMessage {
				msg := `{"method":"SUBSCRIPTION","params":["spot@public.miniTicker.v3.api.pb@BTCUSDT@UTC+8"]}`
				return []handlers.WebsocketEncodedMessage{[]byte(msg)}
			},
			expectedErr: false,
		},
		{
			name: "multiple currency pairs",
			cps: []types.ProviderTicker{
				btcusdt,
				ethusdt,
				atomusdc,
			},
			cfg: mexc.DefaultWebSocketConfig,
			expected: func() []handlers.WebsocketEncodedMessage {
				msg1 := `{"method":"SUBSCRIPTION","params":["spot@public.miniTicker.v3.api.pb@BTCUSDT@UTC+8"]}`
				msg2 := `{"method":"SUBSCRIPTION","params":["spot@public.miniTicker.v3.api.pb@ETHUSDT@UTC+8"]}`
				msg3 := `{"method":"SUBSCRIPTION","params":["spot@public.miniTicker.v3.api.pb@ATOMUSDC@UTC+8"]}`
				return []handlers.WebsocketEncodedMessage{[]byte(msg1), []byte(msg2), []byte(msg3)}
			},
			expectedErr: false,
		},
		{
			name: "multiple currency pairs with batch",
			cps: []types.ProviderTicker{
				btcusdt,
				ethusdt,
			},
			cfg: batchCfg,
			expected: func() []handlers.WebsocketEncodedMessage {
				msg1 := `{"method":"SUBSCRIPTION","params":["spot@public.miniTicker.v3.api.pb@BTCUSDT@UTC+8","spot@public.miniTicker.v3.api.pb@ETHUSDT@UTC+8"]}`
				return []handlers.WebsocketEncodedMessage{[]byte(msg1)}
			},
			expectedErr: false,
		},
		{
			name: "multiple currency pairs with batch and remainder",
			cps: []types.ProviderTicker{
				btcusdt,
				ethusdt,
				atomusdc,
			},
			cfg: batchCfg,
			expected: func() []handlers.WebsocketEncodedMessage {
				msg1 := `{"method":"SUBSCRIPTION","params":["spot@public.miniTicker.v3.api.pb@BTCUSDT@UTC+8","spot@public.miniTicker.v3.api.pb@ETHUSDT@UTC+8"]}`
				msg2 := `{"method":"SUBSCRIPTION","params":["spot@public.miniTicker.v3.api.pb@ATOMUSDC@UTC+8"]}`
				return []handlers.WebsocketEncodedMessage{[]byte(msg1), []byte(msg2)}
			},
			expectedErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wsHandler, err := mexc.NewWebSocketDataHandler(logger, tc.cfg)
			require.NoError(t, err)

			msgs, err := wsHandler.CreateMessages(tc.cps)
			if tc.expectedErr {
				require.Error(t, err)
				return
			}
			require.Equal(t, tc.expected(), msgs)
		})
	}
}

func TestHeartBeatMessages(t *testing.T) {
	wsHandler, err := mexc.NewWebSocketDataHandler(logger, mexc.DefaultWebSocketConfig)
	require.NoError(t, err)

	expected := []handlers.WebsocketEncodedMessage{
		[]byte(`{"id":0,"code":0,"msg":"PING"}`),
	}

	msgs, err := wsHandler.HeartBeatMessages()
	require.NoError(t, err)
	require.Equal(t, expected, msgs)
}
