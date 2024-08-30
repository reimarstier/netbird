package cmd

import (
	"context"
	"github.com/netbirdio/netbird/iface"
	"github.com/netbirdio/netbird/util"
	"os"
	"testing"
	"time"

	"github.com/netbirdio/netbird/client/internal"
)

var cliAddr string

func TestUpDaemon(t *testing.T) {
	mgmAddr := startTestingServices(t)

	tempDir := t.TempDir()
	confPath := tempDir + "/config.json"

	ctx := internal.CtxInitState(context.Background())
	state := internal.CtxGetState(ctx)

	_, cliLis := startClientDaemon(t, ctx, "http://"+mgmAddr, confPath)

	cliAddr = cliLis.Addr().String()

	daemonAddr = "tcp://" + cliAddr
	rootCmd.SetArgs([]string{
		"login",
		"--daemon-addr", "tcp://" + cliAddr,
		"--setup-key", "A2C8E62B-38F5-4553-B31E-DD66C696CEBB",
		"--log-file", "",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("expected no error while running up command, got %v", err)
		return
	}
	time.Sleep(time.Second * 3)
	if status, err := state.Status(); err != nil && status != internal.StatusIdle {
		t.Errorf("wrong status after login: %s, %v", internal.StatusIdle, err)
		return
	}

	// Test the setup-key-file flag.
	tempFile, err := os.CreateTemp("", "setup-key")
	if err != nil {
		t.Errorf("could not create temp file, got error %v", err)
		return
	}
	defer os.Remove(tempFile.Name())
	if _, err := tempFile.Write([]byte("A2C8E62B-38F5-4553-B31E-DD66C696CEBB")); err != nil {
		t.Errorf("could not write to temp file, got error %v", err)
		return
	}
	if err := tempFile.Close(); err != nil {
		t.Errorf("unable to close file, got error %v", err)
	}
	rootCmd.SetArgs([]string{
		"login",
		"--daemon-addr", "tcp://" + cliAddr,
		"--setup-key-file", tempFile.Name(),
		"--log-file", "",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("expected no error while running up command, got %v", err)
		return
	}
	time.Sleep(time.Second * 3)
	if status, err := state.Status(); err != nil && status != internal.StatusIdle {
		t.Errorf("wrong status after login: %s, %v", internal.StatusIdle, err)
		return
	}

	rootCmd.SetArgs([]string{
		"up",
		"--daemon-addr", "tcp://" + cliAddr,
		"--log-file", "",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("expected no error while running up command, got %v", err)
		return
	}
	time.Sleep(time.Second * 3)
	if status, err := state.Status(); err != nil && status != internal.StatusConnected {
		t.Errorf("wrong status after connect: %s, %v", status, err)
		return
	}

	// validate generated config
	actualConf := &internal.Config{}
	_, err = util.ReadJson(confPath, actualConf)
	if err != nil {
		t.Errorf("expected proper config file written, got broken %v", err)
	}
	if actualConf.WgIfaceMtu != iface.DefaultMTU {
		t.Errorf("expected WgIfaceMtu %d got %d", iface.DefaultMTU, actualConf.WgIfaceMtu)
	}

	rootCmd.SetArgs([]string{
		"status",
		"--daemon-addr", "tcp://" + cliAddr,
	})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("expected no error while running up command, got %v", err)
		return
	}
	time.Sleep(time.Second * 3)

	rootCmd.SetErr(nil)
	rootCmd.SetArgs([]string{"down", "--daemon-addr", "tcp://" + cliAddr})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("expected no error while running up command, got %v", err)
		return
	}
	// we can't check status here, because context already canceled
}
