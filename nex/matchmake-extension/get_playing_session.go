package nex_matchmake_extension

import (
	"encoding/hex"
	"fmt"

	nex "github.com/PretendoNetwork/nex-go"
	matchmake_extension "github.com/PretendoNetwork/nex-protocols-go/matchmake-extension"
	"github.com/PretendoNetwork/splatoon-secure/globals"
)

func GetPlayingSession(err error, client *nex.Client, callID uint32, listPID []uint32) {
	rmcResponseStream := nex.NewStreamOut(globals.NEXServer)
	rmcResponseStream.WriteUInt32LE(0)

	rmcResponseBody := rmcResponseStream.Bytes()
	fmt.Println(hex.EncodeToString(rmcResponseBody))

	rmcResponse := nex.NewRMCResponse(matchmake_extension.ProtocolID, callID)
	rmcResponse.SetSuccess(matchmake_extension.MethodGetPlayingSession, rmcResponseBody)

	rmcResponseBytes := rmcResponse.Bytes()

	responsePacket, _ := nex.NewPacketV1(client, nil)

	responsePacket.SetVersion(1)
	responsePacket.SetSource(0xA1)
	responsePacket.SetDestination(0xAF)
	responsePacket.SetType(nex.DataPacket)
	responsePacket.SetPayload(rmcResponseBytes)

	responsePacket.AddFlag(nex.FlagNeedsAck)
	responsePacket.AddFlag(nex.FlagReliable)

	globals.NEXServer.Send(responsePacket)
}
