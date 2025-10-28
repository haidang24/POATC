// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title DataTraceability
 * @dev Smart contract for data traceability and provenance tracking on POATC blockchain
 */
contract DataTraceability {
    
    struct DataRecord {
        uint256 id;
        string dataHash;        // IPFS hash or SHA256 of data
        string dataType;        // Type of data (e.g., "product", "document", "transaction")
        string description;     // Human-readable description
        address creator;        // Address that created this record
        uint256 timestamp;      // Block timestamp
        string metadata;        // JSON metadata
        bool verified;          // Verification status
        address verifiedBy;     // Address that verified the record
        uint256 verifiedAt;     // Verification timestamp
    }
    
    struct TraceStep {
        uint256 recordId;
        string action;          // Action performed (e.g., "created", "transferred", "updated")
        address actor;          // Address performing the action
        uint256 timestamp;      // When action occurred
        string details;         // Additional details
    }
    
    // State variables
    mapping(uint256 => DataRecord) public records;
    mapping(uint256 => TraceStep[]) public traceHistory;
    mapping(address => uint256[]) public userRecords;
    mapping(bytes32 => bool) public usedHashes;
    
    uint256 public recordCount;
    address public owner;
    
    // Events
    event RecordCreated(
        uint256 indexed recordId,
        string dataHash,
        string dataType,
        address indexed creator,
        uint256 timestamp
    );
    
    event RecordVerified(
        uint256 indexed recordId,
        address indexed verifier,
        uint256 timestamp
    );
    
    event TraceStepAdded(
        uint256 indexed recordId,
        string action,
        address indexed actor,
        uint256 timestamp
    );
    
    event OwnershipTransferred(
        address indexed previousOwner,
        address indexed newOwner
    );
    
    // Modifiers
    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this");
        _;
    }
    
    modifier recordExists(uint256 _recordId) {
        require(_recordId > 0 && _recordId <= recordCount, "Record does not exist");
        _;
    }
    
    modifier onlyRecordCreator(uint256 _recordId) {
        require(records[_recordId].creator == msg.sender, "Only record creator can call this");
        _;
    }
    
    constructor() {
        owner = msg.sender;
        recordCount = 0;
    }
    
    /**
     * @dev Create a new data record
     */
    function createRecord(
        string memory _dataHash,
        string memory _dataType,
        string memory _description,
        string memory _metadata
    ) public returns (uint256) {
        require(bytes(_dataHash).length > 0, "Data hash cannot be empty");
        require(bytes(_dataType).length > 0, "Data type cannot be empty");
        
        bytes32 hashKey = keccak256(abi.encodePacked(_dataHash));
        require(!usedHashes[hashKey], "Data hash already exists");
        
        recordCount++;
        uint256 newRecordId = recordCount;
        
        records[newRecordId] = DataRecord({
            id: newRecordId,
            dataHash: _dataHash,
            dataType: _dataType,
            description: _description,
            creator: msg.sender,
            timestamp: block.timestamp,
            metadata: _metadata,
            verified: false,
            verifiedBy: address(0),
            verifiedAt: 0
        });
        
        usedHashes[hashKey] = true;
        userRecords[msg.sender].push(newRecordId);
        
        // Add initial trace step
        _addTraceStep(newRecordId, "created", "Record created on blockchain");
        
        emit RecordCreated(newRecordId, _dataHash, _dataType, msg.sender, block.timestamp);
        
        return newRecordId;
    }
    
    /**
     * @dev Verify a data record (only owner can verify)
     */
    function verifyRecord(uint256 _recordId) public onlyOwner recordExists(_recordId) {
        require(!records[_recordId].verified, "Record already verified");
        
        records[_recordId].verified = true;
        records[_recordId].verifiedBy = msg.sender;
        records[_recordId].verifiedAt = block.timestamp;
        
        _addTraceStep(_recordId, "verified", "Record verified by contract owner");
        
        emit RecordVerified(_recordId, msg.sender, block.timestamp);
    }
    
    /**
     * @dev Add a trace step to record history
     */
    function addTraceStep(
        uint256 _recordId,
        string memory _action,
        string memory _details
    ) public recordExists(_recordId) onlyRecordCreator(_recordId) {
        _addTraceStep(_recordId, _action, _details);
    }
    
    /**
     * @dev Internal function to add trace step
     */
    function _addTraceStep(
        uint256 _recordId,
        string memory _action,
        string memory _details
    ) internal {
        traceHistory[_recordId].push(TraceStep({
            recordId: _recordId,
            action: _action,
            actor: msg.sender,
            timestamp: block.timestamp,
            details: _details
        }));
        
        emit TraceStepAdded(_recordId, _action, msg.sender, block.timestamp);
    }
    
    /**
     * @dev Get record details
     */
    function getRecord(uint256 _recordId) public view recordExists(_recordId) returns (
        uint256 id,
        string memory dataHash,
        string memory dataType,
        string memory description,
        address creator,
        uint256 timestamp,
        string memory metadata,
        bool verified,
        address verifiedBy,
        uint256 verifiedAt
    ) {
        DataRecord memory record = records[_recordId];
        return (
            record.id,
            record.dataHash,
            record.dataType,
            record.description,
            record.creator,
            record.timestamp,
            record.metadata,
            record.verified,
            record.verifiedBy,
            record.verifiedAt
        );
    }
    
    /**
     * @dev Get trace history for a record
     */
    function getTraceHistory(uint256 _recordId) public view recordExists(_recordId) returns (TraceStep[] memory) {
        return traceHistory[_recordId];
    }
    
    /**
     * @dev Get records created by a user
     */
    function getUserRecords(address _user) public view returns (uint256[] memory) {
        return userRecords[_user];
    }
    
    /**
     * @dev Get trace step count for a record
     */
    function getTraceStepCount(uint256 _recordId) public view recordExists(_recordId) returns (uint256) {
        return traceHistory[_recordId].length;
    }
    
    /**
     * @dev Transfer ownership
     */
    function transferOwnership(address newOwner) public onlyOwner {
        require(newOwner != address(0), "New owner cannot be zero address");
        address oldOwner = owner;
        owner = newOwner;
        emit OwnershipTransferred(oldOwner, newOwner);
    }
    
    /**
     * @dev Get contract info
     */
    function getContractInfo() public view returns (
        address contractOwner,
        uint256 totalRecords,
        uint256 blockNumber,
        uint256 blockTimestamp
    ) {
        return (owner, recordCount, block.number, block.timestamp);
    }
}

