package main

import (
	"example/scanner/walk"
	_ "net/http/pprof"
	"os"

	api "github.com/chaitin/libveinmind/go"
	"github.com/chaitin/libveinmind/go/cmd"
	"github.com/chaitin/libveinmind/go/plugin"
	"github.com/chaitin/libveinmind/go/plugin/log"
	"github.com/chaitin/veinmind-common-go/service/report"

	_ "github.com/chaitin/veinmind-tools/plugins/go/veinmind-malicious/config"
	_ "github.com/chaitin/veinmind-tools/plugins/go/veinmind-malicious/database"
	_ "github.com/chaitin/veinmind-tools/plugins/go/veinmind-malicious/database/model"
)

var (
	ReportService = &report.Service{}
	rootCmd       = &cmd.Command{}
	scanCmd       = &cmd.Command{Use: "scan"}

	scanImageCmd = &cmd.Command{
		Use:   "image",
		Short: "Scan image malicious files",
	}
)

func scan(_ *cmd.Command, image api.Image) error {

	err := walk.Scan(image)
	if err != nil {
		log.Error(err)
		return nil
	}

	return nil
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.AddCommand(report.MapReportCmd(cmd.MapImageCommand(scanImageCmd, scan), ReportService))
	rootCmd.AddCommand(cmd.NewInfoCommand(plugin.Manifest{
		Name:        "magic-shield-test",
		Author:      "cnapp-team",
		Description: "cnapp scanner image file",
	}))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
