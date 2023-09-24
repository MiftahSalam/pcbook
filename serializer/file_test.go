package serializer

import (
	"pc-book/pb"
	"pc-book/sample"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestWriteProtobufToBinFile(t *testing.T) {
	t.Parallel()

	binFile := "../tmp/laptop.bin"
	jsonFile := "../tmp/laptop.json"

	laptop1 := sample.NewLaptop()
	err := WriteProtobufToBinaryFile(laptop1, binFile)
	require.NoError(t, err)

	laptop2 := &pb.Laptop{}

	err = ReadBinFileToProtobuf(binFile, laptop2)

	require.NoError(t, err)
	require.True(t, proto.Equal(laptop1, laptop2))

	err = WriteProtobufToJsonFile(laptop1, jsonFile)
	require.NoError(t, err)
}
