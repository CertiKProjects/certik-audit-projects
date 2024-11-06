package main

import (
	"flag"
	"log"

	"bitbucket.org/wemade-tree/waffle/modules/console/compile"
	"bitbucket.org/wemade-tree/waffle/modules/deploy/utils"
	"bitbucket.org/wemade-tree/waffle/modules/migration"
	"bitbucket.org/wemade-tree/waffle/modules/sender/network"
	"bitbucket.org/wemade-tree/waffle/projects/ControlHubNFT/deploy/scripts"
)

var ConsoleDir = utils.GetWorkDir("../projects/ControlHubNFT/console.sol")

func main() {
	cfgPath := flag.String("config", "", "conf file path")
	keystore := flag.String("keystore", "", "Keystore path")
	threshold := flag.Int("threshold", 1, "threshold")
	filter := flag.String("filter", "", "filters")
	from := flag.String("from", "", "from address")
	dataDir := flag.String("datadir", "./data", "data dir")
	isRecoreMode := flag.Bool("record", false, "isRecoreMode")
	isTestMode := flag.Bool("test", false, "isTestMode")
	chain := flag.String("chain", "", "chain name")
	isVerifyingCode := flag.Bool("verify", false, "isVerifyingCode")
	flag.Parse()

	cMap, cAry := compile.CompileContractsBoth(ConsoleDir)

	if len(cAry) == 0 {
		log.Fatal("No contracts found in cAry")
	}

	netObj := network.NewNetwork(
		*from, *dataDir, *chain, cMap, cAry, *cfgPath, *keystore,
		*threshold, false, false, *isVerifyingCode, *isRecoreMode, *isTestMode,
	)

	scripts.InitMigration()
	migration.NewMigration(cAry, *filter, "", netObj, scripts.MigrationMap).Run()

	log.Println("GameServerNFT deployment finished.")
}
