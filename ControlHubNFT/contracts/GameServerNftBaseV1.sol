// SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import "../../../contracts/openzeppelin-contracts-upgradeable/token/ERC721/extensions/ERC721EnumerableUpgradeable.sol";
import "../../../contracts/openzeppelin-contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "../../../contracts/openzeppelin-contracts-upgradeable/token/ERC721/extensions/ERC721URIStorageUpgradeable.sol";
import "../../../contracts/openzeppelin-contracts-upgradeable/security/ReentrancyGuardUpgradeable.sol";
import "../../../projects/Blacklist/contracts/BlackOrWhiteChecker.sol";
import "../../../projects/Blacklist/contracts/interface/IBlackOrWhiteList.sol";

/**
 * @title GameServerNftBaseV1
 * @dev Base contract for managing server-related NFTs. Provides functionalities
 *      such as role-based access control, token minting, metadata management,
 *      and server information storage.
 */
contract GameServerNftBaseV1 is Initializable, ERC721EnumerableUpgradeable, ERC721URIStorageUpgradeable, AccessControlUpgradeable, ReentrancyGuardUpgradeable, BlackOrWhiteChecker {
    bytes32 public constant GAME_SERVER_SUPER_ADMIN = keccak256("game_server_super_admin");
    bytes32 public constant GAME_SERVER_ADMIN = keccak256("game_server_admin");
    IBlackOrWhiteList public blackOrWhiteList;

    struct ServerInfo {
        string serverId;
        string gameName;
        string developer;
    }

    mapping(uint256 => ServerInfo) public serverInfoByTokenId;
    uint256 public nextTokenId;
    uint256[50] private __gap;

    event ServerMinted(address indexed to, uint256 indexed tokenId, string serverId, string gameName);
    event MetadataUpdated(uint256 indexed tokenId, string metadataURI);
    event ServerInfoDeleted(uint256 indexed tokenId);
    event AdminRoleGranted(address indexed account, address indexed grantedBy);
    event AdminRoleRevoked(address indexed account, address indexed revokedBy);
    event SuperAdminRoleChanged(address indexed oldSuperAdmin, address indexed newSuperAdmin);

    /**
     * @dev Initializes the contract and sets up roles.
     * @param superAdmin The address that will receive the GAME_SERVER_SUPER_ADMIN.
     */
    function initialize(address superAdmin, IBlackOrWhiteList _blackOrWhiteList) external initializer {
        __ERC721_init("GameServerNFT", "GameServerNFT");
        __ERC721Enumerable_init();
        __ERC721URIStorage_init();
        __AccessControl_init();
        __ReentrancyGuard_init();

        // Assign the GAME_SERVER_SUPER_ADMIN to the provided superAdmin address
        _setupRole(GAME_SERVER_SUPER_ADMIN, superAdmin);

        // Set GAME_SERVER_ADMIN to GAME_SERVER_SUPER_ADMIN only
        _setRoleAdmin(GAME_SERVER_ADMIN, GAME_SERVER_SUPER_ADMIN);

        // By setting GAME_SERVER_SUPER_ADMIN as its own admin, we make GAME_SERVER_SUPER_ADMINs capable of managing other GAME_SERVER_SUPER_ADMINs.
        _setRoleAdmin(GAME_SERVER_SUPER_ADMIN, GAME_SERVER_SUPER_ADMIN);

        // Initialize the blacklist/whitelist contract
        blackOrWhiteList = _blackOrWhiteList;

        nextTokenId = 1;
    }

    modifier onlyAdmin() {
        require(hasRole(GAME_SERVER_SUPER_ADMIN, msg.sender) || hasRole(GAME_SERVER_ADMIN, msg.sender), "GameServerNFT: Access denied! Requires GAME_SERVER_SUPER_ADMIN or GAME_SERVER_ADMIN");
        _;
    }

    /**
     * @dev Returns the token URI for a given token.
     * @param tokenId ID of the token to retrieve the URI for.
     */
    function tokenURI(uint256 tokenId) public view override(ERC721Upgradeable, ERC721URIStorageUpgradeable) returns (string memory) {
        return super.tokenURI(tokenId);
    }

    /**
     * @dev Checks if an account has the GAME_SERVER_ADMIN role.
     * @param adminAccount Address to check for the role.
     * @return True if the adminAccount has the role, otherwise false.
     */
    function isAdmin(address adminAccount) public view returns (bool) {
        return hasRole(GAME_SERVER_ADMIN, adminAccount);
    }

    /**
     * @dev Checks if an account has the GAME_SERVER_SUPER_ADMIN role.
     * @param superAdminAccount Address to check for the role.
     * @return True if the superAdminAccount has the role, otherwise false.
     */
    function isSuperAdmin(address superAdminAccount) public view returns (bool) {
        return hasRole(GAME_SERVER_SUPER_ADMIN, superAdminAccount);
    }

    /**
     * @dev Changes the GAME_SERVER_SUPER_ADMIN to a new address and revokes it from the old address.
     * @param newSuperAdmin The address to be assigned the new GAME_SERVER_SUPER_ADMIN role.
     */
    function changeSuperAdmin(address newSuperAdmin) external {
        require(newSuperAdmin != address(0), "GameServerNFT: New super admin cannot be the zero address");
        require(isSuperAdmin(msg.sender), "GameServerNFT: Invalid super admin or unauthorized");

        grantRole(GAME_SERVER_SUPER_ADMIN, newSuperAdmin);
        revokeRole(GAME_SERVER_SUPER_ADMIN, msg.sender);

        emit SuperAdminRoleChanged(msg.sender, newSuperAdmin);
    }

    /**
     * @dev Grants the GAME_SERVER_ADMIN to a list of accounts.
     * @param accounts Array of addresses to grant the role to.
     */
    function grantAdminRoles(address[] memory accounts) external {

        for (uint256 i = 0; i < accounts.length; i++) {
            if (!isAdmin(accounts[i])) {
                grantRole(GAME_SERVER_ADMIN, accounts[i]);
                emit AdminRoleGranted(accounts[i], msg.sender);
            }
        }
    }

    /**
     * @dev Revokes the GAME_SERVER_ADMIN from a list of accounts.
     * @param adminAccounts Array of addresses to revoke the role from.
     */
    function revokeAdminRoles(address[] memory adminAccounts) external {

        for (uint256 i = 0; i < adminAccounts.length; i++) {
            require(isAdmin(adminAccounts[i]), "GameServerNFT: Account does not have GAME_SERVER_ADMIN roles");
            revokeRole(GAME_SERVER_ADMIN, adminAccounts[i]);

            emit AdminRoleRevoked(adminAccounts[i], msg.sender);
        }
    }

    /**
     * @dev Returns all token IDs owned by an account.
     * @param owner Address to check the tokens of.
     * @return Array of token IDs owned by the address.
     */
    function balanceOfAll(address owner) external view returns (uint256[] memory) {
        uint256 balance = balanceOf(owner);
        uint256[] memory tokenIds = new uint256[](balance);

        for (uint256 i = 0; i < balance; i++) {
            tokenIds[i] = tokenOfOwnerByIndex(owner, i);
        }
        return tokenIds;
    }

    /**
     * @dev Updates the metadata URI and hash for a token.
     * @param tokenId ID of the token to update.
     * @param newMetadataURI New URI to be set.
     */
    function updateMetadata(uint256 tokenId, string memory newMetadataURI) external onlyAdmin{
        require(_exists(tokenId), "GameServerNFT: Token does not exist.");

        _setTokenURI(tokenId, newMetadataURI);

        emit MetadataUpdated(tokenId, newMetadataURI);
    }

    /**
     * @dev Checks if the contract supports a given interface.
     * @param interfaceId ID of the interface to check.
     * @return True if the interface is supported, otherwise false.
     */
    function supportsInterface(bytes4 interfaceId) public view override(ERC721EnumerableUpgradeable, ERC721URIStorageUpgradeable, AccessControlUpgradeable) returns (bool) {
        return super.supportsInterface(interfaceId);
    }

    /**
     * @dev Internal function to mint an NFT and set its token URI.
     * @param to Address to receive the newly minted NFT.
     * @param tokenId ID of the token to be minted.
     * @param metadataURI URI pointing to the metadata for the NFT.
     */
    function _mintServerNFTBase(address to, uint256 tokenId, string memory metadataURI) internal onlyAdmin {
        _safeMint(to, tokenId);
        _setTokenURI(tokenId, metadataURI);
    }

    /**
     * @dev Internal function to set server information.
     * @param tokenId ID of the token representing the server.
     * @param serverId ID of the server.
     * @param gameName Name of the game the server belongs to.
     * @param developer Developer or entity responsible for the server.
     */
    function _setServerInfo(
        uint256 tokenId,
        string memory serverId,
        string memory gameName,
        string memory developer
    ) internal {
        serverInfoByTokenId[tokenId] = ServerInfo({
        serverId: serverId,
        gameName: gameName,
        developer: developer
        });
    }

    /**
     * @dev Override for handling token transfers.
     */
    function _beforeTokenTransfer(
        address from,
        address to,
        uint256 tokenId,
        uint256 batchSize
    ) internal override(ERC721Upgradeable, ERC721EnumerableUpgradeable){
        super._beforeTokenTransfer(from, to, tokenId, batchSize);
    }

    /**
     * @dev Deletes the server information for a token.
     * @param tokenId ID of the token to delete information for.
     */
    function _deleteServerInfo(uint256 tokenId) private {
        delete serverInfoByTokenId[tokenId];
        emit ServerInfoDeleted(tokenId);
    }

    /**
     * @dev Burns a token and deletes its server information.
     * @param tokenId ID of the token to burn.
     */
    function _burn(uint256 tokenId) internal override(ERC721Upgradeable, ERC721URIStorageUpgradeable) {
        super._burn(tokenId);
        _deleteServerInfo(tokenId);
    }
}


