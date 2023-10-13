package dump

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"
	"memcached-dump/internal/client"
)

var Cmd *cobra.Command

func init() {
	Cmd = &cobra.Command{
		Use:     "dump",
		Short:   "读取memcached所有key到另一个memcached实例中",
		Example: "dump 127.0.0.1:11211 127.0.0.1:11212",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				cobra.CheckErr(fmt.Errorf("invalid memcached params %v", args))
			}

			address := args[0]
			srcAddr, err := net.ResolveTCPAddr("tcp", address)
			if err != nil {
				cobra.CheckErr(fmt.Errorf("invalid memcached source address %s %s", address, err))
			}

			target := args[1]
			dstAddr, err := net.ResolveTCPAddr("tcp", target)
			if err != nil {
				cobra.CheckErr(fmt.Errorf("invalid memcached target address %s %s", target, err))
			}

			cli, err := client.NewClient(srcAddr)
			if err != nil {
				cobra.CheckErr(fmt.Errorf("connect memcached failed %s", address))
			}
			defer cli.Close()

			num, err := cli.DumpTo(dstAddr)
			if err != nil {
				cobra.CheckErr(fmt.Errorf("get memcached dump failed %s", err))
			}
			fmt.Printf("memcached %s to %s success %d\n", address, target, num)
		},
	}
}
