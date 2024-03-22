// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract Onboarding {
    string public message;

    constructor() {
        message = "Welcome to the team!";
    }

    function setMessage(string memory _message) public {
        message = _message;
    }

    function getMessage() public view returns (string memory) {
        return message;
    }
}