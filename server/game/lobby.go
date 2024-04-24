package game

import (
	"time"
)

type LobbyRoom struct {
	Mode     uint8
	Idx      uint32
	HostSidx uint32
	Guests   []LobbyGuest
}

func (r *LobbyRoom) PlayerCount() int {
	return 1 + len(r.Guests)
}

type LobbyGuest struct {
	Ready         bool
	Sidx          uint32
	InvitedBySidx uint32
}

type LobbyInvitation struct {
	InvitorSidx uint32
	InviteeSidx uint32
}

// rules:
// sidx: player which will be the host and not yet belong to a lobby
// mode: mode of the lobby
func lobbyCreate(sidx uint32, mode uint8) {}

// rules:
// sidx: player which currently inside of a lobby and requesting to leave
func lobbyLeave(sidx uint32) {}

// rules:
// sidx: a player who request to join
// lobbySidx: a player which belong to a lobby and want to be joined
func lobbyRequestJoin(sidx, lobbySidx uint32) {}

// rules:
// sidx: owner of a lobby
// otherSidx: a player which previously request to join
// accept: respond status (accept or decline the request)
func lobbyRespondJoinRequest(sidx, otherSidx uint32, accept bool) {}

// rules:
// sidx: player who owns a lobby which will be dismissed
func lobbyDismiss(sidx uint32) {}

// rules:
// sidx: owner of a lobby
// mode: mode to set
func lobbySetMode(sidx uint32, mode uint8) {}

// rules:
// sidx: owner of a lobby
// targetSidx: player inside of owner lobby which will be the host
func lobbySetHost(sidx, targetSidx uint32) {}

// rules:
// sidx: player inside of lobby
// ready: ready state of player lobby
func lobbySetReady(sidx uint32, ready bool) {}

// rules:
// sidx: owner of a lobby
// start (start: true): host
// cancel (start: false): all
func lobbySetMatchmaking(sidx uint32, start bool) {}

// rules:
// sidx: a player who issues an invite, must be inside of a lobby
// targetSidx: a player who will receive the invitation
func lobbyInvitePlayer(sidx, targetSidx uint32) {}

// rules:
// sidx: a player who respond the invitation
// invitorSidx: a player who previously issues an invitation
// accept: the respond of invitation which is accept or decline
func lobbyRespondInvitation(sidx, invitorSidx uint32, accept bool) {}

// rules:
// sidx: owner of a lobby
// targetSidx: player inside of owner lobby
func lobbyKickPlayer(sidx, targetSidx uint32) {}

type PlayerConfig struct {
	Sidx            uint32
	PemmelId        uint32
	TeamSide        uint8
	SkinId          uint16
	SpawnEffectId   uint16
	RecallEffectId  uint16
	AlertStyleId    uint16
	LoadingProgress uint8
}

type MatchConfig struct {
	Mode           uint8
	Id             uint32
	SunSideSkinId  uint16
	MoonSideSkinId uint16
	PlayerConfigs  [10]PlayerConfig
	Begin          time.Time
	End            time.Time
}

// list of standby lobby
// list of queued lobby -> chan queued -> matchmaking goroutine
// list of live match (MatchConfig)

var standby []LobbyRoom
var match []MatchConfig

// test case #1:
// target = 5
// pair = 2
// matchmaking = [1,2,3,4,3,2,2]
// output:
// [0,3] = sums up to 5
// [1,2] = sums up to 5
// matchmaking found!
//
// test case #2:
// target = 5
// pair = 1
// matchmaking = [1,3,2,1]
// output:
// [0,1,3] = sums up to 5
// matchmaking found
