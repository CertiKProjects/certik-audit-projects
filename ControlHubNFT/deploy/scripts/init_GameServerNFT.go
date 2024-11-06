package scripts

import "bitbucket.org/wemade-tree/waffle/modules/sender/network"

var MigrationMap = map[string]network.IDeploy{}

func appendMigrationMap(name string, d network.IDeploy) {
	MigrationMap[name] = d
}

func InitMigration() {
	appendMigrationMap("GameServerNFTProxy", &GameServerNFTProxy{})
	appendMigrationMap("GameServerNFT", &GameServerNFT{})
}
