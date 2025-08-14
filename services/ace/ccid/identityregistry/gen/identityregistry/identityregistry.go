// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package identityregistry

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// IdentityregistryMetaData contains all meta data concerning the Identityregistry contract.
var IdentityregistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"attachPolicyEngine\",\"inputs\":[{\"name\":\"policyEngine\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"clearContext\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getAccounts\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getContext\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getIdentity\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPolicyEngine\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"policyEngine\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerIdentities\",\"inputs\":[{\"name\":\"ccids\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"accounts\",\"type\":\"address[]\",\"internalType\":\"address[]\"},{\"name\":\"context\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registerIdentity\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"context\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"removeIdentity\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"context\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setContext\",\"inputs\":[{\"name\":\"context\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"IdentityRegistered\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"IdentityRemoved\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PolicyEngineAttached\",\"inputs\":[{\"name\":\"policyEngine\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"IdentityAlreadyRegistered\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"IdentityNotFound\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"InvalidConfiguration\",\"inputs\":[{\"name\":\"errorReason\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"PolicyEngineUndefined\",\"inputs\":[]}]",
}

// IdentityregistryABI is the input ABI used to generate the binding from.
// Deprecated: Use IdentityregistryMetaData.ABI instead.
var IdentityregistryABI = IdentityregistryMetaData.ABI

// Identityregistry is an auto generated Go binding around an Ethereum contract.
type Identityregistry struct {
	IdentityregistryCaller     // Read-only binding to the contract
	IdentityregistryTransactor // Write-only binding to the contract
	IdentityregistryFilterer   // Log filterer for contract events
}

// IdentityregistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type IdentityregistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IdentityregistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IdentityregistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IdentityregistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IdentityregistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IdentityregistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IdentityregistrySession struct {
	Contract     *Identityregistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IdentityregistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IdentityregistryCallerSession struct {
	Contract *IdentityregistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// IdentityregistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IdentityregistryTransactorSession struct {
	Contract     *IdentityregistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// IdentityregistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type IdentityregistryRaw struct {
	Contract *Identityregistry // Generic contract binding to access the raw methods on
}

// IdentityregistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IdentityregistryCallerRaw struct {
	Contract *IdentityregistryCaller // Generic read-only contract binding to access the raw methods on
}

// IdentityregistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IdentityregistryTransactorRaw struct {
	Contract *IdentityregistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIdentityregistry creates a new instance of Identityregistry, bound to a specific deployed contract.
func NewIdentityregistry(address common.Address, backend bind.ContractBackend) (*Identityregistry, error) {
	contract, err := bindIdentityregistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Identityregistry{IdentityregistryCaller: IdentityregistryCaller{contract: contract}, IdentityregistryTransactor: IdentityregistryTransactor{contract: contract}, IdentityregistryFilterer: IdentityregistryFilterer{contract: contract}}, nil
}

// NewIdentityregistryCaller creates a new read-only instance of Identityregistry, bound to a specific deployed contract.
func NewIdentityregistryCaller(address common.Address, caller bind.ContractCaller) (*IdentityregistryCaller, error) {
	contract, err := bindIdentityregistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IdentityregistryCaller{contract: contract}, nil
}

// NewIdentityregistryTransactor creates a new write-only instance of Identityregistry, bound to a specific deployed contract.
func NewIdentityregistryTransactor(address common.Address, transactor bind.ContractTransactor) (*IdentityregistryTransactor, error) {
	contract, err := bindIdentityregistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IdentityregistryTransactor{contract: contract}, nil
}

// NewIdentityregistryFilterer creates a new log filterer instance of Identityregistry, bound to a specific deployed contract.
func NewIdentityregistryFilterer(address common.Address, filterer bind.ContractFilterer) (*IdentityregistryFilterer, error) {
	contract, err := bindIdentityregistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IdentityregistryFilterer{contract: contract}, nil
}

// bindIdentityregistry binds a generic wrapper to an already deployed contract.
func bindIdentityregistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IdentityregistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Identityregistry *IdentityregistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Identityregistry.Contract.IdentityregistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Identityregistry *IdentityregistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Identityregistry.Contract.IdentityregistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Identityregistry *IdentityregistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Identityregistry.Contract.IdentityregistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Identityregistry *IdentityregistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Identityregistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Identityregistry *IdentityregistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Identityregistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Identityregistry *IdentityregistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Identityregistry.Contract.contract.Transact(opts, method, params...)
}

