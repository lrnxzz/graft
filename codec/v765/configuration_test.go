package v765_test

import (
	"reflect"
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec"
	v765 "github.com/lrnxzz/go-craft/codec/v765"
	"github.com/lrnxzz/go-craft/nbt"
)

func TestClientInformationPreservesSettings(t *testing.T) {
	original := &v765.ClientInformation{
		Locale:              "en_us",
		ViewDistance:        12,
		ChatMode:            0,
		ChatColors:          true,
		DisplayedSkinParts:  0x7F,
		MainHand:            1,
		EnableTextFiltering: false,
		EnableServerListing: true,
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StateConfiguration, codec.Serverbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("client information not registered")
	}

	if got := decoded.(*v765.ClientInformation); *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}
}

func TestConfigKeepAliveDecodesInBothDirections(t *testing.T) {
	proto := v765.Protocol()

	challenge := &v765.ConfigKeepAlive{
		KeepAliveID: 1234567890,
	}
	decoded, ok, err := proto.Decode(codec.StateConfiguration, codec.Clientbound, codec.EncodeFrame(challenge))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("config keep-alive not registered clientbound")
	}
	if got := decoded.(*v765.ConfigKeepAlive); *got != *challenge {
		t.Errorf("clientbound: got %+v, want %+v", got, challenge)
	}

	response := &v765.ConfigKeepAliveResponse{
		KeepAliveID: 1234567890,
	}
	decoded, ok, err = proto.Decode(codec.StateConfiguration, codec.Serverbound, codec.EncodeFrame(response))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("config keep-alive response not registered serverbound")
	}
	if got := decoded.(*v765.ConfigKeepAliveResponse); *got != *response {
		t.Errorf("serverbound: got %+v, want %+v", got, response)
	}
}

func TestConfigDisconnectCarriesNBTReason(t *testing.T) {
	original := &v765.ConfigDisconnect{
		Reason: codec.NBT{
			"text": nbt.String("kicked"),
		},
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StateConfiguration, codec.Clientbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("config disconnect not registered")
	}

	if got := decoded.(*v765.ConfigDisconnect); !reflect.DeepEqual(got.Reason, original.Reason) {
		t.Errorf("reason = %#v, want %#v", got.Reason, original.Reason)
	}
}

func TestRegistryDataCarriesNBTCodec(t *testing.T) {
	original := &v765.RegistryData{
		Codec: codec.NBT{
			"minecraft:dimension_type": nbt.Compound{
				"type": nbt.String("minecraft:dimension_type"),
			},
		},
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StateConfiguration, codec.Clientbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("registry data not registered")
	}

	if got := decoded.(*v765.RegistryData); !reflect.DeepEqual(got.Codec, original.Codec) {
		t.Errorf("codec = %#v, want %#v", got.Codec, original.Codec)
	}
}

func TestFeatureFlagsCarriesFeatureList(t *testing.T) {
	original := &v765.FeatureFlags{
		Features: codec.Slice[gocraft.Identifier]{"minecraft:vanilla", "minecraft:bundle"},
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StateConfiguration, codec.Clientbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("feature flags not registered")
	}

	if got := decoded.(*v765.FeatureFlags); !reflect.DeepEqual(got, original) {
		t.Errorf("got %+v, want %+v", got, original)
	}
}
