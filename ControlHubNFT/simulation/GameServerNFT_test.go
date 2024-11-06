package simulation

import (
	"bitbucket.org/wemade-tree/waffle/modules/backend"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/big"
	"reflect"
	"testing"
)

type GameServerMarket struct {
	client                    *backend.Client
	proxyAdmin                *backend.Account
	baseAdmin                 *backend.Account
	nonAdmin                  *backend.Account
	nonAdmin2                 *backend.Account
	nonAdmin3                 *backend.Account
	gameServerManagementProxy *backend.Contract
	gameServerNFT             *backend.Contract
	blacklist                 *backend.Contract
}

func NewGameServerMarket(t *testing.T) *GameServerMarket {
	p := &GameServerMarket{
		client:                    backend.NewClient(t),
		proxyAdmin:                backend.GenAccount(),
		baseAdmin:                 backend.GenAccount(),
		nonAdmin:                  backend.GenAccount(),
		nonAdmin2:                 backend.GenAccount(),
		nonAdmin3:                 backend.GenAccount(),
		gameServerManagementProxy: backend.NewContract(t, "../contracts/proxy/GameServerNFTProxy.sol", "GameServerNFTProxy"),
		gameServerNFT:             backend.NewContract(t, "../contracts/proxy/GameServerNFT.sol", "GameServerNFT"),
		blacklist:                 backend.NewContract(t, "../../../projects/Blacklist/contracts/BlackOrWhiteList.sol", "BlackOrWhiteList"),
	}

	p.client.TransferWemix(t, p.client.OwnerKey, backend.ToWei(1000000), p.proxyAdmin.Address)
	p.client.TransferWemix(t, p.client.OwnerKey, backend.ToWei(1000000), p.baseAdmin.Address)
	p.client.TransferWemix(t, p.client.OwnerKey, backend.ToWei(1000000), p.nonAdmin.Address)
	p.client.TransferWemix(t, p.client.OwnerKey, backend.ToWei(1000000), p.nonAdmin2.Address)
	p.client.TransferWemix(t, p.client.OwnerKey, backend.ToWei(1000000), p.nonAdmin3.Address)

	p.deployGameServerContracts(t)

	return p
}

func (p *GameServerMarket) deployGameServerContracts(t *testing.T) {

	p.blacklist.Deploy(t, p.client)
	p.gameServerNFT.Deploy(t, p.client)

	initData, err := p.gameServerNFT.Abi.Pack("initialize", p.baseAdmin.Address, p.blacklist.Address)
	if err != nil {
		t.Fatalf("Failed to pack initData for proxy deployment: %v", err)
	}

	p.gameServerManagementProxy.Deploy(t, p.client, p.gameServerNFT.Address, p.client.Owner, initData)

	if p.gameServerManagementProxy.Address.Hex() == "0x0000000000000000000000000000000000000000" {
		t.Fatalf("Failed to deploy GameServerManagementProxy: Proxy address is zero")
	}

	t.Log("GameServerNFT address: ", p.gameServerNFT.Address.Hex())
	t.Log("GameServerManagementProxy address: ", p.gameServerManagementProxy.Address.Hex())
	t.Log("GameServerNFT & GameServerManagementProxy deployed successfully")
}