// GetAccounts is a free data retrieval call binding the contract method 0xd7c89f5b.
//
// Solidity: function getAccounts(bytes32 ccid) view returns(address[])
func (_Identityregistry *IdentityregistryCaller) GetAccounts(opts *bind.CallOpts, ccid [32]byte) ([]common.Address, error) {
	var out []interface{}
	err := _Identityregistry.contract.Call(opts, &out, "getAccounts", ccid)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetAccounts is a free data retrieval call binding the contract method 0xd7c89f5b.
//
// Solidity: function getAccounts(bytes32 ccid) view returns(address[])
func (_Identityregistry *IdentityregistrySession) GetAccounts(ccid [32]byte) ([]common.Address, error) {
	return _Identityregistry.Contract.GetAccounts(&_Identityregistry.CallOpts, ccid)
}

// GetAccounts is a free data retrieval call binding the contract method 0xd7c89f5b.
//
// Solidity: function getAccounts(bytes32 ccid) view returns(address[])
func (_Identityregistry *IdentityregistryCallerSession) GetAccounts(ccid [32]byte) ([]common.Address, error) {
	return _Identityregistry.Contract.GetAccounts(&_Identityregistry.CallOpts, ccid)
}

// GetContext is a free data retrieval call binding the contract method 0x127f0f07.
//
// Solidity: function getContext() view returns(bytes)
func (_Identityregistry *IdentityregistryCaller) GetContext(opts *bind.CallOpts) ([]byte, error) {
	var out []interface{}
	err := _Identityregistry.contract.Call(opts, &out, "getContext")

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetContext is a free data retrieval call binding the contract method 0x127f0f07.
//
// Solidity: function getContext() view returns(bytes)
func (_Identityregistry *IdentityregistrySession) GetContext() ([]byte, error) {
	return _Identityregistry.Contract.GetContext(&_Identityregistry.CallOpts)
}

// GetContext is a free data retrieval call binding the contract method 0x127f0f07.
//
// Solidity: function getContext() view returns(bytes)
func (_Identityregistry *IdentityregistryCallerSession) GetContext() ([]byte, error) {
	return _Identityregistry.Contract.GetContext(&_Identityregistry.CallOpts)
}

// GetIdentity is a free data retrieval call binding the contract method 0x2fea7b81.
//
// Solidity: function getIdentity(address account) view returns(bytes32)
func (_Identityregistry *IdentityregistryCaller) GetIdentity(opts *bind.CallOpts, account common.Address) ([32]byte, error) {
	var out []interface{}
	err := _Identityregistry.contract.Call(opts, &out, "getIdentity", account)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetIdentity is a free data retrieval call binding the contract method 0x2fea7b81.
//
// Solidity: function getIdentity(address account) view returns(bytes32)
func (_Identityregistry *IdentityregistrySession) GetIdentity(account common.Address) ([32]byte, error) {
	return _Identityregistry.Contract.GetIdentity(&_Identityregistry.CallOpts, account)
}

// GetIdentity is a free data retrieval call binding the contract method 0x2fea7b81.
//
// Solidity: function getIdentity(address account) view returns(bytes32)
func (_Identityregistry *IdentityregistryCallerSession) GetIdentity(account common.Address) ([32]byte, error) {
	return _Identityregistry.Contract.GetIdentity(&_Identityregistry.CallOpts, account)
}

// GetPolicyEngine is a free data retrieval call binding the contract method 0x68317312.
//
// Solidity: function getPolicyEngine() view returns(address)
func (_Identityregistry *IdentityregistryCaller) GetPolicyEngine(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Identityregistry.contract.Call(opts, &out, "getPolicyEngine")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetPolicyEngine is a free data retrieval call binding the contract method 0x68317312.
//
// Solidity: function getPolicyEngine() view returns(address)
func (_Identityregistry *IdentityregistrySession) GetPolicyEngine() (common.Address, error) {
	return _Identityregistry.Contract.GetPolicyEngine(&_Identityregistry.CallOpts)
}

// GetPolicyEngine is a free data retrieval call binding the contract method 0x68317312.
//
// Solidity: function getPolicyEngine() view returns(address)
func (_Identityregistry *IdentityregistryCallerSession) GetPolicyEngine() (common.Address, error) {
	return _Identityregistry.Contract.GetPolicyEngine(&_Identityregistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Identityregistry *IdentityregistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Identityregistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Identityregistry *IdentityregistrySession) Owner() (common.Address, error) {
	return _Identityregistry.Contract.Owner(&_Identityregistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Identityregistry *IdentityregistryCallerSession) Owner() (common.Address, error) {
	return _Identityregistry.Contract.Owner(&_Identityregistry.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Identityregistry *IdentityregistryCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Identityregistry.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Identityregistry *IdentityregistrySession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Identityregistry.Contract.SupportsInterface(&_Identityregistry.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Identityregistry *IdentityregistryCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Identityregistry.Contract.SupportsInterface(&_Identityregistry.CallOpts, interfaceId)
}

// AttachPolicyEngine is a paid mutator transaction binding the contract method 0xf7332e86.
//
// Solidity: function attachPolicyEngine(address policyEngine) returns()
func (_Identityregistry *IdentityregistryTransactor) AttachPolicyEngine(opts *bind.TransactOpts, policyEngine common.Address) (*types.Transaction, error) {
	return _Identityregistry.contract.Transact(opts, "attachPolicyEngine", policyEngine)
}

// AttachPolicyEngine is a paid mutator transaction binding the contract method 0xf7332e86.
//
// Solidity: function attachPolicyEngine(address policyEngine) returns()
func (_Identityregistry *IdentityregistrySession) AttachPolicyEngine(policyEngine common.Address) (*types.Transaction, error) {
	return _Identityregistry.Contract.AttachPolicyEngine(&_Identityregistry.TransactOpts, policyEngine)
}

// AttachPolicyEngine is a paid mutator transaction binding the contract method 0xf7332e86.
//
// Solidity: function attachPolicyEngine(address policyEngine) returns()
func (_Identityregistry *IdentityregistryTransactorSession) AttachPolicyEngine(policyEngine common.Address) (*types.Transaction, error) {
	return _Identityregistry.Contract.AttachPolicyEngine(&_Identityregistry.TransactOpts, policyEngine)
}

// ClearContext is a paid mutator transaction binding the contract method 0xdb99bddd.
//
// Solidity: function clearContext() returns()
func (_Identityregistry *IdentityregistryTransactor) ClearContext(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Identityregistry.contract.Transact(opts, "clearContext")
}

// ClearContext is a paid mutator transaction binding the contract method 0xdb99bddd.
//
// Solidity: function clearContext() returns()
func (_Identityregistry *IdentityregistrySession) ClearContext() (*types.Transaction, error) {
	return _Identityregistry.Contract.ClearContext(&_Identityregistry.TransactOpts)
}

// ClearContext is a paid mutator transaction binding the contract method 0xdb99bddd.
//
// Solidity: function clearContext() returns()
func (_Identityregistry *IdentityregistryTransactorSession) ClearContext() (*types.Transaction, error) {
	return _Identityregistry.Contract.ClearContext(&_Identityregistry.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address policyEngine) returns()
func (_Identityregistry *IdentityregistryTransactor) Initialize(opts *bind.TransactOpts, policyEngine common.Address) (*types.Transaction, error) {
	return _Identityregistry.contract.Transact(opts, "initialize", policyEngine)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address policyEngine) returns()
func (_Identityregistry *IdentityregistrySession) Initialize(policyEngine common.Address) (*types.Transaction, error) {
	return _Identityregistry.Contract.Initialize(&_Identityregistry.TransactOpts, policyEngine)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address policyEngine) returns()
func (_Identityregistry *IdentityregistryTransactorSession) Initialize(policyEngine common.Address) (*types.Transaction, error) {
	return _Identityregistry.Contract.Initialize(&_Identityregistry.TransactOpts, policyEngine)
}

// RegisterIdentities is a paid mutator transaction binding the contract method 0x2694f92f.
//
// Solidity: function registerIdentities(bytes32[] ccids, address[] accounts, bytes context) returns()
func (_Identityregistry *IdentityregistryTransactor) RegisterIdentities(opts *bind.TransactOpts, ccids [][32]byte, accounts []common.Address, context []byte) (*types.Transaction, error) {
	return _Identityregistry.contract.Transact(opts, "registerIdentities", ccids, accounts, context)
}

// RegisterIdentities is a paid mutator transaction binding the contract method 0x2694f92f.
//
// Solidity: function registerIdentities(bytes32[] ccids, address[] accounts, bytes context) returns()
func (_Identityregistry *IdentityregistrySession) RegisterIdentities(ccids [][32]byte, accounts []common.Address, context []byte) (*types.Transaction, error) {
	return _Identityregistry.Contract.RegisterIdentities(&_Identityregistry.TransactOpts, ccids, accounts, context)
}

// RegisterIdentities is a paid mutator transaction binding the contract method 0x2694f92f.
//
// Solidity: function registerIdentities(bytes32[] ccids, address[] accounts, bytes context) returns()
func (_Identityregistry *IdentityregistryTransactorSession) RegisterIdentities(ccids [][32]byte, accounts []common.Address, context []byte) (*types.Transaction, error) {
	return _Identityregistry.Contract.RegisterIdentities(&_Identityregistry.TransactOpts, ccids, accounts, context)
}

// RegisterIdentity is a paid mutator transaction binding the contract method 0xf176f9a1.
//
// Solidity: function registerIdentity(bytes32 ccid, address account, bytes context) returns()
func (_Identityregistry *IdentityregistryTransactor) RegisterIdentity(opts *bind.TransactOpts, ccid [32]byte, account common.Address, context []byte) (*types.Transaction, error) {
	return _Identityregistry.contract.Transact(opts, "registerIdentity", ccid, account, context)
}

// RegisterIdentity is a paid mutator transaction binding the contract method 0xf176f9a1.
//
// Solidity: function registerIdentity(bytes32 ccid, address account, bytes context) returns()
func (_Identityregistry *IdentityregistrySession) RegisterIdentity(ccid [32]byte, account common.Address, context []byte) (*types.Transaction, error) {
	return _Identityregistry.Contract.RegisterIdentity(&_Identityregistry.TransactOpts, ccid, account, context)
}

// RegisterIdentity is a paid mutator transaction binding the contract method 0xf176f9a1.
//
// Solidity: function registerIdentity(bytes32 ccid, address account, bytes context) returns()
func (_Identityregistry *IdentityregistryTransactorSession) RegisterIdentity(ccid [32]byte, account common.Address, context []byte) (*types.Transaction, error) {
	return _Identityregistry.Contract.RegisterIdentity(&_Identityregistry.TransactOpts, ccid, account, context)
}

// RemoveIdentity is a paid mutator transaction binding the contract method 0x322b93d6.
//
// Solidity: function removeIdentity(bytes32 ccid, address account, bytes context) returns()
func (_Identityregistry *IdentityregistryTransactor) RemoveIdentity(opts *bind.TransactOpts, ccid [32]byte, account common.Address, context []byte) (*types.Transaction, error) {
	return _Identityregistry.contract.Transact(opts, "removeIdentity", ccid, account, context)
}

// RemoveIdentity is a paid mutator transaction binding the contract method 0x322b93d6.
//
// Solidity: function removeIdentity(bytes32 ccid, address account, bytes context) returns()
func (_Identityregistry *IdentityregistrySession) RemoveIdentity(ccid [32]byte, account common.Address, context []byte) (*types.Transaction, error) {
	return _Identityregistry.Contract.RemoveIdentity(&_Identityregistry.TransactOpts, ccid, account, context)
}

// RemoveIdentity is a paid mutator transaction binding the contract method 0x322b93d6.
//
// Solidity: function removeIdentity(bytes32 ccid, address account, bytes context) returns()
func (_Identityregistry *IdentityregistryTransactorSession) RemoveIdentity(ccid [32]byte, account common.Address, context []byte) (*types.Transaction, error) {
	return _Identityregistry.Contract.RemoveIdentity(&_Identityregistry.TransactOpts, ccid, account, context)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Identityregistry *IdentityregistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Identityregistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Identityregistry *IdentityregistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _Identityregistry.Contract.RenounceOwnership(&_Identityregistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Identityregistry *IdentityregistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Identityregistry.Contract.RenounceOwnership(&_Identityregistry.TransactOpts)
}

// SetContext is a paid mutator transaction binding the contract method 0x2f7482be.
//
// Solidity: function setContext(bytes context) returns()
func (_Identityregistry *IdentityregistryTransactor) SetContext(opts *bind.TransactOpts, context []byte) (*types.Transaction, error) {
	return _Identityregistry.contract.Transact(opts, "setContext", context)
}

// SetContext is a paid mutator transaction binding the contract method 0x2f7482be.
//
// Solidity: function setContext(bytes context) returns()
func (_Identityregistry *IdentityregistrySession) SetContext(context []byte) (*types.Transaction, error) {
	return _Identityregistry.Contract.SetContext(&_Identityregistry.TransactOpts, context)
}

// SetContext is a paid mutator transaction binding the contract method 0x2f7482be.
//
// Solidity: function setContext(bytes context) returns()
func (_Identityregistry *IdentityregistryTransactorSession) SetContext(context []byte) (*types.Transaction, error) {
	return _Identityregistry.Contract.SetContext(&_Identityregistry.TransactOpts, context)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Identityregistry *IdentityregistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Identityregistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Identityregistry *IdentityregistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Identityregistry.Contract.TransferOwnership(&_Identityregistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Identityregistry *IdentityregistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Identityregistry.Contract.TransferOwnership(&_Identityregistry.TransactOpts, newOwner)
}

// IdentityregistryIdentityRegisteredIterator is returned from FilterIdentityRegistered and is used to iterate over the raw logs and unpacked data for IdentityRegistered events raised by the Identityregistry contract.
type IdentityregistryIdentityRegisteredIterator struct {
	Event *IdentityregistryIdentityRegistered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IdentityregistryIdentityRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityregistryIdentityRegistered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IdentityregistryIdentityRegistered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IdentityregistryIdentityRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityregistryIdentityRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityregistryIdentityRegistered represents a IdentityRegistered event raised by the Identityregistry contract.
type IdentityregistryIdentityRegistered struct {
	Ccid    [32]byte
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterIdentityRegistered is a free log retrieval operation binding the contract event 0xd0c90c8049cdcd414cc4e521cc98d9c70f84ba695cde160691c0aebc48df058a.
//
// Solidity: event IdentityRegistered(bytes32 indexed ccid, address indexed account)
func (_Identityregistry *IdentityregistryFilterer) FilterIdentityRegistered(opts *bind.FilterOpts, ccid [][32]byte, account []common.Address) (*IdentityregistryIdentityRegisteredIterator, error) {

	var ccidRule []interface{}
	for _, ccidItem := range ccid {
		ccidRule = append(ccidRule, ccidItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Identityregistry.contract.FilterLogs(opts, "IdentityRegistered", ccidRule, accountRule)
	if err != nil {
		return nil, err
	}
	return &IdentityregistryIdentityRegisteredIterator{contract: _Identityregistry.contract, event: "IdentityRegistered", logs: logs, sub: sub}, nil
}

// WatchIdentityRegistered is a free log subscription operation binding the contract event 0xd0c90c8049cdcd414cc4e521cc98d9c70f84ba695cde160691c0aebc48df058a.
//
// Solidity: event IdentityRegistered(bytes32 indexed ccid, address indexed account)
func (_Identityregistry *IdentityregistryFilterer) WatchIdentityRegistered(opts *bind.WatchOpts, sink chan<- *IdentityregistryIdentityRegistered, ccid [][32]byte, account []common.Address) (event.Subscription, error) {

	var ccidRule []interface{}
	for _, ccidItem := range ccid {
		ccidRule = append(ccidRule, ccidItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Identityregistry.contract.WatchLogs(opts, "IdentityRegistered", ccidRule, accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityregistryIdentityRegistered)
				if err := _Identityregistry.contract.UnpackLog(event, "IdentityRegistered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseIdentityRegistered is a log parse operation binding the contract event 0xd0c90c8049cdcd414cc4e521cc98d9c70f84ba695cde160691c0aebc48df058a.
//
// Solidity: event IdentityRegistered(bytes32 indexed ccid, address indexed account)
func (_Identityregistry *IdentityregistryFilterer) ParseIdentityRegistered(log types.Log) (*IdentityregistryIdentityRegistered, error) {
	event := new(IdentityregistryIdentityRegistered)
	if err := _Identityregistry.contract.UnpackLog(event, "IdentityRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IdentityregistryIdentityRemovedIterator is returned from FilterIdentityRemoved and is used to iterate over the raw logs and unpacked data for IdentityRemoved events raised by the Identityregistry contract.
type IdentityregistryIdentityRemovedIterator struct {
	Event *IdentityregistryIdentityRemoved // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IdentityregistryIdentityRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityregistryIdentityRemoved)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IdentityregistryIdentityRemoved)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IdentityregistryIdentityRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityregistryIdentityRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityregistryIdentityRemoved represents a IdentityRemoved event raised by the Identityregistry contract.
type IdentityregistryIdentityRemoved struct {
	Ccid    [32]byte
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterIdentityRemoved is a free log retrieval operation binding the contract event 0x8a54ce55b942a912ca47afe2d31670056cd8dadf54f94dced757178a992fe79e.
//
// Solidity: event IdentityRemoved(bytes32 indexed ccid, address indexed account)
func (_Identityregistry *IdentityregistryFilterer) FilterIdentityRemoved(opts *bind.FilterOpts, ccid [][32]byte, account []common.Address) (*IdentityregistryIdentityRemovedIterator, error) {

	var ccidRule []interface{}
	for _, ccidItem := range ccid {
		ccidRule = append(ccidRule, ccidItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Identityregistry.contract.FilterLogs(opts, "IdentityRemoved", ccidRule, accountRule)
	if err != nil {
		return nil, err
	}
	return &IdentityregistryIdentityRemovedIterator{contract: _Identityregistry.contract, event: "IdentityRemoved", logs: logs, sub: sub}, nil
}

// WatchIdentityRemoved is a free log subscription operation binding the contract event 0x8a54ce55b942a912ca47afe2d31670056cd8dadf54f94dced757178a992fe79e.
//
// Solidity: event IdentityRemoved(bytes32 indexed ccid, address indexed account)
func (_Identityregistry *IdentityregistryFilterer) WatchIdentityRemoved(opts *bind.WatchOpts, sink chan<- *IdentityregistryIdentityRemoved, ccid [][32]byte, account []common.Address) (event.Subscription, error) {

	var ccidRule []interface{}
	for _, ccidItem := range ccid {
		ccidRule = append(ccidRule, ccidItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Identityregistry.contract.WatchLogs(opts, "IdentityRemoved", ccidRule, accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityregistryIdentityRemoved)
				if err := _Identityregistry.contract.UnpackLog(event, "IdentityRemoved", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseIdentityRemoved is a log parse operation binding the contract event 0x8a54ce55b942a912ca47afe2d31670056cd8dadf54f94dced757178a992fe79e.
//
// Solidity: event IdentityRemoved(bytes32 indexed ccid, address indexed account)
func (_Identityregistry *IdentityregistryFilterer) ParseIdentityRemoved(log types.Log) (*IdentityregistryIdentityRemoved, error) {
	event := new(IdentityregistryIdentityRemoved)
	if err := _Identityregistry.contract.UnpackLog(event, "IdentityRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IdentityregistryInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Identityregistry contract.
type IdentityregistryInitializedIterator struct {
	Event *IdentityregistryInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IdentityregistryInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityregistryInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IdentityregistryInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IdentityregistryInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityregistryInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityregistryInitialized represents a Initialized event raised by the Identityregistry contract.
type IdentityregistryInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Identityregistry *IdentityregistryFilterer) FilterInitialized(opts *bind.FilterOpts) (*IdentityregistryInitializedIterator, error) {

	logs, sub, err := _Identityregistry.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &IdentityregistryInitializedIterator{contract: _Identityregistry.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Identityregistry *IdentityregistryFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *IdentityregistryInitialized) (event.Subscription, error) {

	logs, sub, err := _Identityregistry.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityregistryInitialized)
				if err := _Identityregistry.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Identityregistry *IdentityregistryFilterer) ParseInitialized(log types.Log) (*IdentityregistryInitialized, error) {
	event := new(IdentityregistryInitialized)
	if err := _Identityregistry.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IdentityregistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Identityregistry contract.
type IdentityregistryOwnershipTransferredIterator struct {
	Event *IdentityregistryOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IdentityregistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityregistryOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IdentityregistryOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IdentityregistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityregistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityregistryOwnershipTransferred represents a OwnershipTransferred event raised by the Identityregistry contract.
type IdentityregistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Identityregistry *IdentityregistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*IdentityregistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Identityregistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &IdentityregistryOwnershipTransferredIterator{contract: _Identityregistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Identityregistry *IdentityregistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *IdentityregistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Identityregistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityregistryOwnershipTransferred)
				if err := _Identityregistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Identityregistry *IdentityregistryFilterer) ParseOwnershipTransferred(log types.Log) (*IdentityregistryOwnershipTransferred, error) {
	event := new(IdentityregistryOwnershipTransferred)
	if err := _Identityregistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IdentityregistryPolicyEngineAttachedIterator is returned from FilterPolicyEngineAttached and is used to iterate over the raw logs and unpacked data for PolicyEngineAttached events raised by the Identityregistry contract.
type IdentityregistryPolicyEngineAttachedIterator struct {
	Event *IdentityregistryPolicyEngineAttached // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IdentityregistryPolicyEngineAttachedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityregistryPolicyEngineAttached)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IdentityregistryPolicyEngineAttached)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IdentityregistryPolicyEngineAttachedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityregistryPolicyEngineAttachedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityregistryPolicyEngineAttached represents a PolicyEngineAttached event raised by the Identityregistry contract.
type IdentityregistryPolicyEngineAttached struct {
	PolicyEngine common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterPolicyEngineAttached is a free log retrieval operation binding the contract event 0x57d241970863a27bedbf58b705b45a0b267f76f9a3a7fd432e217a37e4173fac.
//
// Solidity: event PolicyEngineAttached(address indexed policyEngine)
func (_Identityregistry *IdentityregistryFilterer) FilterPolicyEngineAttached(opts *bind.FilterOpts, policyEngine []common.Address) (*IdentityregistryPolicyEngineAttachedIterator, error) {

	var policyEngineRule []interface{}
	for _, policyEngineItem := range policyEngine {
		policyEngineRule = append(policyEngineRule, policyEngineItem)
	}

	logs, sub, err := _Identityregistry.contract.FilterLogs(opts, "PolicyEngineAttached", policyEngineRule)
	if err != nil {
		return nil, err
	}
	return &IdentityregistryPolicyEngineAttachedIterator{contract: _Identityregistry.contract, event: "PolicyEngineAttached", logs: logs, sub: sub}, nil
}

// WatchPolicyEngineAttached is a free log subscription operation binding the contract event 0x57d241970863a27bedbf58b705b45a0b267f76f9a3a7fd432e217a37e4173fac.
//
// Solidity: event PolicyEngineAttached(address indexed policyEngine)
func (_Identityregistry *IdentityregistryFilterer) WatchPolicyEngineAttached(opts *bind.WatchOpts, sink chan<- *IdentityregistryPolicyEngineAttached, policyEngine []common.Address) (event.Subscription, error) {

	var policyEngineRule []interface{}
	for _, policyEngineItem := range policyEngine {
		policyEngineRule = append(policyEngineRule, policyEngineItem)
	}

	logs, sub, err := _Identityregistry.contract.WatchLogs(opts, "PolicyEngineAttached", policyEngineRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityregistryPolicyEngineAttached)
				if err := _Identityregistry.contract.UnpackLog(event, "PolicyEngineAttached", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePolicyEngineAttached is a log parse operation binding the contract event 0x57d241970863a27bedbf58b705b45a0b267f76f9a3a7fd432e217a37e4173fac.
//
// Solidity: event PolicyEngineAttached(address indexed policyEngine)
func (_Identityregistry *IdentityregistryFilterer) ParsePolicyEngineAttached(log types.Log) (*IdentityregistryPolicyEngineAttached, error) {
	event := new(IdentityregistryPolicyEngineAttached)
	if err := _Identityregistry.contract.UnpackLog(event, "PolicyEngineAttached", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
