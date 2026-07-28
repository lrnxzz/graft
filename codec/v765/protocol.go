package v765

import gocraft "github.com/lrnxzz/go-craft"

const ProtocolVersion = 765

func Protocol() *gocraft.Protocol {
	proto := gocraft.NewProtocol()

	gocraft.Bind[Handshake](proto)

	gocraft.Bind[LoginStart](proto)
	gocraft.Bind[LoginAcknowledged](proto)
	gocraft.Bind[LoginDisconnect](proto)
	gocraft.Bind[EncryptionBegin](proto)
	gocraft.Bind[LoginSuccess](proto)
	gocraft.Bind[SetCompression](proto)

	gocraft.Bind[ClientInformation](proto)
	gocraft.Bind[AcknowledgeConfiguration](proto)
	gocraft.Bind[ConfigKeepAliveResponse](proto)
	gocraft.Bind[ConfigPong](proto)
	gocraft.Bind[ConfigDisconnect](proto)
	gocraft.Bind[FinishConfiguration](proto)
	gocraft.Bind[ConfigKeepAlive](proto)
	gocraft.Bind[ConfigPing](proto)
	gocraft.Bind[RegistryData](proto)
	gocraft.Bind[FeatureFlags](proto)

	gocraft.Bind[ConfirmTeleport](proto)
	gocraft.Bind[PlayKeepAliveResponse](proto)
	gocraft.Bind[SetPlayerPosition](proto)
	gocraft.Bind[SetPlayerPositionRotation](proto)
	gocraft.Bind[HeldItemChange](proto)
	gocraft.Bind[ClickContainer](proto)
	gocraft.Bind[CloseContainer](proto)
	gocraft.Bind[PlayerAction](proto)
	gocraft.Bind[UseItemOn](proto)
	gocraft.Bind[ChunkBatchReceived](proto)
	gocraft.Bind[ChatMessage](proto)
	gocraft.Bind[ChatCommand](proto)
	gocraft.Bind[SystemChat](proto)
	gocraft.Bind[PlayerChat](proto)
	gocraft.Bind[DisguisedChat](proto)
	gocraft.Bind[JoinGame](proto)
	gocraft.Bind[ChunkBatchStart](proto)
	gocraft.Bind[ChunkBatchFinished](proto)
	gocraft.Bind[PlayKeepAlive](proto)
	gocraft.Bind[SyncPlayerPosition](proto)
	gocraft.Bind[ChunkData](proto)
	gocraft.Bind[UnloadChunk](proto)
	gocraft.Bind[BlockUpdate](proto)
	gocraft.Bind[SectionBlocksUpdate](proto)
	gocraft.Bind[SetHealth](proto)
	gocraft.Bind[PlayerAbilities](proto)
	gocraft.Bind[SetExperience](proto)
	gocraft.Bind[SetContainerContent](proto)
	gocraft.Bind[SetContainerSlot](proto)
	gocraft.Bind[SetHeldItem](proto)
	gocraft.Bind[AcknowledgeBlockChange](proto)
	gocraft.Bind[PlayDisconnect](proto)

	return proto
}
