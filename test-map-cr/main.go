package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	hazelcast "github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-platform-operator/controllers/protocol/codec"
	"github.com/hazelcast/hazelcast-platform-operator/controllers/protocol/types"
)

var (
	mapName     = flag.String("n", "newMap", "Map Name")
	address     = flag.String("ad", "127.0.0.1:5701", "address of cluster")
	usePublicIP = flag.Bool("p", false, "use Public IP")
)

func GetMapConfig(ctx context.Context, client *hazelcast.Client, mapName string) (types.MapConfig, error) {
	client.GetMap(ctx, mapName)
	ci := hazelcast.NewClientInternal(client)
	req := codec.EncodeMCGetMapConfigRequest(
		mapName, //name
	)
	resp, err := ci.InvokeOnRandomTarget(ctx, req, nil)
	if err != nil {
		fmt.Printf("got error %+v\n", err)
		return types.MapConfig{}, err
	}
	return codec.DecodeMCGetMapConfigResponse(resp), nil
}

func UpdateMapConfig(ctx context.Context, client *hazelcast.Client, mapName string) error {
	client.GetMap(ctx, mapName)
	ci := hazelcast.NewClientInternal(client)
	req := codec.EncodeMCUpdateMapConfigRequest(
		mapName, //name
		0,
		0,
		types.EvictionPolicyTypeEncode[types.EvictionPolicyNone],
		false,
		0,
		types.MaxSizePolicyEncode[types.MaxSizePolicyPerNode],
	)
	_, err := ci.InvokeOnRandomTarget(ctx, req, nil)
	if err != nil {
		fmt.Printf("got error %+v\n", err)
		return err
	}
	return nil
}

func AddMapConfig(ctx context.Context, client *hazelcast.Client, mapName string) error {
	ci := hazelcast.NewClientInternal(client)

	addMapInput := types.DefaultAddMapConfigInput()
	addMapInput.Name = mapName
	addMapInput.IndexConfigs = []types.IndexConfig{{
		Name: "index",
		//Type:       types.IndexTypeHash,
		Attributes: []string{"string"},
	}}

	req := codec.EncodeDynamicConfigAddMapConfigRequest(&addMapInput)
	resp, err := ci.InvokeOnRandomTarget(ctx, req, nil)
	fmt.Printf("AddMapConfig response is %+v\n", resp)
	if err != nil {
		fmt.Printf("got error %+v\n", err)
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	fmt.Printf("mapName: [-n] '%s', address: [-ad] '%s', usePublicIP: [-p] '%t'\n", *mapName, *address, *usePublicIP)

	config := hazelcast.Config{}
	cc := &config.Cluster
	cc.Network.SetAddresses(*address)
	cc.Discovery.UsePublicIP = *usePublicIP
	ctx := context.TODO()
	client, err := hazelcast.StartNewClientWithConfig(ctx, config)
	if err != nil {
		panic(err)
	}

	err = AddMapConfig(ctx, client, *mapName)
	if err != nil {
		panic(err)
	}
	mc, err := GetMapConfig(ctx, client, *mapName)
	if err != nil {
		panic(err)
	}
	prettyPrint(mc)

	err = UpdateMapConfig(ctx, client, *mapName)
	if err != nil {
		panic(err)
	}

	mc2, err := GetMapConfig(ctx, client, *mapName)
	if err != nil {
		panic(err)
	}
	prettyPrint(mc2)

	m, err := client.GetMap(ctx, *mapName)
	m.Put(ctx, "key", "val")
}

func prettyPrint(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(s))
}
