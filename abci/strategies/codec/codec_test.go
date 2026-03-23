package codec_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cmtabci "github.com/cometbft/cometbft/abci/types"

	compression "github.com/skip-mev/slinky/abci/strategies/codec"
	vetypes "github.com/skip-mev/slinky/abci/ve/types"
)

func TestDefaultVoteExtensionCodec(t *testing.T) {
	t.Run("test encoding / decoding", func(t *testing.T) {
		// create a sample vote extension
		ve := vetypes.OracleVoteExtension{
			Prices: map[uint64][]byte{
				1: []byte("1"),
				2: []byte("2"),
			},
		}
		// encode it
		codec := compression.NewDefaultVoteExtensionCodec()
		bz, err := codec.Encode(ve)
		require.NoError(t, err)

		// decode it
		decodedVe, err := codec.Decode(bz)
		require.NoError(t, err)

		// make sure it's the same
		require.Equal(t, ve.Prices, decodedVe.Prices)
	})

	t.Run("test decoding empty byte array", func(t *testing.T) {
		codec := compression.NewDefaultVoteExtensionCodec()
		_, err := codec.Decode([]byte{})
		require.Nil(t, err)
	})
}

func TestCompressionVoteExtensionCodec(t *testing.T) {
	t.Run("test encoding / decoding", func(t *testing.T) {
		// create a sample vote extension
		samplePrice := []byte("nocapongodskiptoonicewititshiiiiiiiii")
		ve := vetypes.OracleVoteExtension{
			Prices: make(map[uint64][]byte),
		}

		// add 200 prices
		for i := uint64(0); i < 200; i++ {
			ve.Prices[i] = samplePrice
		}

		// create a codec
		defaultCodec := compression.NewDefaultVoteExtensionCodec()
		codec := compression.NewCompressionVoteExtensionCodec(defaultCodec, compression.NewZLibCompressor())

		// encode it
		bz, err := codec.Encode(ve)
		require.NoError(t, err)

		defaultBz, err := defaultCodec.Encode(ve)
		require.NoError(t, err)

		// make sure it's smaller
		require.True(t, len(bz) < len(defaultBz))

		// decode it
		decodedVe, err := codec.Decode(bz)
		require.NoError(t, err)

		// make sure it's the same
		require.Equal(t, ve.Prices, decodedVe.Prices)
	})

	t.Run("test decoding empty byte array", func(t *testing.T) {
		codec := compression.NewCompressionVoteExtensionCodec(compression.NewDefaultVoteExtensionCodec(), compression.NewZLibCompressor())
		_, err := codec.Decode([]byte{})
		require.Nil(t, err)
	})
}

func TestDefaultExtendedCommitCodec(t *testing.T) {
	t.Run("test encoding / decoding", func(t *testing.T) {
		// create a sample extended commit info
		eci := cmtabci.ExtendedCommitInfo{
			Round: 1,
			Votes: []cmtabci.ExtendedVoteInfo{
				{
					Validator: cmtabci.Validator{
						Address: []byte("1"),
						Power:   10,
					},
					VoteExtension:      []byte("1"),
					ExtensionSignature: []byte("1"),
				},
			},
		}

		// encode it
		codec := compression.NewDefaultExtendedCommitCodec()
		bz, err := codec.Encode(eci)
		require.NoError(t, err)

		// decode it
		decodedEci, err := codec.Decode(bz)
		require.NoError(t, err)

		// make sure it's the same
		require.Equal(t, eci, decodedEci)
	})

	t.Run("test decoding empty byte array", func(t *testing.T) {
		codec := compression.NewDefaultExtendedCommitCodec()
		_, err := codec.Decode([]byte{})
		require.Nil(t, err)
	})
}

func TestCompressionExtendedCommitCodec(t *testing.T) {
	t.Run("test encoding / decoding", func(t *testing.T) {
		// create a sample extended commit info
		eci := cmtabci.ExtendedCommitInfo{
			Round: 1,
			Votes: []cmtabci.ExtendedVoteInfo{
				{
					Validator: cmtabci.Validator{
						Address: []byte("1"),
						Power:   10,
					},
					VoteExtension:      []byte("1"),
					ExtensionSignature: []byte("1"),
				},
			},
		}

		// create a codec
		defaultCodec := compression.NewDefaultExtendedCommitCodec()
		codec := compression.NewCompressionExtendedCommitCodec(defaultCodec, compression.NewZStdCompressor())

		// encode it
		bz, err := codec.Encode(eci)
		require.NoError(t, err)

		// decode it
		decodedEci, err := codec.Decode(bz)
		require.NoError(t, err)

		// make sure it's the same
		require.Equal(t, eci, decodedEci)
	})

	t.Run("test decoding empty byte array", func(t *testing.T) {
		codec := compression.NewCompressionExtendedCommitCodec(compression.NewDefaultExtendedCommitCodec(), compression.NewZStdCompressor())
		_, err := codec.Decode([]byte{})
		require.NoError(t, err)
	})
}

func TestZLibCompressor_CompressDecompress_UnderLimit_NoError(t *testing.T) {
	origLimit := compression.VoteExtensionSizeLimit
	t.Cleanup(func() { compression.VoteExtensionSizeLimit = origLimit })

	comp := compression.NewZLibCompressor()

	// Ensure payload size is strictly below the current limit.
	payloadLen := origLimit / 2

	payload := make([]byte, int(payloadLen))
	compressed, err := comp.Compress(payload)
	require.NoError(t, err)

	// Decompress should not be truncated because limit is unchanged.
	out, err := comp.Decompress(compressed)
	require.NoError(t, err)
	require.Equal(t, payload, out)
}

func TestZLibCompressor_Compress_OverLimit_Error(t *testing.T) {
	origLimit := compression.VoteExtensionSizeLimit
	t.Cleanup(func() { compression.VoteExtensionSizeLimit = origLimit })

	comp := compression.NewZLibCompressor()

	payloadLen := origLimit + 1

	payload := make([]byte, int(payloadLen))
	_, err := comp.Compress(payload)
	require.Error(t, err)
	require.Equal(t, "zlib compression limit reached", err.Error())
}

func TestZLibCompressor_Decompress_OverLimit_Error(t *testing.T) {
	// This test checks the expected behavior when the decompressed output would exceed the limit.
	// Decompress may return an "unexpected EOF" error; per requirements it's treated as a valid outcome.
	origLimit := compression.VoteExtensionSizeLimit
	t.Cleanup(func() { compression.VoteExtensionSizeLimit = origLimit })

	comp := compression.NewZLibCompressor()

	// Keep limits relative: compress under a higher limit, decompress under a smaller limit.
	decompressLimit := origLimit / 8

	compressLimit := origLimit

	payloadLen := decompressLimit * 2

	// Create compressed data that expands to more than decompressLimit bytes.
	compression.VoteExtensionSizeLimit = compressLimit
	payload := make([]byte, int(payloadLen))
	compressed, err := comp.Compress(payload)
	require.NoError(t, err)

	// Lower the decompression limit to trigger the error/short read path.
	compression.VoteExtensionSizeLimit = decompressLimit
	_, err = comp.Decompress(compressed)
	require.Error(t, err)
	require.ErrorIs(t, err, compression.ErrZLibDecompressionLimit)
}
