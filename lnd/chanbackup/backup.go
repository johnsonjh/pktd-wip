package chanbackup

import (
	"net"

	"github.com/pkt-cash/pktd/btcec"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/lnd/channeldb"
	"github.com/pkt-cash/pktd/pktlog/log"
	"github.com/pkt-cash/pktd/wire"
)

// LiveChannelSource is an interface that allows us to query for the set of
// live channels. A live channel is one that is open, and has not had a
// commitment transaction broadcast.
type LiveChannelSource interface {
	// FetchAllChannels returns all known live channels.
	FetchAllChannels() ([]*channeldb.OpenChannel, er.R)

	// FetchChannel attempts to locate a live channel identified by the
	// passed chanPoint.
	FetchChannel(chanPoint wire.OutPoint) (*channeldb.OpenChannel, er.R)

	// AddrsForNode returns all known addresses for the target node public
	// key.
	AddrsForNode(nodePub *btcec.PublicKey) ([]net.Addr, er.R)
}

// assembleChanBackup attempts to assemble a static channel backup for the
// passed open channel. The backup includes all information required to restore
// the channel, as well as addressing information so we can find the peer and
// reconnect to them to initiate the protocol.
func assembleChanBackup(chanSource LiveChannelSource,
	openChan *channeldb.OpenChannel) (*Single, er.R) {
	log.Debugf("Crafting backup for ChannelPoint(%v)",
		openChan.FundingOutpoint)

	// First, we'll query the channel source to obtain all the addresses
	// that are are associated with the peer for this channel.
	nodeAddrs, err := chanSource.AddrsForNode(openChan.IdentityPub)
	if err != nil {
		return nil, err
	}

	single := NewSingle(openChan, nodeAddrs)

	return &single, nil
}

// FetchBackupForChan attempts to create a plaintext static channel backup for
// the target channel identified by its channel point. If we're unable to find
// the target channel, then an error will be returned.
func FetchBackupForChan(chanPoint wire.OutPoint,
	chanSource LiveChannelSource) (*Single, er.R) {
	// First, we'll query the channel source to see if the channel is known
	// and open within the database.
	targetChan, err := chanSource.FetchChannel(chanPoint)
	if err != nil {
		// If we can't find the channel, then we return with an error,
		// as we have nothing to  backup.
		return nil, er.Errorf("unable to find target channel")
	}

	// Once we have the target channel, we can assemble the backup using
	// the source to obtain any extra information that we may need.
	staticChanBackup, err := assembleChanBackup(chanSource, targetChan)
	if err != nil {
		return nil, er.Errorf("unable to create chan backup: %v", err)
	}

	return staticChanBackup, nil
}

// FetchStaticChanBackups will return a plaintext static channel back up for
// all known active/open channels within the passed channel source.
func FetchStaticChanBackups(chanSource LiveChannelSource) ([]Single, er.R) {
	// First, we'll query the backup source for information concerning all
	// currently open and available channels.
	openChans, err := chanSource.FetchAllChannels()
	if err != nil {
		return nil, err
	}

	// Now that we have all the channels, we'll use the chanSource to
	// obtain any auxiliary information we need to craft a backup for each
	// channel.
	staticChanBackups := make([]Single, 0, len(openChans))
	for _, openChan := range openChans {
		chanBackup, err := assembleChanBackup(chanSource, openChan)
		if err != nil {
			return nil, err
		}

		staticChanBackups = append(staticChanBackups, *chanBackup)
	}

	return staticChanBackups, nil
}
