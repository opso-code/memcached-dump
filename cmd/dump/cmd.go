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
		Short:   "dump data to local file - 导出memcached所有key到当前目录文件中",
		Example: "dump 127.0.0.1:11211",
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

			num, err := cli.DumpToFile()
			if err != nil {
				cobra.CheckErr(fmt.Errorf("get memcached keys failed %s", err))
			}
			fmt.Printf("memcached %s items dump to file success count %d\n", address, num)
		},
	}
}
