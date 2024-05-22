// SPDX-License-Identifier: MIT
pragma solidity >=0.7.0 <0.9.0;

contract Padded {
    event Test0(bytes data);

    function emitEvent() public {
        bytes memory myData = hex"010203"; 
        bytes32 eventSignature = keccak256("Test0(bytes)"); 
        assembly {
            let dataPtr := add(myData, 32)
            let dataLength := mload(myData)
            log1(dataPtr, dataLength, eventSignature)
        }
    }
}