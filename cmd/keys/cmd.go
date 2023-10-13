package keys

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"
	"memcached-dump/internal/client"
)

var Cmd *cobra.Command

func init() {
	Cmd = &cobra.Command{
		Use:     "keys",
		Short:   "列出memcached所有key",
		Example: "keys 127.0.0.1:11211",
		Run: func(cmd *cobra.Command, args []string) {
			address := "127.0.0.1:11211"
			if len(args) > 0 {
				address = args[0]
			}
			addr, err := net.ResolveTCPAddr("tcp", address)
			if err != nil {
				cobra.CheckErr(fmt.Errorf("invalid memcached address %s %s", address, err))
			}

			cli, err := client.NewClient(addr)
			if err != nil {
				cobra.CheckErr(fmt.Errorf("connect memcached failed %s", address))
			}
			defer cli.Close()

			keys, err := cli.GetKeys()
			if err != nil {
				cobra.CheckErr(fmt.Errorf("get memcached keys failed %s", err))
			}
			for _, key := range keys {
				fmt.Printf("Key: %s,\tSize: %d,\tExpireAt: %d\n", key.Name, key.Size, key.Exp)
			}
		},
	}
}
