pragma solidity ^0.8.0;

contract S{

    mapping(bytes32 => string) private _Map;
    string[] private _Things;

    function add(string calldata thing) public{
        _Things.push(thing);
    }

    function get(uint256 key) public view returns (string memory){
        return _Things[key];
    }

    function len() public view returns (uint256){
        return _Things.length;
    }

    function set(bytes32 key, string calldata value) public {
        _Map[key] = value;
    }
 
    function lookup(bytes32 key) public view returns (string memory){
        return _Map[key];
    }

}
