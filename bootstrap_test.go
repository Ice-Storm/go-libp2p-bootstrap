package bootstrap

import (
	"context"
	"testing"

	libp2p "github.com/libp2p/go-libp2p"
	require "github.com/stretchr/testify/require"
)

var bootstrapPeers = []string{
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	"/ip4/104.236.176.52/tcp/4001/ipfs/QmSoLnSGccFuZQJzRadHn95W2CrSFmZuTdDWP8HXaHca9z",
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	"/ip4/162.243.248.213/tcp/4001/ipfs/QmSoLueR4xBeUbY9WZ9xGUUxunbKWcrNFTDAadQJmocnWm",
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
	"/ip4/178.62.61.185/tcp/4001/ipfs/QmSoLMeWqB7YGVLJN3pNLQpmmEk35v6wYtsMGLzSr5QBU3",
	"/ip4/104.236.151.122/tcp/4001/ipfs/QmSoLju6m7xTh3DuokvT3886QRYqxAzb1kShaanJgW36yx",
}

func TestNewBootstrap(t *testing.T) {

	ctx := context.Background()
	h, err := libp2p.New(ctx)
	require.Nil(t, err)

	err, bootstrap := NewBootstrap(h, bootstrapPeers, 4)
	require.Nil(t, err)

	require.Equal(t, len(bootstrapPeers), len(bootstrap.bootstrapPeers))
	require.Equal(t, 4, bootstrap.minPeers)

}

func TestLockInterfaceListener(t *testing.T) {

	ctx := context.Background()
	h, err := libp2p.New(ctx)
	require.Nil(t, err)

	err, bootstrap := NewBootstrap(h, bootstrapPeers, 4)
	require.Nil(t, err)

	//isInterfaceListenerLocked should be false as a default
	require.False(t, bootstrap.isInterfaceListenerLocked())
	//Lock the interface listener
	bootstrap.lockInterfaceListener()
	//isInterfaceListenerLocked should be true since we locked it
	require.True(t, bootstrap.isInterfaceListenerLocked())

}

func TestLockInterfaceListenerError(t *testing.T) {

	ctx := context.Background()
	h, err := libp2p.New(ctx)
	require.Nil(t, err)

	err, bootstrap := NewBootstrap(h, bootstrapPeers, 4)
	require.Nil(t, err)

	require.Panics(t, func() {
		bootstrap.lockInterfaceListener()
		//Second lock should panic since
		//we need to unlock before we lock again
		bootstrap.lockInterfaceListener()
	})

}

func TestUnlockInterfaceListenerError(t *testing.T) {
	ctx := context.Background()
	h, err := libp2p.New(ctx)
	require.Nil(t, err)

	err, bootstrap := NewBootstrap(h, bootstrapPeers, 4)
	require.Nil(t, err)

	require.Panics(t, func() {
		bootstrap.unlockInterfaceListener()
		//Second lock should panic since
		//we need to unlock before we lock again
		bootstrap.unlockInterfaceListener()
	})
}
