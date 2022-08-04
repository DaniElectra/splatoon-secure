package main

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"math"

	nex "github.com/PretendoNetwork/nex-go"
	nexproto "github.com/PretendoNetwork/nex-protocols-go"
)

var testGid uint32

func autoMatchmakeWithParam_Postpone(err error, client *nex.Client, callID uint32, matchmakeSession *nexproto.MatchmakeSession, sourceGid uint32) {

	gameConfig := matchmakeSession.Attributes[2]
	fmt.Println(strconv.FormatUint(uint64(gameConfig), 2))
	fmt.Println("===== MATCHMAKE SESSION SEARCH =====")
	fmt.Println("REGION: " + strconv.Itoa((int)(matchmakeSession.Attributes[3])))
	//fmt.Println("REGION: " + regionList[matchmakeSession.Attributes[3]])
	fmt.Println("GAME MODE: " + strconv.Itoa((int)(matchmakeSession.GameMode)))
	//fmt.Println("GAME MODE: " + gameModes[matchmakeSession.GameMode])
	//fmt.Println("CC: " + ccList[gameConfig%0b111])
	gameConfig = gameConfig >> 12
	//fmt.Println("DLC MODE: " + dlcModes[matchmakeSession.Attributes[5]&0xF])
	//fmt.Println("ITEM MODE: " + itemModes[gameConfig%0b1111])
	gameConfig = gameConfig >> 8
	//fmt.Println("VEHICLE MODE: " + vehicleModes[gameConfig%0b11])
	gameConfig = gameConfig >> 4
	//fmt.Println("CONTROLLER MODE: " + controllerModes[gameConfig%0b11])
	//fmt.Println("HAVE GUEST PLAYER: " + strconv.FormatBool(false))

	var gid uint32

	fmt.Println(sourceGid)

	
	if((int)(matchmakeSession.GameMode) == 12){
		var team uint32
		if(matchmakeSession.Attributes[3] == 0){
			team = 1
		}else{
			team = 0
		}
		gid = findRoom(matchmakeSession.GameMode, true, team, matchmakeSession.Attributes[2], uint32(1), matchmakeSession.Attributes[5]&0xF)
	}else{
		gid = findRoom(matchmakeSession.GameMode, true, matchmakeSession.Attributes[3], matchmakeSession.Attributes[2], uint32(1), matchmakeSession.Attributes[5]&0xF)
	}
	if gid == math.MaxUint32 {
		gid = newRoom(client.PID(), matchmakeSession.GameMode, true, matchmakeSession.Attributes[3], matchmakeSession.Attributes[2], uint32(1), matchmakeSession.Attributes[5]&0xF)
	}

	if((int)(matchmakeSession.GameMode) == 12){
		rmcMessage := nex.RMCRequest{}
		rmcMessage.SetProtocolID(0xe)
		rmcMessage.SetCallID(0xffff0000+callID)
		rmcMessage.SetMethodID(0x1)
	
		hostpidString := fmt.Sprintf("%.8x",(client.PID()))
		hostpidString = hostpidString[6:8] + hostpidString[4:6] + hostpidString[2:4] + hostpidString[0:2]
		gidString := fmt.Sprintf("%.8x",(gid))
		gidString = gidString[6:8] + gidString[4:6] + gidString[2:4] + gidString[0:2]
	
		for _, pid := range getRoomPlayers(sourceGid) {
			if(pid == 0){
				continue
			}
			targetClient := nexServer.FindClientFromPID(uint32(pid))
			if targetClient != nil {
				clientPidString := fmt.Sprintf("%.8x",(pid))
				clientPidString = clientPidString[6:8] + clientPidString[4:6] + clientPidString[2:4] + clientPidString[0:2]
	
				data, _ := hex.DecodeString("0017000000"+hostpidString+"90DC0100"+gidString+clientPidString+"01000001000000")
				rmcMessage.SetParameters(data)
				rmcMessageBytes := rmcMessage.Bytes()
				messagePacket, _ := nex.NewPacketV1(targetClient, nil)
				messagePacket.SetVersion(1)
				messagePacket.SetSource(0xA1)
				messagePacket.SetDestination(0xAF)
				messagePacket.SetType(nex.DataPacket)
				messagePacket.SetPayload(rmcMessageBytes)
	
				messagePacket.AddFlag(nex.FlagNeedsAck)
				messagePacket.AddFlag(nex.FlagReliable)
	
				nexServer.Send(messagePacket)
			}else{
				fmt.Println("not found")
			}
		}
	}

	fmt.Println("GATHERING ID: " + strconv.Itoa((int)(gid)))

	addPlayerToRoom(gid, client.PID(), uint32(1))

	hostpid, gamemode, region, gconfig, dlcmode := getRoomInfo(gid)
	sessionKey := "00000000000000000000000000000000"

	matchmakeSession.Gathering.ID = gid
	matchmakeSession.Gathering.OwnerPID = hostpid
	matchmakeSession.Gathering.HostPID = hostpid
	matchmakeSession.Gathering.MinimumParticipants = 1
	matchmakeSession.SessionKey = []byte(sessionKey)
	matchmakeSession.GameMode = gamemode
	matchmakeSession.Attributes[3] = region
	matchmakeSession.Attributes[2] = gconfig
	matchmakeSession.Attributes[5] = dlcmode

	rmcResponseStream := nex.NewStreamOut(nexServer)
	/*rmcResponseStream.WriteString("MatchmakeSession")
	lengthStream := nex.NewStreamOut(nexServer)
	lengthStream.WriteStructure(matchmakeSession.Gathering)
	lengthStream.WriteStructure(matchmakeSession)
	matchmakeSessionLength := uint32(len(lengthStream.Bytes()))
	rmcResponseStream.WriteUInt32LE(matchmakeSessionLength + 4)
	rmcResponseStream.WriteUInt32LE(matchmakeSessionLength)*/
	rmcResponseStream.WriteStructure(matchmakeSession.Gathering)
	rmcResponseStream.WriteStructure(matchmakeSession)

	rmcResponseBody := rmcResponseStream.Bytes()
	fmt.Println(hex.EncodeToString(rmcResponseBody))
	hostpidString := fmt.Sprintf("%.8x",(hostpid))
	hostpidString = hostpidString[6:8] + hostpidString[4:6] + hostpidString[2:4] + hostpidString[0:2]
	clientPidString := fmt.Sprintf("%.8x",(client.PID()))
	clientPidString = clientPidString[6:8] + clientPidString[4:6] + clientPidString[2:4] + clientPidString[0:2]
	gidString := fmt.Sprintf("%.8x",(gid))
	gidString = gidString[6:8] + gidString[4:6] + gidString[2:4] + gidString[0:2]
	data, _ := hex.DecodeString("0023000000"+gidString+hostpidString+hostpidString+"000008005f00000000000000000a000000000000010000035c01000001000000060000008108020107000000020000000100000010000000000000000101000000d4000000088100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ea801c8b0000000000010100410000000010011c010000006420000000161466a08c8df18b118ed5a67650a47435f081d09804a7c1902b145e18eff47c00000000001c000000020000000400405352000301050040474952000103000000000000008f7e9e961f000000010000000000000000")

	rmcResponse := nex.NewRMCResponse(nexproto.MatchmakeExtensionProtocolID, callID)
	rmcResponse.SetSuccess(nexproto.MatchmakeExtensionMethodAutoMatchmakeWithParam_Postpone, data)

	rmcResponseBytes := rmcResponse.Bytes()

	responsePacket, _ := nex.NewPacketV1(client, nil)

	responsePacket.SetVersion(1)
	responsePacket.SetSource(0xA1)
	responsePacket.SetDestination(0xAF)
	responsePacket.SetType(nex.DataPacket)
	responsePacket.SetPayload(rmcResponseBytes)

	responsePacket.AddFlag(nex.FlagNeedsAck)
	responsePacket.AddFlag(nex.FlagReliable)

	nexServer.Send(responsePacket)
	
	rmcMessage := nex.RMCRequest{}
	rmcMessage.SetProtocolID(0xe)
	rmcMessage.SetCallID(0xffff0000+callID)
	rmcMessage.SetMethodID(0x1)
	if(matchmakeSession.GameMode == 12){
		//gidString := fmt.Sprintf("%.8x",(testGid))
		//gidString = gidString[6:8] + gidString[4:6] + gidString[2:4] + gidString[0:2]
		data, _ = hex.DecodeString("0017000000"+hostpidString+"B90B0000"+gidString+clientPidString+"01000004000000")
	}else{
		data, _ = hex.DecodeString("0017000000"+hostpidString+"B90B0000"+gidString+clientPidString+"01000001000000")
		matchmakeSession.GameMode = 2 
	}
	fmt.Println(hex.EncodeToString(data))
	rmcMessage.SetParameters(data)
	rmcMessageBytes := rmcMessage.Bytes()
	
	targetClient := nexServer.FindClientFromPID(uint32(hostpid))

	messagePacket, _ := nex.NewPacketV1(targetClient, nil)
	messagePacket.SetVersion(1)
	messagePacket.SetSource(0xA1)
	messagePacket.SetDestination(0xAF)
	messagePacket.SetType(nex.DataPacket)
	messagePacket.SetPayload(rmcMessageBytes)

	messagePacket.AddFlag(nex.FlagNeedsAck)
	messagePacket.AddFlag(nex.FlagReliable)

	nexServer.Send(messagePacket)

	messagePacket, _ = nex.NewPacketV1(client, nil)
	messagePacket.SetVersion(1)
	messagePacket.SetSource(0xA1)
	messagePacket.SetDestination(0xAF)
	messagePacket.SetType(nex.DataPacket)
	messagePacket.SetPayload(rmcMessageBytes)

	messagePacket.AddFlag(nex.FlagNeedsAck)
	messagePacket.AddFlag(nex.FlagReliable)

	nexServer.Send(messagePacket)
}
