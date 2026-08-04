package v765_test

import (
	"reflect"
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec"
	v765 "github.com/lrnxzz/go-craft/codec/v765"
)

func TestJoinGamePreservesAllFields(t *testing.T) {
	original := &v765.JoinGame{
		EntityID:            42,
		Hardcore:            false,
		Worlds:              codec.Slice[gocraft.Identifier]{"minecraft:overworld", "minecraft:the_nether"},
		MaxPlayers:          20,
		ViewDistance:        10,
		SimulationDistance:  10,
		EnableRespawnScreen: true,
		DimensionType:       "minecraft:overworld",
		DimensionName:       "minecraft:overworld",
		HashedSeed:          -1234567890,
		GameMode:            0,
		PreviousGameMode:    -1,
		Flat:                true,
		Death: codec.Some(v765.DeathLocation{
			DimensionName: "minecraft:the_nether",
			Location: gocraft.Position{
				X: 10,
				Y: 64,
				Z: -20,
			},
		}),
		PortalCooldown: 0,
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StatePlay, codec.Clientbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("join game not registered")
	}

	if got := decoded.(*v765.JoinGame); !reflect.DeepEqual(got, original) {
		t.Errorf("got %+v, want %+v", got, original)
	}
}

func TestSyncPlayerPositionCarriesTeleportID(t *testing.T) {
	original := &v765.SyncPlayerPosition{
		X:          128.5,
		Y:          64.0,
		Z:          -256.25,
		Yaw:        90,
		Pitch:      -45,
		Flags:      0x1F,
		TeleportID: 7,
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StatePlay, codec.Clientbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("sync player position not registered")
	}

	got := decoded.(*v765.SyncPlayerPosition)
	if *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}
	if got.TeleportID != 7 {
		t.Errorf("teleport id = %d, want 7 (the bot must echo this in ConfirmTeleport)", got.TeleportID)
	}
}

func TestPlayKeepAliveDistinctIDsPerDirection(t *testing.T) {
	var (
		clientbound v765.PlayKeepAlive
		serverbound v765.PlayKeepAliveResponse
	)

	if clientbound.ID() == serverbound.ID() {
		t.Fatal("clientbound and serverbound play keep-alive must have distinct ids")
	}

	proto := v765.Protocol()

	received := &v765.PlayKeepAlive{
		KeepAliveID: 555,
	}
	echoed, ok, err := proto.Decode(codec.StatePlay, codec.Clientbound, codec.EncodeFrame(received))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("play keep-alive not registered")
	}
	if got := echoed.(*v765.PlayKeepAlive); got.KeepAliveID != 555 {
		t.Errorf("keep alive id = %d, want 555", got.KeepAliveID)
	}

	reply := &v765.PlayKeepAliveResponse{
		KeepAliveID: 555,
	}
	confirmed, ok, err := proto.Decode(codec.StatePlay, codec.Serverbound, codec.EncodeFrame(reply))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("play keep-alive response not registered")
	}
	if got := confirmed.(*v765.PlayKeepAliveResponse); got.KeepAliveID != 555 {
		t.Errorf("keep alive reply id = %d, want 555", got.KeepAliveID)
	}
}

func TestConfirmTeleportCarriesTeleportID(t *testing.T) {
	original := &v765.ConfirmTeleport{
		TeleportID: 7,
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StatePlay, codec.Serverbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("confirm teleport not registered")
	}

	if got := decoded.(*v765.ConfirmTeleport); *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}
}

func TestPlayerAbilitiesAppliesFlags(t *testing.T) {
	original := &v765.PlayerAbilities{
		Flags:       0x0D,
		FlyingSpeed: 0.05,
		FieldOfView: 0.1,
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StatePlay, codec.Clientbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("player abilities not registered")
	}

	got := decoded.(*v765.PlayerAbilities)
	if *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}

	player := &gocraft.Player{}
	got.Apply(player)

	if !player.Abilities.Invulnerable || !player.Abilities.AllowFlight || !player.Abilities.InstantBuild {
		t.Errorf("abilities = %+v, want invulnerable, allow flight and instant build", player.Abilities)
	}
	if player.Abilities.Flying {
		t.Errorf("abilities = %+v, want not flying", player.Abilities)
	}
	if player.Abilities.FlySpeed != 0.05 {
		t.Errorf("fly speed = %v, want 0.05", player.Abilities.FlySpeed)
	}
}

func TestSetExperienceAppliesProgress(t *testing.T) {
	original := &v765.SetExperience{
		Bar:             0.5,
		Level:           30,
		TotalExperience: 825,
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StatePlay, codec.Clientbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("set experience not registered")
	}

	got := decoded.(*v765.SetExperience)
	if *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}

	player := &gocraft.Player{}
	got.Apply(player)

	if player.Experience != 0.5 || player.Level != 30 || player.TotalExperience != 825 {
		t.Errorf("player xp = (%v, %v, %v), want (0.5, 30, 825)", player.Experience, player.Level, player.TotalExperience)
	}
}
