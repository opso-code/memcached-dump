package stats

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"

	"memcached-dump/internal/client"
)

var Cmd *cobra.Command

func init() {
	Cmd = &cobra.Command{
		Use:     "stats",
		Short:   "Execute stats command - 查询memcached stats信息",
		Example: "version 127.0.0.1:11211",
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

			info, err := cli.Stats()
			if err != nil {
				cobra.CheckErr(fmt.Errorf("get memcached version failed %s", err))
			}
			fmt.Printf("memcached %s stats:\n", address)
			fmt.Println(strings.Join(info.Raw, "\n"))
		},
	}
}
