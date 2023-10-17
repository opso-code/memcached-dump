package cmd

import (
	"fmt"
	"os"
	"runtime"

	"memcached-dump/cmd/count"
	"memcached-dump/cmd/dump"
	"memcached-dump/cmd/transfer"

	"github.com/spf13/cobra"

	"memcached-dump/cmd/keys"
	"memcached-dump/cmd/stats"
	"memcached-dump/internal/version"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "memcached-dump",
	Short: "memcached导出工具",
	Long: `memcached-dump
如果你的memcached实例版本是大于1.4.31，且不为1.5.1/1.5.2/1.5.3，可以尝试使用这个工具导出所有数据，否则可能数据导出不全
If your version of Memcached is greater than 1.4.31, and it is not 1.5.1/1.5.2/1.5.3, you can try to use this tool to export all data, otherwise the data export may be incomplete.
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = fmt.Sprintf("%s %s/%s", version.BuildVersion, runtime.GOOS, runtime.GOARCH)
	rootCmd.AddCommand(keys.Cmd, stats.Cmd, dump.Cmd, count.Cmd, transfer.Cmd)
}
