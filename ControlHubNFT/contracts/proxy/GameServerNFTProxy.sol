// SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import "../../../../contracts/openzeppelin-contracts/proxy/transparent/TransparentUpgradeableProxy.sol";

/**
 * @title GameServerNFTProxy
 * @dev This contract serves as an upgradeable proxy for managing GameServerNFT or other related contracts.
 *      It enables upgrading the implementation of the logic contract while retaining the same proxy address.
 *      Admins can upgrade the logic contract via this proxy to ensure flexible and maintainable deployments.
 */
contract GameServerNFTProxy is TransparentUpgradeableProxy {

    /**
     * @param _logic GameServerNFT or PlayDropsMarket implementation contract address
     * @param _admin Admin address that can upgrade the implementation
     * @param _data Initialization data for the contract
     */
    constructor(
        address _logic,
        address _admin,
        bytes memory _data
    ) TransparentUpgradeableProxy(_logic, _admin, _data) {}
}