func TestGameServerMarket(t *testing.T) {

	market := NewGameServerMarket(t)

	t.Run("Grant Admin Role", func(t *testing.T) {
		receipt := market.gameServerManagementProxy.ProxyExecute(t, market.baseAdmin.PrivateKey, market.gameServerNFT, common.Big0, "grantAdminRoles", []common.Address{market.nonAdmin.Address})
		require.NotNil(t, receipt, "Receipt should not be nil, transaction might have failed")

		isAdmin := market.isAdmin(t, market.nonAdmin.Address)
		assert.True(t, isAdmin, "GAME_SERVER_ADMIN should be granted to nonAdmin successfully")
	})

	t.Run("Revoke Admin Role", func(t *testing.T) {

		market.gameServerManagementProxy.ProxyExecute(t, market.baseAdmin.PrivateKey, market.gameServerNFT, common.Big0, "grantAdminRoles", []common.Address{market.nonAdmin.Address})
		t.Log("Success granted to admin")

		market.gameServerManagementProxy.ProxyExecute(t, market.baseAdmin.PrivateKey, market.gameServerNFT, common.Big0, "revokeAdminRoles", []common.Address{market.nonAdmin.Address})
		adminStatus := market.isAdmin(t, market.nonAdmin.Address)
		assert.False(t, adminStatus, "Non-admin role should have been revoked")
	})

	t.Run("Change Super Admin Role", func(t *testing.T) {
		superAdminStatus := market.isSuperAdmin(t, market.baseAdmin.Address)
		assert.True(t, superAdminStatus[market.baseAdmin.Address], "BaseAdmin should have GAME_SERVER_ADMIN")

		market.gameServerManagementProxy.ProxyExecute(t, market.baseAdmin.PrivateKey, market.gameServerNFT, common.Big0, "changeSuperAdmin", market.nonAdmin.Address)

		superAdminStatus = market.isSuperAdmin(t, market.nonAdmin.Address)
		assert.True(t, superAdminStatus[market.nonAdmin.Address], "GAME_SERVER_ADMIN granted to NonAdmin successfully")
	})

	t.Run("Mint Server NFT and check TokenURI", func(t *testing.T) {

		market.gameServerManagementProxy.ProxyExecute(t, market.baseAdmin.PrivateKey, market.gameServerNFT, common.Big0, "mintServerNFT", market.nonAdmin2.Address, "Server_test_123", "TestGame", "WEMADE", "metadataURI")

		result := market.gameServerManagementProxy.ProxyCall(t, market.gameServerNFT, "nextTokenId")
		nextTokenId := result[0].(*big.Int)
		tokenID := new(big.Int).Sub(nextTokenId, big.NewInt(1))
		assert.NotNil(t, tokenID, "Token ID should be non-nil after minting")
		assert.Equal(t, big.NewInt(1), tokenID, "Token ID should be 1")
		t.Log("NFT minted successfully", tokenID, nextTokenId)

		//set token URI
		tokenURIResult := market.gameServerManagementProxy.ProxyCall(t, market.gameServerNFT, "tokenURI", tokenID)
		tokenURI := tokenURIResult[0].(string)
		assert.Equal(t, "metadataURI", tokenURI, "Token URI should match the set value 'metadataURI'")
		t.Log("Token URI verified:", tokenURI)
	})

	t.Run("Batch Mint Server NFT", func(t *testing.T) {

		recipients := []common.Address{market.nonAdmin.Address, market.nonAdmin2.Address}
		serverIds := []string{"Server_123", "Server_456"}
		gameNames := []string{"NightCrows", "NCGame"}
		developers := []string{"WEMADE", "NCSOFT"}
		metadataURIs := []string{"ipfs://metadata1", "ipfs://metadata2"}
		t.Logf("recipients length: %d, serverIds length: %d", len(recipients), len(serverIds))
		t.Logf("gameNames length: %d, developers length: %d, metadataURIs length: %d", len(gameNames), len(serverIds), len(metadataURIs))

		t.Log("Executing batch minting...")
		err := market.gameServerManagementProxy.ProxyExecute(t, market.baseAdmin.PrivateKey, market.gameServerNFT, common.Big0,
			"batchMintServerNFT", recipients, serverIds, gameNames, developers, metadataURIs)

		if err != nil {
			t.Logf("Error type: %s", reflect.TypeOf(err))
		} else {
			t.Log("Batch minting executed successfully")
		}

		result := market.gameServerManagementProxy.ProxyCall(t, market.gameServerNFT, "nextTokenId")
		t.Log("nextTokenId fetched successfully", result)

		nextTokenId := result[0].(*big.Int)
		if nextTokenId.Cmp(big.NewInt(1)) == 0 {
			t.Fatalf("Minting failed: nextTokenId is still 1 after batch mint")
		}
		tokenID1 := new(big.Int).Sub(nextTokenId, big.NewInt(2))
		tokenID2 := new(big.Int).Sub(nextTokenId, big.NewInt(1))

		tokenURI1 := market.getTokenURI(t, tokenID1)
		tokenURI2 := market.getTokenURI(t, tokenID2)

		assert.Equal(t, "ipfs://metadata1", tokenURI1, "Token URI for the first token should match expected value")
		assert.Equal(t, "ipfs://metadata2", tokenURI2, "Token URI for the second token should match expected value")

		t.Log("Batch minting successful: tokenID1", tokenID1, "tokenID2", tokenID2)
	})

	t.Run("Transfer Token with Delegated Approval", func(t *testing.T) {
		leo := market.nonAdmin.Address
		na := market.nonAdmin2.Address
		to := market.nonAdmin3.Address
		t.Log("Minting and Delegating Approval Process", leo.Hex(), na.Hex(), to.Hex())

		market.gameServerManagementProxy.ProxyExecute(
			t,
			market.baseAdmin.PrivateKey,
			market.gameServerNFT,
			common.Big0,
			"mintServerNFT",
			leo,
			"Server_test_123",
			"NightCrows",
			"WEMADE",
			"metadataURI",
		)

		result := market.gameServerManagementProxy.ProxyCall(t, market.gameServerNFT, "nextTokenId")
		nextTokenId := result[0].(*big.Int)
		tokenID := new(big.Int).Sub(nextTokenId, big.NewInt(1))

		market.gameServerManagementProxy.ProxyExecute(
			t,
			market.nonAdmin.PrivateKey,
			market.gameServerNFT,
			common.Big0,
			"approve",
			na,
			tokenID,
		)

		market.gameServerManagementProxy.ProxyExecute(
			t,
			market.nonAdmin2.PrivateKey, // Signed
			market.gameServerNFT,
			common.Big0,
			"safeTransferFrom",
			leo,
			to,
			tokenID,
		)

		newOwner := market.gameServerManagementProxy.ProxyCall(t, market.gameServerNFT, "ownerOf", tokenID)
		assert.Equal(t, newOwner[0], to, "Token should be transferred to the new owner")

		t.Log("Token successfully transferred by delegate from", leo.Hex(), "to", to.Hex())
	})

	t.Run("Update Metadata URI", func(t *testing.T) {
		metadataHash := [32]byte{}
		copy(metadataHash[:], "example-metadata-hash")

		initialURI := "https://mystorageaccount.blob.core.windows.net/wemade/meta.json"

		market.gameServerManagementProxy.ProxyExecute(t, market.baseAdmin.PrivateKey, market.gameServerNFT, common.Big0, "mintServerNFT", market.nonAdmin.Address, "Server_test_123", "TestGame", "WEMADE", initialURI)

		result := market.gameServerManagementProxy.ProxyCall(t, market.gameServerNFT, "nextTokenId")
		nextTokenId := result[0].(*big.Int)
		tokenID := new(big.Int).Sub(nextTokenId, big.NewInt(1))
		t.Log("Metadata test", "tokenID", tokenID, "nextTokenId", nextTokenId)

		newURI := "new-updated-metadata-uri"
		market.updateMetadataURI(t, tokenID, newURI)

		assert.Equal(t, newURI, market.getTokenURI(t, tokenID), "Metadata URI should be updated")
	})

	t.Run("Burn Server NFT", func(t *testing.T) {

		serverID := "serverID_123"
		market.gameServerManagementProxy.ProxyExecute(t, market.baseAdmin.PrivateKey, market.gameServerNFT, common.Big0, "mintServerNFT", market.baseAdmin.Address, serverID, "gameName", "WEMADE", "ipfs://metadata")

		result := market.gameServerManagementProxy.ProxyCall(t, market.gameServerNFT, "nextTokenId")
		nextTokenId := result[0].(*big.Int)
		tokenID := new(big.Int).Sub(nextTokenId, big.NewInt(1))
		t.Log("tokenID", tokenID, "nextTokenId", nextTokenId)

		serverIDString := market.getServerInfo(t, tokenID)
		assert.Equal(t, serverIDString, "serverID_123")
		t.Log("tokenID", tokenID, "nextTokenId", nextTokenId, "serverID", serverIDString)

		market.gameServerManagementProxy.ProxyExecute(t, market.baseAdmin.PrivateKey, market.gameServerNFT, common.Big0, "burnServerNFT", tokenID)

		serverIDString = market.getServerInfo(t, tokenID)
		assert.Equal(t, serverIDString, "")
	})

	t.Run("Check BalanceOfAll", func(t *testing.T) {
		market = NewGameServerMarket(t)

		market.gameServerManagementProxy.ProxyExecute(t, market.baseAdmin.PrivateKey, market.gameServerNFT, common.Big0,
			"mintServerNFT", market.baseAdmin.Address, "Server_test_123", "TestGame", "WEMADE", "metadataURI")

		market.gameServerManagementProxy.ProxyExecute(t, market.baseAdmin.PrivateKey, market.gameServerNFT, common.Big0,
			"mintServerNFT", market.baseAdmin.Address, "serverID_2", "FunGame2", "WEMADE2", "ipfs://metadata2")

		result := market.gameServerManagementProxy.ProxyCall(t, market.gameServerNFT, "balanceOfAll", market.baseAdmin.Address)
		tokenIds := result[0].([]*big.Int)

		t.Log("Tokens owned by baseAdmin:", tokenIds)
		assert.Equal(t, len(tokenIds), 2, "Expected 2 tokens for baseAdmin")

		assert.Equal(t, tokenIds[0].Int64(), int64(1), "First token should have ID 1")
		assert.Equal(t, tokenIds[1].Int64(), int64(2), "Second token should have ID 2")
	})

	t.Run("Check Supports interface", func(t *testing.T) {
		interfaceID := [4]byte{0x80, 0xac, 0x58, 0xcd} // ERC721 인터페이스 ID
		result := market.gameServerManagementProxy.ProxyCall(t, market.gameServerNFT, "supportsInterface", interfaceID)
		assert.Equal(t, true, result[0].(bool), "Contract should support ERC721 interface")
	})
}

