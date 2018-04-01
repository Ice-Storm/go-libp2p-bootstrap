package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"time"

	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

type Bootstrap struct {
	minPeers                int
	bootstrapPeers          []*peerstore.PeerInfo
	host                    host.Host
	notifiee                *net.NotifyBundle
	interfaceListenerLocked bool
	bootstrapInterval       time.Duration
}

//Lock the interface listener
func (b *Bootstrap) lockInterfaceListener() {

	if b.interfaceListenerLocked == true {
		panic("Interface listener is already locked")
	}

	b.interfaceListenerLocked = true
}

//Unlock the interface listener
func (b *Bootstrap) unlockInterfaceListener() {

	if b.interfaceListenerLocked == false {
		panic("Interface listener is already unlocked")
	}

	b.interfaceListenerLocked = false
}

//Is the interface listener locked
func (b *Bootstrap) isInterfaceListenerLocked() bool {
	return b.interfaceListenerLocked
}

//Get the amount of peer's we are connected to
func (b *Bootstrap) amountConnPeers() int {
	return len(b.host.Network().Peers())
}

//Register a network state change handler
func (b *Bootstrap) networkInterfaceListener() {

	//Only register listener when we are connected
	//to too less peer's
	if b.amountConnPeers() >= b.minPeers {
		return
	}

	//Lock down the interface listener
	//to prevent a second listener registration
	b.lockInterfaceListener()

	//Register latest network state
	mas, err := b.host.Network().InterfaceListenAddresses()
	if err != nil {
		panic(err)
	}
	lastNetworkState := len(mas)

	go func() {

		for {

			//Get current network state
			mas, err := b.host.Network().InterfaceListenAddresses()
			if err != nil {
				panic(err)
			}

			//Bootstrap on network delta (delta between the amount of addresses)
			if len(mas) != lastNetworkState {
				lastNetworkState = len(mas)
				b.bootstrap()

				//We can un register the handler when we are connected to enough peer's
				if len(b.host.Network().Peers()) >= b.minPeers {
					break
				}

			}

			//Pause before we continue the dial attempts
			time.Sleep(b.bootstrapInterval)

		}

	}()

}

//Start bootstrapping
func (b *Bootstrap) bootstrap() []error {

	c := make(chan struct{})

	var errorStack []error

	for _, v := range b.bootstrapPeers {

		go func() {
			if b.amountConnPeers() < b.minPeers {
				ctx := context.Background()
				err := b.host.Connect(ctx, *v)
				if err != nil {
					fmt.Println("Failed to connect to: ", v)
					errorStack = append(errorStack, err)
					c <- struct{}{}
					return
				}
				fmt.Println("connected to: ", v)
				c <- struct{}{}
				return
			}
			c <- struct{}{}
		}()

		<-c

	}

	return errorStack

}

//Start bootstrapping
func (b *Bootstrap) Start(bootstrapInterval time.Duration) {

	b.bootstrapInterval = bootstrapInterval

	//Listener
	notifyBundle := net.NotifyBundle{
		DisconnectedF: func(network net.Network, conn net.Conn) {
			fmt.Println("Dropped connnection to peer: ", conn.RemotePeer().String())
			if b.isInterfaceListenerLocked() == false {
				b.networkInterfaceListener()
			}
		},
	}

	//Register listener to react on dropped connections
	b.host.Network().Notify(&notifyBundle)

	if err := b.bootstrap(); err != nil {
		//In case we fail to start,
		//Register network interface listener
		b.networkInterfaceListener()
	}

}

//Create new bootstrap service
func NewBootstrap(h host.Host, bootstrapPeers []string, minPeers int) (error, Bootstrap) {

	if minPeers > len(bootstrapPeers) {
		return errors.New(fmt.Sprintf("Too less bootstrapping nodes. Expected at least: %d, got: %d", minPeers, len(bootstrapPeers))), Bootstrap{}
	}

	var peers []*peerstore.PeerInfo

	for _, v := range bootstrapPeers {
		addr, err := ma.NewMultiaddr(v)

		if err != nil {
			return err, Bootstrap{}
		}

		pInfo, err := peerstore.InfoFromP2pAddr(addr)

		if err != nil {
			return err, Bootstrap{}
		}

		peers = append(peers, pInfo)
	}

	return nil, Bootstrap{
		minPeers:       minPeers,
		bootstrapPeers: peers,
		host:           h,
		interfaceListenerLocked: false,
	}

}
