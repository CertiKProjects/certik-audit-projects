// SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import "../GameServerNftBaseV1.sol";

/**
 * @title GameServerNFT
 * @dev This contract manages minting, updating, and burning NFTs that represent game servers.
 *      It inherits from GameServerNftBase and provides additional functionalities for handling
 *      server information and token metadata.
 */
contract GameServerNFT is GameServerNftBaseV1 {

    constructor() {
        _disableInitializers();
    }

    /**
     * @dev Mints a new server NFT.
     * @param recipient Address to receive the newly minted NFT.
     * @param serverId ID of the server.
     * @param gameName Name of the game the server belongs to.
     * @param developer Developer or entity responsible for the server.
     * @param metadataURI URI pointing to the metadata for the NFT.
     */
    function mintServerNFT(
        address recipient,
        string memory serverId,
        string memory gameName,
        string memory developer,
        string memory metadataURI
    ) public nonReentrant{
        uint256 tokenId = nextTokenId;
        nextTokenId++;

        _mintServerNFTBase(recipient, tokenId, metadataURI);
        _setServerInfo(tokenId, serverId, gameName, developer);

        emit ServerMinted(recipient, tokenId, serverId, gameName);
    }

    /**
     * @dev Batch mint multiple server NFTs at once and emit events for each mint.
     * @param recipients Array of addresses that will receive the newly minted NFTs.
     * @param serverIds Array of server IDs representing the servers for each NFT.
     * @param gameNames Array of game names that the servers belong to for each NFT.
     * @param developers Array of developer names or entities responsible for each server.
     * @param metadataURIs Array of URIs pointing to the metadata for each NFT.
     */
    function batchMintServerNFT(
        address[] memory recipients,
        string[] memory serverIds,
        string[] memory gameNames,
        string[] memory developers,
        string[] memory metadataURIs
    ) external{
        require(recipients.length == serverIds.length
        && recipients.length == gameNames.length
        && recipients.length == developers.length
        && recipients.length == metadataURIs.length, "Input array length mismatch");

        for (uint256 i = 0; i < recipients.length; i++) {
            mintServerNFT(recipients[i], serverIds[i], gameNames[i], developers[i], metadataURIs[i]);
        }
    }

    /**
     * @dev Updates the token URI for a given token.
     * @param tokenId ID of the token.
     * @param metadataURI New URI to be set.
     */
    function setTokenURI(uint256 tokenId, string memory metadataURI) external {
        _setTokenURI(tokenId, metadataURI);
    }


    /**
     * @dev Returns the server information for a given token.
     * @param tokenId ID of the token representing the server.
     * @return serverId Server ID.
     */
    function getServerInfo(uint256 tokenId) external view returns (string memory) {
        ServerInfo memory info = serverInfoByTokenId[tokenId];
        return info.serverId;
    }

      /**
      * @dev Burns the specified token if called by the token owner.
      * @param tokenId ID of the token to burn.
      */
    function burnServerNFT(uint256 tokenId) external nonReentrant {
        require(_exists(tokenId), "GameServerNFT: Token does not exist.");
        require(ownerOf(tokenId) == msg.sender, "GameServerNFT: Caller is not the token owner.");
        _burn(tokenId);
    }

    /**
     * @dev Transfers the token `tokenId` from `from` to `to`, with blacklist checks.
     * @param from The current owner of the token.
     * @param to The address to transfer the token to.
     * @param tokenId The ID of the token to transfer.
     */
    function transferFrom(
        address from,
        address to,
        uint256 tokenId
    ) public virtual override isNotEntireBlacklist(blackOrWhiteList, from) isNotEntireBlacklist(blackOrWhiteList, to) {
        super.transferFrom(from, to, tokenId);
    }

    /**
     * @dev Safely transfers the token `tokenId` from `from` to `to`.
     * @param from Current owner of the token.
     * @param to Address to transfer the token to.
     * @param tokenId ID of the token to transfer.
     * @param data Additional data with no specified format.
     */
    function safeTransferFrom(
        address from,
        address to,
        uint256 tokenId,
        bytes memory data
    ) public override isNotEntireBlacklist(blackOrWhiteList, from) isNotEntireBlacklist(blackOrWhiteList, to) {
        super.safeTransferFrom(from, to, tokenId, data);
    }

    /**
     * @dev Safely transfers the token `tokenId` from `from` to `to`.
     * @param from Current owner of the token.
     * @param to Address to transfer the token to.
     * @param tokenId ID of the token to transfer.
     */
    function safeTransferFrom(
        address from,
        address to,
        uint256 tokenId
    ) public override isNotEntireBlacklist(blackOrWhiteList, from) isNotEntireBlacklist(blackOrWhiteList, to) {
        super.safeTransferFrom(from, to, tokenId);
    }
}


