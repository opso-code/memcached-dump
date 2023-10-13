package cmd

import (
	"fmt"
	"memcached-dump/cmd/count"
	"memcached-dump/cmd/dump"
	"memcached-dump/cmd/store"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"memcached-dump/cmd/keys"
	"memcached-dump/cmd/ver"
	"memcached-dump/internal/version"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "memcached-dump",
	Short: "memcached导出工具",
	Long: `
由于memcached在v1.4.31才支持的lru_crawler metadump命令，之前的版本只能使用stats cachedump命令，有1M的数据大小限制（大概几W个key，看key的长度），所以低版本无法导出完整数据。
如果遇到命令超时程序也会转而使用旧的方式获取key，于是也会有1M大小限制。
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
	rootCmd.AddCommand(keys.Cmd, ver.Cmd, store.Cmd, count.Cmd, dump.Cmd)
}
