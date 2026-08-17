package v765

import (
	"github.com/lrnxzz/graft/codec"
)

const ProtocolVersion = 765

func Protocol() *codec.Protocol {
	proto := codec.NewProtocol()

	codec.Bind[Handshake](proto)

	codec.Bind[LoginStart](proto)
	codec.Bind[LoginAcknowledged](proto)
	codec.Bind[LoginDisconnect](proto)
	codec.Bind[EncryptionBegin](proto)
	codec.Bind[LoginSuccess](proto)
	codec.Bind[SetCompression](proto)

	codec.Bind[ClientInformation](proto)
	codec.Bind[AcknowledgeConfiguration](proto)
	codec.Bind[ConfigKeepAliveResponse](proto)
	codec.Bind[ConfigPong](proto)
	codec.Bind[ConfigDisconnect](proto)
	codec.Bind[FinishConfiguration](proto)
	codec.Bind[ConfigKeepAlive](proto)
	codec.Bind[ConfigPing](proto)
	codec.Bind[RegistryData](proto)
	codec.Bind[FeatureFlags](proto)

	codec.Bind[ConfirmTeleport](proto)
	codec.Bind[PlayKeepAliveResponse](proto)
	codec.Bind[SetPlayerPosition](proto)
	codec.Bind[SetPlayerPositionRotation](proto)
	codec.Bind[HeldItemChange](proto)
	codec.Bind[ClickContainer](proto)
	codec.Bind[CloseContainer](proto)
	codec.Bind[PlayerAction](proto)
	codec.Bind[UseItemOn](proto)
	codec.Bind[ChunkBatchReceived](proto)
	codec.Bind[ChatMessage](proto)
	codec.Bind[ChatCommand](proto)
	codec.Bind[SystemChat](proto)
	codec.Bind[PlayerChat](proto)
	codec.Bind[DisguisedChat](proto)
	codec.Bind[JoinGame](proto)
	codec.Bind[ChunkBatchStart](proto)
	codec.Bind[ChunkBatchFinished](proto)
	codec.Bind[PlayKeepAlive](proto)
	codec.Bind[SyncPlayerPosition](proto)
	codec.Bind[ChunkData](proto)
	codec.Bind[UnloadChunk](proto)
	codec.Bind[BlockUpdate](proto)
	codec.Bind[SectionBlocksUpdate](proto)
	codec.Bind[SetHealth](proto)
	codec.Bind[PlayerAbilities](proto)
	codec.Bind[SetExperience](proto)
	codec.Bind[SetContainerContent](proto)
	codec.Bind[SetContainerSlot](proto)
	codec.Bind[SetHeldItem](proto)
	codec.Bind[AcknowledgeBlockChange](proto)
	codec.Bind[PlayDisconnect](proto)

	return proto
}
