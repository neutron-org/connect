package mexc

import (
	"encoding/base64"
	"fmt"

	goproto "google.golang.org/protobuf/proto"

	mexcpb "github.com/skip-mev/slinky/providers/websockets/mexc/proto"
)

// Lightweight protobuf wire decoder for PublicMiniTickerV3Api
// Schema reference: https://github.com/mexcdevelop/websocket-proto/blob/main/PublicMiniTickerV3Api.proto
// We only need fields:
//  1: symbol (string)
//  2: price  (string)

// decodeMiniTickerProtobuf extracts symbol and price from a protobuf-encoded
// PublicMiniTickerV3Api message. It returns ok=false if the payload does not
// appear to be a valid protobuf with the expected fields.
func decodeMiniTickerProtobuf(message []byte) (symbol string, price string, ok bool) {
	// MEXC may prepend an ASCII topic prefix before the protobuf bytes.
	// Scan forward and attempt to unmarshal from each offset.
	encoded := base64.StdEncoding.EncodeToString(message)
	fmt.Println("\n\nmexc encoded:\n", encoded)

	//for off := 0; off < len(message); off++ {
	fmt.Printf("message: %s\n", string(message))
	var msg mexcpb.PublicMiniTickerV3Api
	if err := goproto.Unmarshal(message[0:], &msg); err != nil {
		return "", "", false
	}
	s := msg.GetSymbol()
	p := msg.GetPrice()
	if s != "" && p != "" {
		return s, p, true
	}
	//}
	return "", "", false
}
