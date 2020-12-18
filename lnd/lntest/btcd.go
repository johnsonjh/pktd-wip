// +build !bitcoind,!neutrino

package lntest

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/pkt-cash/pktd/btcjson"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/chaincfg"
	"github.com/pkt-cash/pktd/integration/rpctest"
	"github.com/pkt-cash/pktd/rpcclient"
)

// logDirPattern is the pattern of the name of the temporary log directory.
const logDirPattern = "%s/.backendlogs"

// temp is used to signal we want to establish a temporary connection using the
// btcd Node API.
//
// NOTE: Cannot be const, since the node API expects a reference.
var temp = "temp"

// BtcdBackendConfig is an implementation of the BackendConfig interface
// backed by a btcd node.
type BtcdBackendConfig struct {
	// rpcConfig houses the connection config to the backing btcd instance.
	rpcConfig rpcclient.ConnConfig

	// harness is the backing btcd instance.
	harness *rpctest.Harness

	// minerAddr is the p2p address of the miner to connect to.
	minerAddr string
}

// A compile time assertion to ensure BtcdBackendConfig meets the BackendConfig
// interface.
var _ BackendConfig = (*BtcdBackendConfig)(nil)

// GenArgs returns the arguments needed to be passed to LND at startup for
// using this node as a chain backend.
func (b BtcdBackendConfig) GenArgs() []string {
	var args []string
	encodedCert := hex.EncodeToString(b.rpcConfig.Certificates)
	args = append(args, "--bitcoin.node=btcd")
	args = append(args, fmt.Sprintf("--btcd.rpchost=%v", b.rpcConfig.Host))
	args = append(args, fmt.Sprintf("--btcd.rpcuser=%v", b.rpcConfig.User))
	args = append(args, fmt.Sprintf("--btcd.rpcpass=%v", b.rpcConfig.Pass))
	args = append(args, fmt.Sprintf("--btcd.rawrpccert=%v", encodedCert))

	return args
}

// ConnectMiner is called to establish a connection to the test miner.
func (b BtcdBackendConfig) ConnectMiner() er.R {
	return b.harness.Node.Node(btcjson.NConnect, b.minerAddr, &temp)
}

// DisconnectMiner is called to disconnect the miner.
func (b BtcdBackendConfig) DisconnectMiner() er.R {
	return b.harness.Node.Node(btcjson.NDisconnect, b.minerAddr, &temp)
}

// Name returns the name of the backend type.
func (b BtcdBackendConfig) Name() string {
	return "btcd"
}

// NewBackend starts a new rpctest.Harness and returns a BtcdBackendConfig for
// that node. miner should be set to the P2P address of the miner to connect
// to.
func NewBackend(miner string, netParams *chaincfg.Params) (
	*BtcdBackendConfig, func() er.R, er.R) {
	baseLogDir := fmt.Sprintf(logDirPattern, GetLogDir())
	args := []string{
		"--rejectnonstd",
		"--txindex",
		"--trickleinterval=100ms",
		"--debuglevel=debug",
		"--logdir=" + baseLogDir,
		"--nowinservice",
		// The miner will get banned and disconnected from the node if
		// its requested data are not found. We add a nobanning flag to
		// make sure they stay connected if it happens.
		"--nobanning",
	}
	chainBackend, err := rpctest.New(netParams, nil, args)
	if err != nil {
		return nil, nil, er.Errorf("unable to create btcd node: %v", err)
	}

	if err := chainBackend.SetUp(false, 0); err != nil {
		return nil, nil, er.Errorf("unable to set up btcd backend: %v", err)
	}

	bd := &BtcdBackendConfig{
		rpcConfig: chainBackend.RPCConfig(),
		harness:   chainBackend,
		minerAddr: miner,
	}

	cleanUp := func() er.R {
		var errStr string
		if err := chainBackend.TearDown(); err != nil {
			errStr += err.String() + "\n"
		}

		// After shutting down the chain backend, we'll make a copy of
		// the log file before deleting the temporary log dir.
		logFile := baseLogDir + "/" + netParams.Name + "/btcd.log"
		logDestination := fmt.Sprintf(
			"%s/output_btcd_chainbackend.log", GetLogDir(),
		)
		err := CopyFile(logDestination, logFile)
		if err != nil {
			errStr += fmt.Sprintf("unable to copy file: %v\n", err)
		}
		if errr := os.RemoveAll(baseLogDir); errr != nil {
			errStr += fmt.Sprintf(
				"cannot remove dir %s: %v\n", baseLogDir, errr,
			)
		}
		if errStr != "" {
			return er.New(errStr)
		}
		return nil
	}

	return bd, cleanUp, nil
}
