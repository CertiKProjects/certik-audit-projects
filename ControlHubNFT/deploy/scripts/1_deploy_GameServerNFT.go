package scripts

import (
	net "bitbucket.org/wemade-tree/waffle/modules/deploy/network"
	"bitbucket.org/wemade-tree/waffle/modules/sender/network"
	"math/big"
)

type GameServerNFT struct {
	*network.Deploy
}

func (g *GameServerNFT) Init(contract *net.Contract) {
	g.Deploy = network.NewDeploy(contract)
}

func (g *GameServerNFT) Loaded() bool {
	return g.Deploy != nil
}

func (g *GameServerNFT) Deployment(net network.INetwork) *net.Receipt {
	value := big.NewInt(0)
	g.Contract.Name = "GameServerNFT"

	return net.Deploy(
		g.Contract,
		value,
	)
}

func (g *GameServerNFT) Validation(net network.INetwork) {

}

func (g *GameServerNFT) Execution(net network.INetwork) error {
	return nil
}
