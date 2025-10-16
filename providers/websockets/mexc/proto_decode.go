package mexc

import (
	"fmt"

	mexcpb "github.com/skip-mev/slinky/providers/websockets/mexc/proto"
	goproto "google.golang.org/protobuf/proto"
)

// Lightweight protobuf wire decoder for PublicMiniTickerV3Api
// Schema reference: https://github.com/mexcdevelop/websocket-proto/blob/main/PublicMiniTickerV3Api.proto
// We only need fields:
//  1: symbol (string)
//  2: price  (string)

// decodeMiniTickerProtobuf extracts symbol and price from a protobuf-encoded
// PublicMiniTickerV3Api message. It returns error if the payload does not
// appear to be a valid protobuf with the expected fields.
func decodeMiniTickerProtobuf(message []byte) (string, string, string, error) {
	var wrapper mexcpb.PushDataV3ApiWrapper
	if err := goproto.Unmarshal(message, &wrapper); err != nil {
		return "", "", "", fmt.Errorf("failed to unmarshal PushDataV3ApiWrapper from proto: %w", err)
	}
	// Check that field of the oneof is set to "PublicMiniTickerV3Api"
	var minitickerMsg *mexcpb.PublicMiniTickerV3Api
	switch body := wrapper.Body.(type) {
	case *mexcpb.PushDataV3ApiWrapper_PublicMiniTicker:
		minitickerMsg = body.PublicMiniTicker
	default:
		return "", "", "", fmt.Errorf("no PublicMiniTicker in this message (found %T)", body)
	}
	symbol := minitickerMsg.GetSymbol()
	price := minitickerMsg.GetPrice()
	channel := wrapper.Channel

	if symbol != "" && price != "" {
		return channel, symbol, price, nil
	} else {
		return "", "", "", fmt.Errorf("empty symbol=%s or price=%s", symbol, price)
	}
}
