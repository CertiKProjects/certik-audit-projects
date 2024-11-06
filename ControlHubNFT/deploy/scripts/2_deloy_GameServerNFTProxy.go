package scripts

import (
	"bitbucket.org/wemade-tree/waffle/modules/console/execute"
	net "bitbucket.org/wemade-tree/waffle/modules/deploy/network"
	"bitbucket.org/wemade-tree/waffle/modules/sender/network"
	log "bitbucket.org/wemade-tree/wemix-go-tree/common/clog"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"reflect"
)

type GameServerNFTProxy struct {
	*network.Deploy
}

func (g *GameServerNFTProxy) Init(contract *net.Contract) {
	g.Deploy = network.NewDeploy(contract)
}

func (g *GameServerNFTProxy) Loaded() bool {
	return g.Deploy != nil
}

func (g *GameServerNFTProxy) Deployment(net network.INetwork) *net.Receipt {

	var logicContract = "GameServerNFT"
	logic := net.GetContract(logicContract)
	g.Contract.Logic = logic

	value := big.NewInt(0)
	proxyAdmin := net.GetConfig().Roles["proxy_admin"].Address
	superAdmin := net.GetConfig().Roles["game_server_super_admin"].Address
	blacklist := net.GetAddress("BlackOrWhiteList")

	initData, err := logic.Abi.Pack("initialize", superAdmin, blacklist)
	if err != nil {
		log.Error("Failed to pack initData for proxy deployment", "error", err)
		return nil
	}

	g.Contract.Name = "GameServerNFTProxy"

	return net.Deploy(
		g.Contract,
		value,
		logic.Address,
		proxyAdmin,
		initData,
	)
}

func (g *GameServerNFTProxy) Validation(net network.INetwork) {
}

func (g *GameServerNFTProxy) Execution(net network.INetwork) error {

	proxy := net.GetContract("GameServerNFTProxy")
	logic := net.GetContract("GameServerNFT")
	if proxy == nil || logic == nil {
		log.Error("Failed to get contract", "proxyIsEmpty", reflect.ValueOf(proxy).IsZero(), "logicIsEmpty", reflect.ValueOf(logic).IsZero())
		return fmt.Errorf("contract is nil")
	}
	sender := net.Sender()

	admins := []common.Address{
		net.GetConfig().Roles["game_server_admin_role"].Address,
	}

	validAdmins := []common.Address{}
	zeroAddress := common.Address{}

	for _, admin := range admins {
		if admin != zeroAddress {
			validAdmins = append(validAdmins, admin)
		} else {
			log.Warn("Zero address detected and skipped")
		}
	}

	execute.ExecuteContract(sender, sender.Contracts["GameServerNFTProxy"], "grantAdminRoles", big.NewInt(0), false, validAdmins)

	return nil
}