func (p *GameServerMarket) isAdmin(t *testing.T, account common.Address) bool {
	result := p.gameServerManagementProxy.ProxyCall(t, p.gameServerNFT, "isAdmin", account)
	if result == nil {
		t.Fatalf("Failed to check admin role for account: %s", account.Hex())
	}
	return result[0].(bool)
}

func (p *GameServerMarket) isSuperAdmin(t *testing.T, account common.Address) map[common.Address]bool {
	superAdminRoleResult := p.gameServerManagementProxy.ProxyCall(t, p.gameServerNFT, "SUPER_ADMIN_ROLE")
	SuperAdminRoleHash := superAdminRoleResult[0].([32]byte)

	superAdminStatus := make(map[common.Address]bool)

	result := p.gameServerManagementProxy.ProxyCall(t, p.gameServerNFT, "hasRole", SuperAdminRoleHash, account)
	if result == nil {
		t.Fatalf("Failed to check super admin role for account: %s", account.Hex())
	}
	superAdminStatus[account] = result[0].(bool)

	return superAdminStatus
}

func (p *GameServerMarket) updateMetadataURI(t *testing.T, tokenID *big.Int, newURI string) {
	p.gameServerManagementProxy.ProxyExecute(t, p.baseAdmin.PrivateKey, p.gameServerNFT, nil, "setTokenURI", tokenID, newURI)
}

func (p *GameServerMarket) getTokenURI(t *testing.T, tokenID *big.Int) string {
	result := p.gameServerManagementProxy.ProxyCall(t, p.gameServerNFT, "tokenURI", tokenID)
	return result[0].(string)
}

func (p *GameServerMarket) getServerInfo(t *testing.T, tokenID *big.Int) string {
	result := p.gameServerManagementProxy.ProxyCall(t, p.gameServerNFT, "getServerInfo", tokenID)
	if result == nil {
		t.Fatal("Failed to get server info")
	}
	serverID := result[0].(string)
	return serverID
}
