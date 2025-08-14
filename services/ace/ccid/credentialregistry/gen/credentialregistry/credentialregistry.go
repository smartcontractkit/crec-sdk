// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package credentialregistry

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

// ICredentialRegistryCredential is an auto generated low-level Go binding around an user-defined struct.
type ICredentialRegistryCredential struct {
	ExpiresAt      *big.Int
	CredentialData []byte
}

// CredentialregistryMetaData contains all meta data concerning the Credentialregistry contract.
var CredentialregistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"attachPolicyEngine\",\"inputs\":[{\"name\":\"policyEngine\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"clearContext\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getContext\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCredential\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structICredentialRegistry.Credential\",\"components\":[{\"name\":\"expiresAt\",\"type\":\"uint40\",\"internalType\":\"uint40\"},{\"name\":\"credentialData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCredentialTypes\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCredentials\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeIds\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structICredentialRegistry.Credential[]\",\"components\":[{\"name\":\"expiresAt\",\"type\":\"uint40\",\"internalType\":\"uint40\"},{\"name\":\"credentialData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPolicyEngine\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"policyEngine\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isCredentialExpired\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerCredential\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"expiresAt\",\"type\":\"uint40\",\"internalType\":\"uint40\"},{\"name\":\"credentialData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"context\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registerCredentials\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeIds\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"expiresAt\",\"type\":\"uint40\",\"internalType\":\"uint40\"},{\"name\":\"credentialDatas\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"},{\"name\":\"context\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"removeCredential\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"context\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renewCredential\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"expiresAt\",\"type\":\"uint40\",\"internalType\":\"uint40\"},{\"name\":\"context\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setContext\",\"inputs\":[{\"name\":\"context\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"validate\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"context\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"validateAll\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeIds\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"context\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"CredentialRegistered\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"expiresAt\",\"type\":\"uint40\",\"indexed\":false,\"internalType\":\"uint40\"},{\"name\":\"credentialData\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CredentialRemoved\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CredentialRenewed\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"expiresAt\",\"type\":\"uint40\",\"indexed\":false,\"internalType\":\"uint40\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PolicyEngineAttached\",\"inputs\":[{\"name\":\"policyEngine\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"CredentialAlreadyRegistered\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"CredentialNotFound\",\"inputs\":[{\"name\":\"ccid\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"InvalidConfiguration\",\"inputs\":[{\"name\":\"errorReason\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"PolicyEngineUndefined\",\"inputs\":[]}]",
}

// CredentialregistryABI is the input ABI used to generate the binding from.
// Deprecated: Use CredentialregistryMetaData.ABI instead.
var CredentialregistryABI = CredentialregistryMetaData.ABI

// Credentialregistry is an auto generated Go binding around an Ethereum contract.
type Credentialregistry struct {
	CredentialregistryCaller     // Read-only binding to the contract
	CredentialregistryTransactor // Write-only binding to the contract
	CredentialregistryFilterer   // Log filterer for contract events
}

// CredentialregistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type CredentialregistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CredentialregistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CredentialregistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CredentialregistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CredentialregistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CredentialregistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CredentialregistrySession struct {
	Contract     *Credentialregistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// CredentialregistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CredentialregistryCallerSession struct {
	Contract *CredentialregistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// CredentialregistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CredentialregistryTransactorSession struct {
	Contract     *CredentialregistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// CredentialregistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type CredentialregistryRaw struct {
	Contract *Credentialregistry // Generic contract binding to access the raw methods on
}

// CredentialregistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CredentialregistryCallerRaw struct {
	Contract *CredentialregistryCaller // Generic read-only contract binding to access the raw methods on
}

// CredentialregistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CredentialregistryTransactorRaw struct {
	Contract *CredentialregistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCredentialregistry creates a new instance of Credentialregistry, bound to a specific deployed contract.
func NewCredentialregistry(address common.Address, backend bind.ContractBackend) (*Credentialregistry, error) {
	contract, err := bindCredentialregistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Credentialregistry{CredentialregistryCaller: CredentialregistryCaller{contract: contract}, CredentialregistryTransactor: CredentialregistryTransactor{contract: contract}, CredentialregistryFilterer: CredentialregistryFilterer{contract: contract}}, nil
}

// NewCredentialregistryCaller creates a new read-only instance of Credentialregistry, bound to a specific deployed contract.
func NewCredentialregistryCaller(address common.Address, caller bind.ContractCaller) (*CredentialregistryCaller, error) {
	contract, err := bindCredentialregistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CredentialregistryCaller{contract: contract}, nil
}

// NewCredentialregistryTransactor creates a new write-only instance of Credentialregistry, bound to a specific deployed contract.
func NewCredentialregistryTransactor(address common.Address, transactor bind.ContractTransactor) (*CredentialregistryTransactor, error) {
	contract, err := bindCredentialregistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CredentialregistryTransactor{contract: contract}, nil
}

// NewCredentialregistryFilterer creates a new log filterer instance of Credentialregistry, bound to a specific deployed contract.
func NewCredentialregistryFilterer(address common.Address, filterer bind.ContractFilterer) (*CredentialregistryFilterer, error) {
	contract, err := bindCredentialregistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CredentialregistryFilterer{contract: contract}, nil
}

// bindCredentialregistry binds a generic wrapper to an already deployed contract.
func bindCredentialregistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := CredentialregistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Credentialregistry *CredentialregistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Credentialregistry.Contract.CredentialregistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Credentialregistry *CredentialregistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Credentialregistry.Contract.CredentialregistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Credentialregistry *CredentialregistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Credentialregistry.Contract.CredentialregistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Credentialregistry *CredentialregistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Credentialregistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Credentialregistry *CredentialregistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Credentialregistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Credentialregistry *CredentialregistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Credentialregistry.Contract.contract.Transact(opts, method, params...)
}

// GetContext is a free data retrieval call binding the contract method 0x127f0f07.
//
// Solidity: function getContext() view returns(bytes)
func (_Credentialregistry *CredentialregistryCaller) GetContext(opts *bind.CallOpts) ([]byte, error) {
	var out []interface{}
	err := _Credentialregistry.contract.Call(opts, &out, "getContext")

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetContext is a free data retrieval call binding the contract method 0x127f0f07.
//
// Solidity: function getContext() view returns(bytes)
func (_Credentialregistry *CredentialregistrySession) GetContext() ([]byte, error) {
	return _Credentialregistry.Contract.GetContext(&_Credentialregistry.CallOpts)
}

// GetContext is a free data retrieval call binding the contract method 0x127f0f07.
//
// Solidity: function getContext() view returns(bytes)
func (_Credentialregistry *CredentialregistryCallerSession) GetContext() ([]byte, error) {
	return _Credentialregistry.Contract.GetContext(&_Credentialregistry.CallOpts)
}

// GetCredential is a free data retrieval call binding the contract method 0xffa0a37d.
//
// Solidity: function getCredential(bytes32 ccid, bytes32 credentialTypeId) view returns((uint40,bytes))
func (_Credentialregistry *CredentialregistryCaller) GetCredential(opts *bind.CallOpts, ccid [32]byte, credentialTypeId [32]byte) (ICredentialRegistryCredential, error) {
	var out []interface{}
	err := _Credentialregistry.contract.Call(opts, &out, "getCredential", ccid, credentialTypeId)

	if err != nil {
		return *new(ICredentialRegistryCredential), err
	}

	out0 := *abi.ConvertType(out[0], new(ICredentialRegistryCredential)).(*ICredentialRegistryCredential)

	return out0, err

}

// GetCredential is a free data retrieval call binding the contract method 0xffa0a37d.
//
// Solidity: function getCredential(bytes32 ccid, bytes32 credentialTypeId) view returns((uint40,bytes))
func (_Credentialregistry *CredentialregistrySession) GetCredential(ccid [32]byte, credentialTypeId [32]byte) (ICredentialRegistryCredential, error) {
	return _Credentialregistry.Contract.GetCredential(&_Credentialregistry.CallOpts, ccid, credentialTypeId)
}

// GetCredential is a free data retrieval call binding the contract method 0xffa0a37d.
//
// Solidity: function getCredential(bytes32 ccid, bytes32 credentialTypeId) view returns((uint40,bytes))
func (_Credentialregistry *CredentialregistryCallerSession) GetCredential(ccid [32]byte, credentialTypeId [32]byte) (ICredentialRegistryCredential, error) {
	return _Credentialregistry.Contract.GetCredential(&_Credentialregistry.CallOpts, ccid, credentialTypeId)
}

// GetCredentialTypes is a free data retrieval call binding the contract method 0x2bd5faf5.
//
// Solidity: function getCredentialTypes(bytes32 ccid) view returns(bytes32[])
func (_Credentialregistry *CredentialregistryCaller) GetCredentialTypes(opts *bind.CallOpts, ccid [32]byte) ([][32]byte, error) {
	var out []interface{}
	err := _Credentialregistry.contract.Call(opts, &out, "getCredentialTypes", ccid)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetCredentialTypes is a free data retrieval call binding the contract method 0x2bd5faf5.
//
// Solidity: function getCredentialTypes(bytes32 ccid) view returns(bytes32[])
func (_Credentialregistry *CredentialregistrySession) GetCredentialTypes(ccid [32]byte) ([][32]byte, error) {
	return _Credentialregistry.Contract.GetCredentialTypes(&_Credentialregistry.CallOpts, ccid)
}

// GetCredentialTypes is a free data retrieval call binding the contract method 0x2bd5faf5.
//
// Solidity: function getCredentialTypes(bytes32 ccid) view returns(bytes32[])
func (_Credentialregistry *CredentialregistryCallerSession) GetCredentialTypes(ccid [32]byte) ([][32]byte, error) {
	return _Credentialregistry.Contract.GetCredentialTypes(&_Credentialregistry.CallOpts, ccid)
}

// GetCredentials is a free data retrieval call binding the contract method 0xabae51b0.
//
// Solidity: function getCredentials(bytes32 ccid, bytes32[] credentialTypeIds) view returns((uint40,bytes)[])
func (_Credentialregistry *CredentialregistryCaller) GetCredentials(opts *bind.CallOpts, ccid [32]byte, credentialTypeIds [][32]byte) ([]ICredentialRegistryCredential, error) {
	var out []interface{}
	err := _Credentialregistry.contract.Call(opts, &out, "getCredentials", ccid, credentialTypeIds)

	if err != nil {
		return *new([]ICredentialRegistryCredential), err
	}

	out0 := *abi.ConvertType(out[0], new([]ICredentialRegistryCredential)).(*[]ICredentialRegistryCredential)

	return out0, err

}

// GetCredentials is a free data retrieval call binding the contract method 0xabae51b0.
//
// Solidity: function getCredentials(bytes32 ccid, bytes32[] credentialTypeIds) view returns((uint40,bytes)[])
func (_Credentialregistry *CredentialregistrySession) GetCredentials(ccid [32]byte, credentialTypeIds [][32]byte) ([]ICredentialRegistryCredential, error) {
	return _Credentialregistry.Contract.GetCredentials(&_Credentialregistry.CallOpts, ccid, credentialTypeIds)
}

// GetCredentials is a free data retrieval call binding the contract method 0xabae51b0.
//
// Solidity: function getCredentials(bytes32 ccid, bytes32[] credentialTypeIds) view returns((uint40,bytes)[])
func (_Credentialregistry *CredentialregistryCallerSession) GetCredentials(ccid [32]byte, credentialTypeIds [][32]byte) ([]ICredentialRegistryCredential, error) {
	return _Credentialregistry.Contract.GetCredentials(&_Credentialregistry.CallOpts, ccid, credentialTypeIds)
}

// GetPolicyEngine is a free data retrieval call binding the contract method 0x68317312.
//
// Solidity: function getPolicyEngine() view returns(address)
func (_Credentialregistry *CredentialregistryCaller) GetPolicyEngine(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Credentialregistry.contract.Call(opts, &out, "getPolicyEngine")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetPolicyEngine is a free data retrieval call binding the contract method 0x68317312.
//
// Solidity: function getPolicyEngine() view returns(address)
func (_Credentialregistry *CredentialregistrySession) GetPolicyEngine() (common.Address, error) {
	return _Credentialregistry.Contract.GetPolicyEngine(&_Credentialregistry.CallOpts)
}

// GetPolicyEngine is a free data retrieval call binding the contract method 0x68317312.
//
// Solidity: function getPolicyEngine() view returns(address)
func (_Credentialregistry *CredentialregistryCallerSession) GetPolicyEngine() (common.Address, error) {
	return _Credentialregistry.Contract.GetPolicyEngine(&_Credentialregistry.CallOpts)
}

// IsCredentialExpired is a free data retrieval call binding the contract method 0xa1f5de49.
//
// Solidity: function isCredentialExpired(bytes32 ccid, bytes32 credentialTypeId) view returns(bool)
func (_Credentialregistry *CredentialregistryCaller) IsCredentialExpired(opts *bind.CallOpts, ccid [32]byte, credentialTypeId [32]byte) (bool, error) {
	var out []interface{}
	err := _Credentialregistry.contract.Call(opts, &out, "isCredentialExpired", ccid, credentialTypeId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsCredentialExpired is a free data retrieval call binding the contract method 0xa1f5de49.
//
// Solidity: function isCredentialExpired(bytes32 ccid, bytes32 credentialTypeId) view returns(bool)
func (_Credentialregistry *CredentialregistrySession) IsCredentialExpired(ccid [32]byte, credentialTypeId [32]byte) (bool, error) {
	return _Credentialregistry.Contract.IsCredentialExpired(&_Credentialregistry.CallOpts, ccid, credentialTypeId)
}

// IsCredentialExpired is a free data retrieval call binding the contract method 0xa1f5de49.
//
// Solidity: function isCredentialExpired(bytes32 ccid, bytes32 credentialTypeId) view returns(bool)
func (_Credentialregistry *CredentialregistryCallerSession) IsCredentialExpired(ccid [32]byte, credentialTypeId [32]byte) (bool, error) {
	return _Credentialregistry.Contract.IsCredentialExpired(&_Credentialregistry.CallOpts, ccid, credentialTypeId)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Credentialregistry *CredentialregistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Credentialregistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Credentialregistry *CredentialregistrySession) Owner() (common.Address, error) {
	return _Credentialregistry.Contract.Owner(&_Credentialregistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Credentialregistry *CredentialregistryCallerSession) Owner() (common.Address, error) {
	return _Credentialregistry.Contract.Owner(&_Credentialregistry.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Credentialregistry *CredentialregistryCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Credentialregistry.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Credentialregistry *CredentialregistrySession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Credentialregistry.Contract.SupportsInterface(&_Credentialregistry.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Credentialregistry *CredentialregistryCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Credentialregistry.Contract.SupportsInterface(&_Credentialregistry.CallOpts, interfaceId)
}

// Validate is a free data retrieval call binding the contract method 0x6310b8bb.
//
// Solidity: function validate(bytes32 ccid, bytes32 credentialTypeId, bytes context) view returns(bool)
func (_Credentialregistry *CredentialregistryCaller) Validate(opts *bind.CallOpts, ccid [32]byte, credentialTypeId [32]byte, context []byte) (bool, error) {
	var out []interface{}
	err := _Credentialregistry.contract.Call(opts, &out, "validate", ccid, credentialTypeId, context)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Validate is a free data retrieval call binding the contract method 0x6310b8bb.
//
// Solidity: function validate(bytes32 ccid, bytes32 credentialTypeId, bytes context) view returns(bool)
func (_Credentialregistry *CredentialregistrySession) Validate(ccid [32]byte, credentialTypeId [32]byte, context []byte) (bool, error) {
	return _Credentialregistry.Contract.Validate(&_Credentialregistry.CallOpts, ccid, credentialTypeId, context)
}

// Validate is a free data retrieval call binding the contract method 0x6310b8bb.
//
// Solidity: function validate(bytes32 ccid, bytes32 credentialTypeId, bytes context) view returns(bool)
func (_Credentialregistry *CredentialregistryCallerSession) Validate(ccid [32]byte, credentialTypeId [32]byte, context []byte) (bool, error) {
	return _Credentialregistry.Contract.Validate(&_Credentialregistry.CallOpts, ccid, credentialTypeId, context)
}

// ValidateAll is a free data retrieval call binding the contract method 0x3d52959e.
//
// Solidity: function validateAll(bytes32 ccid, bytes32[] credentialTypeIds, bytes context) view returns(bool)
func (_Credentialregistry *CredentialregistryCaller) ValidateAll(opts *bind.CallOpts, ccid [32]byte, credentialTypeIds [][32]byte, context []byte) (bool, error) {
	var out []interface{}
	err := _Credentialregistry.contract.Call(opts, &out, "validateAll", ccid, credentialTypeIds, context)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ValidateAll is a free data retrieval call binding the contract method 0x3d52959e.
//
// Solidity: function validateAll(bytes32 ccid, bytes32[] credentialTypeIds, bytes context) view returns(bool)
func (_Credentialregistry *CredentialregistrySession) ValidateAll(ccid [32]byte, credentialTypeIds [][32]byte, context []byte) (bool, error) {
	return _Credentialregistry.Contract.ValidateAll(&_Credentialregistry.CallOpts, ccid, credentialTypeIds, context)
}

// ValidateAll is a free data retrieval call binding the contract method 0x3d52959e.
//
// Solidity: function validateAll(bytes32 ccid, bytes32[] credentialTypeIds, bytes context) view returns(bool)
func (_Credentialregistry *CredentialregistryCallerSession) ValidateAll(ccid [32]byte, credentialTypeIds [][32]byte, context []byte) (bool, error) {
	return _Credentialregistry.Contract.ValidateAll(&_Credentialregistry.CallOpts, ccid, credentialTypeIds, context)
}

// AttachPolicyEngine is a paid mutator transaction binding the contract method 0xf7332e86.
//
// Solidity: function attachPolicyEngine(address policyEngine) returns()
func (_Credentialregistry *CredentialregistryTransactor) AttachPolicyEngine(opts *bind.TransactOpts, policyEngine common.Address) (*types.Transaction, error) {
	return _Credentialregistry.contract.Transact(opts, "attachPolicyEngine", policyEngine)
}

// AttachPolicyEngine is a paid mutator transaction binding the contract method 0xf7332e86.
//
// Solidity: function attachPolicyEngine(address policyEngine) returns()
func (_Credentialregistry *CredentialregistrySession) AttachPolicyEngine(policyEngine common.Address) (*types.Transaction, error) {
	return _Credentialregistry.Contract.AttachPolicyEngine(&_Credentialregistry.TransactOpts, policyEngine)
}

// AttachPolicyEngine is a paid mutator transaction binding the contract method 0xf7332e86.
//
// Solidity: function attachPolicyEngine(address policyEngine) returns()
func (_Credentialregistry *CredentialregistryTransactorSession) AttachPolicyEngine(policyEngine common.Address) (*types.Transaction, error) {
	return _Credentialregistry.Contract.AttachPolicyEngine(&_Credentialregistry.TransactOpts, policyEngine)
}

// ClearContext is a paid mutator transaction binding the contract method 0xdb99bddd.
//
// Solidity: function clearContext() returns()
func (_Credentialregistry *CredentialregistryTransactor) ClearContext(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Credentialregistry.contract.Transact(opts, "clearContext")
}

// ClearContext is a paid mutator transaction binding the contract method 0xdb99bddd.
//
// Solidity: function clearContext() returns()
func (_Credentialregistry *CredentialregistrySession) ClearContext() (*types.Transaction, error) {
	return _Credentialregistry.Contract.ClearContext(&_Credentialregistry.TransactOpts)
}

// ClearContext is a paid mutator transaction binding the contract method 0xdb99bddd.
//
// Solidity: function clearContext() returns()
func (_Credentialregistry *CredentialregistryTransactorSession) ClearContext() (*types.Transaction, error) {
	return _Credentialregistry.Contract.ClearContext(&_Credentialregistry.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address policyEngine) returns()
func (_Credentialregistry *CredentialregistryTransactor) Initialize(opts *bind.TransactOpts, policyEngine common.Address) (*types.Transaction, error) {
	return _Credentialregistry.contract.Transact(opts, "initialize", policyEngine)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address policyEngine) returns()
func (_Credentialregistry *CredentialregistrySession) Initialize(policyEngine common.Address) (*types.Transaction, error) {
	return _Credentialregistry.Contract.Initialize(&_Credentialregistry.TransactOpts, policyEngine)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address policyEngine) returns()
func (_Credentialregistry *CredentialregistryTransactorSession) Initialize(policyEngine common.Address) (*types.Transaction, error) {
	return _Credentialregistry.Contract.Initialize(&_Credentialregistry.TransactOpts, policyEngine)
}

// RegisterCredential is a paid mutator transaction binding the contract method 0x8ab73618.
//
// Solidity: function registerCredential(bytes32 ccid, bytes32 credentialTypeId, uint40 expiresAt, bytes credentialData, bytes context) returns()
func (_Credentialregistry *CredentialregistryTransactor) RegisterCredential(opts *bind.TransactOpts, ccid [32]byte, credentialTypeId [32]byte, expiresAt *big.Int, credentialData []byte, context []byte) (*types.Transaction, error) {
	return _Credentialregistry.contract.Transact(opts, "registerCredential", ccid, credentialTypeId, expiresAt, credentialData, context)
}

// RegisterCredential is a paid mutator transaction binding the contract method 0x8ab73618.
//
// Solidity: function registerCredential(bytes32 ccid, bytes32 credentialTypeId, uint40 expiresAt, bytes credentialData, bytes context) returns()
func (_Credentialregistry *CredentialregistrySession) RegisterCredential(ccid [32]byte, credentialTypeId [32]byte, expiresAt *big.Int, credentialData []byte, context []byte) (*types.Transaction, error) {
	return _Credentialregistry.Contract.RegisterCredential(&_Credentialregistry.TransactOpts, ccid, credentialTypeId, expiresAt, credentialData, context)
}

// RegisterCredential is a paid mutator transaction binding the contract method 0x8ab73618.
//
// Solidity: function registerCredential(bytes32 ccid, bytes32 credentialTypeId, uint40 expiresAt, bytes credentialData, bytes context) returns()
func (_Credentialregistry *CredentialregistryTransactorSession) RegisterCredential(ccid [32]byte, credentialTypeId [32]byte, expiresAt *big.Int, credentialData []byte, context []byte) (*types.Transaction, error) {
	return _Credentialregistry.Contract.RegisterCredential(&_Credentialregistry.TransactOpts, ccid, credentialTypeId, expiresAt, credentialData, context)
}

// RegisterCredentials is a paid mutator transaction binding the contract method 0xf6635baf.
//
// Solidity: function registerCredentials(bytes32 ccid, bytes32[] credentialTypeIds, uint40 expiresAt, bytes[] credentialDatas, bytes context) returns()
func (_Credentialregistry *CredentialregistryTransactor) RegisterCredentials(opts *bind.TransactOpts, ccid [32]byte, credentialTypeIds [][32]byte, expiresAt *big.Int, credentialDatas [][]byte, context []byte) (*types.Transaction, error) {
	return _Credentialregistry.contract.Transact(opts, "registerCredentials", ccid, credentialTypeIds, expiresAt, credentialDatas, context)
}

// RegisterCredentials is a paid mutator transaction binding the contract method 0xf6635baf.
//
// Solidity: function registerCredentials(bytes32 ccid, bytes32[] credentialTypeIds, uint40 expiresAt, bytes[] credentialDatas, bytes context) returns()
func (_Credentialregistry *CredentialregistrySession) RegisterCredentials(ccid [32]byte, credentialTypeIds [][32]byte, expiresAt *big.Int, credentialDatas [][]byte, context []byte) (*types.Transaction, error) {
	return _Credentialregistry.Contract.RegisterCredentials(&_Credentialregistry.TransactOpts, ccid, credentialTypeIds, expiresAt, credentialDatas, context)
}

// RegisterCredentials is a paid mutator transaction binding the contract method 0xf6635baf.
//
// Solidity: function registerCredentials(bytes32 ccid, bytes32[] credentialTypeIds, uint40 expiresAt, bytes[] credentialDatas, bytes context) returns()
func (_Credentialregistry *CredentialregistryTransactorSession) RegisterCredentials(ccid [32]byte, credentialTypeIds [][32]byte, expiresAt *big.Int, credentialDatas [][]byte, context []byte) (*types.Transaction, error) {
	return _Credentialregistry.Contract.RegisterCredentials(&_Credentialregistry.TransactOpts, ccid, credentialTypeIds, expiresAt, credentialDatas, context)
}

// RemoveCredential is a paid mutator transaction binding the contract method 0x93893fcb.
//
// Solidity: function removeCredential(bytes32 ccid, bytes32 credentialTypeId, bytes context) returns()
func (_Credentialregistry *CredentialregistryTransactor) RemoveCredential(opts *bind.TransactOpts, ccid [32]byte, credentialTypeId [32]byte, context []byte) (*types.Transaction, error) {
	return _Credentialregistry.contract.Transact(opts, "removeCredential", ccid, credentialTypeId, context)
}

// RemoveCredential is a paid mutator transaction binding the contract method 0x93893fcb.
//
// Solidity: function removeCredential(bytes32 ccid, bytes32 credentialTypeId, bytes context) returns()
func (_Credentialregistry *CredentialregistrySession) RemoveCredential(ccid [32]byte, credentialTypeId [32]byte, context []byte) (*types.Transaction, error) {
	return _Credentialregistry.Contract.RemoveCredential(&_Credentialregistry.TransactOpts, ccid, credentialTypeId, context)
}

// RemoveCredential is a paid mutator transaction binding the contract method 0x93893fcb.
//
// Solidity: function removeCredential(bytes32 ccid, bytes32 credentialTypeId, bytes context) returns()
func (_Credentialregistry *CredentialregistryTransactorSession) RemoveCredential(ccid [32]byte, credentialTypeId [32]byte, context []byte) (*types.Transaction, error) {
	return _Credentialregistry.Contract.RemoveCredential(&_Credentialregistry.TransactOpts, ccid, credentialTypeId, context)
}

// RenewCredential is a paid mutator transaction binding the contract method 0x22fe00bb.
//
// Solidity: function renewCredential(bytes32 ccid, bytes32 credentialTypeId, uint40 expiresAt, bytes context) returns()
func (_Credentialregistry *CredentialregistryTransactor) RenewCredential(opts *bind.TransactOpts, ccid [32]byte, credentialTypeId [32]byte, expiresAt *big.Int, context []byte) (*types.Transaction, error) {
	return _Credentialregistry.contract.Transact(opts, "renewCredential", ccid, credentialTypeId, expiresAt, context)
}

// RenewCredential is a paid mutator transaction binding the contract method 0x22fe00bb.
//
// Solidity: function renewCredential(bytes32 ccid, bytes32 credentialTypeId, uint40 expiresAt, bytes context) returns()
func (_Credentialregistry *CredentialregistrySession) RenewCredential(ccid [32]byte, credentialTypeId [32]byte, expiresAt *big.Int, context []byte) (*types.Transaction, error) {
	return _Credentialregistry.Contract.RenewCredential(&_Credentialregistry.TransactOpts, ccid, credentialTypeId, expiresAt, context)
}

// RenewCredential is a paid mutator transaction binding the contract method 0x22fe00bb.
//
// Solidity: function renewCredential(bytes32 ccid, bytes32 credentialTypeId, uint40 expiresAt, bytes context) returns()
func (_Credentialregistry *CredentialregistryTransactorSession) RenewCredential(ccid [32]byte, credentialTypeId [32]byte, expiresAt *big.Int, context []byte) (*types.Transaction, error) {
	return _Credentialregistry.Contract.RenewCredential(&_Credentialregistry.TransactOpts, ccid, credentialTypeId, expiresAt, context)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Credentialregistry *CredentialregistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Credentialregistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Credentialregistry *CredentialregistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _Credentialregistry.Contract.RenounceOwnership(&_Credentialregistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Credentialregistry *CredentialregistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Credentialregistry.Contract.RenounceOwnership(&_Credentialregistry.TransactOpts)
}

// SetContext is a paid mutator transaction binding the contract method 0x2f7482be.
//
// Solidity: function setContext(bytes context) returns()
func (_Credentialregistry *CredentialregistryTransactor) SetContext(opts *bind.TransactOpts, context []byte) (*types.Transaction, error) {
	return _Credentialregistry.contract.Transact(opts, "setContext", context)
}

// SetContext is a paid mutator transaction binding the contract method 0x2f7482be.
//
// Solidity: function setContext(bytes context) returns()
func (_Credentialregistry *CredentialregistrySession) SetContext(context []byte) (*types.Transaction, error) {
	return _Credentialregistry.Contract.SetContext(&_Credentialregistry.TransactOpts, context)
}

// SetContext is a paid mutator transaction binding the contract method 0x2f7482be.
//
// Solidity: function setContext(bytes context) returns()
func (_Credentialregistry *CredentialregistryTransactorSession) SetContext(context []byte) (*types.Transaction, error) {
	return _Credentialregistry.Contract.SetContext(&_Credentialregistry.TransactOpts, context)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Credentialregistry *CredentialregistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Credentialregistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Credentialregistry *CredentialregistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Credentialregistry.Contract.TransferOwnership(&_Credentialregistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Credentialregistry *CredentialregistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Credentialregistry.Contract.TransferOwnership(&_Credentialregistry.TransactOpts, newOwner)
}

// CredentialregistryCredentialRegisteredIterator is returned from FilterCredentialRegistered and is used to iterate over the raw logs and unpacked data for CredentialRegistered events raised by the Credentialregistry contract.
type CredentialregistryCredentialRegisteredIterator struct {
	Event *CredentialregistryCredentialRegistered // Event containing the contract specifics and raw log

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
func (it *CredentialregistryCredentialRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CredentialregistryCredentialRegistered)
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
		it.Event = new(CredentialregistryCredentialRegistered)
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
func (it *CredentialregistryCredentialRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CredentialregistryCredentialRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CredentialregistryCredentialRegistered represents a CredentialRegistered event raised by the Credentialregistry contract.
type CredentialregistryCredentialRegistered struct {
	Ccid             [32]byte
	CredentialTypeId [32]byte
	ExpiresAt        *big.Int
	CredentialData   []byte
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterCredentialRegistered is a free log retrieval operation binding the contract event 0x7e9980c30ec32e33b4cd0a29a359faf580622f314894f08391c798c8354f89c7.
//
// Solidity: event CredentialRegistered(bytes32 indexed ccid, bytes32 indexed credentialTypeId, uint40 expiresAt, bytes credentialData)
func (_Credentialregistry *CredentialregistryFilterer) FilterCredentialRegistered(opts *bind.FilterOpts, ccid [][32]byte, credentialTypeId [][32]byte) (*CredentialregistryCredentialRegisteredIterator, error) {

	var ccidRule []interface{}
	for _, ccidItem := range ccid {
		ccidRule = append(ccidRule, ccidItem)
	}
	var credentialTypeIdRule []interface{}
	for _, credentialTypeIdItem := range credentialTypeId {
		credentialTypeIdRule = append(credentialTypeIdRule, credentialTypeIdItem)
	}

	logs, sub, err := _Credentialregistry.contract.FilterLogs(opts, "CredentialRegistered", ccidRule, credentialTypeIdRule)
	if err != nil {
		return nil, err
	}
	return &CredentialregistryCredentialRegisteredIterator{contract: _Credentialregistry.contract, event: "CredentialRegistered", logs: logs, sub: sub}, nil
}

// WatchCredentialRegistered is a free log subscription operation binding the contract event 0x7e9980c30ec32e33b4cd0a29a359faf580622f314894f08391c798c8354f89c7.
//
// Solidity: event CredentialRegistered(bytes32 indexed ccid, bytes32 indexed credentialTypeId, uint40 expiresAt, bytes credentialData)
func (_Credentialregistry *CredentialregistryFilterer) WatchCredentialRegistered(opts *bind.WatchOpts, sink chan<- *CredentialregistryCredentialRegistered, ccid [][32]byte, credentialTypeId [][32]byte) (event.Subscription, error) {

	var ccidRule []interface{}
	for _, ccidItem := range ccid {
		ccidRule = append(ccidRule, ccidItem)
	}
	var credentialTypeIdRule []interface{}
	for _, credentialTypeIdItem := range credentialTypeId {
		credentialTypeIdRule = append(credentialTypeIdRule, credentialTypeIdItem)
	}

	logs, sub, err := _Credentialregistry.contract.WatchLogs(opts, "CredentialRegistered", ccidRule, credentialTypeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CredentialregistryCredentialRegistered)
				if err := _Credentialregistry.contract.UnpackLog(event, "CredentialRegistered", log); err != nil {
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

// ParseCredentialRegistered is a log parse operation binding the contract event 0x7e9980c30ec32e33b4cd0a29a359faf580622f314894f08391c798c8354f89c7.
//
// Solidity: event CredentialRegistered(bytes32 indexed ccid, bytes32 indexed credentialTypeId, uint40 expiresAt, bytes credentialData)
func (_Credentialregistry *CredentialregistryFilterer) ParseCredentialRegistered(log types.Log) (*CredentialregistryCredentialRegistered, error) {
	event := new(CredentialregistryCredentialRegistered)
	if err := _Credentialregistry.contract.UnpackLog(event, "CredentialRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CredentialregistryCredentialRemovedIterator is returned from FilterCredentialRemoved and is used to iterate over the raw logs and unpacked data for CredentialRemoved events raised by the Credentialregistry contract.
type CredentialregistryCredentialRemovedIterator struct {
	Event *CredentialregistryCredentialRemoved // Event containing the contract specifics and raw log

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
func (it *CredentialregistryCredentialRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CredentialregistryCredentialRemoved)
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
		it.Event = new(CredentialregistryCredentialRemoved)
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
func (it *CredentialregistryCredentialRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CredentialregistryCredentialRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CredentialregistryCredentialRemoved represents a CredentialRemoved event raised by the Credentialregistry contract.
type CredentialregistryCredentialRemoved struct {
	Ccid             [32]byte
	CredentialTypeId [32]byte
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterCredentialRemoved is a free log retrieval operation binding the contract event 0x9094e025ca0806e26dfbcc0f0784f043a038cfaeafcce2e05cbcfc6bb206835e.
//
// Solidity: event CredentialRemoved(bytes32 indexed ccid, bytes32 indexed credentialTypeId)
func (_Credentialregistry *CredentialregistryFilterer) FilterCredentialRemoved(opts *bind.FilterOpts, ccid [][32]byte, credentialTypeId [][32]byte) (*CredentialregistryCredentialRemovedIterator, error) {

	var ccidRule []interface{}
	for _, ccidItem := range ccid {
		ccidRule = append(ccidRule, ccidItem)
	}
	var credentialTypeIdRule []interface{}
	for _, credentialTypeIdItem := range credentialTypeId {
		credentialTypeIdRule = append(credentialTypeIdRule, credentialTypeIdItem)
	}

	logs, sub, err := _Credentialregistry.contract.FilterLogs(opts, "CredentialRemoved", ccidRule, credentialTypeIdRule)
	if err != nil {
		return nil, err
	}
	return &CredentialregistryCredentialRemovedIterator{contract: _Credentialregistry.contract, event: "CredentialRemoved", logs: logs, sub: sub}, nil
}

// WatchCredentialRemoved is a free log subscription operation binding the contract event 0x9094e025ca0806e26dfbcc0f0784f043a038cfaeafcce2e05cbcfc6bb206835e.
//
// Solidity: event CredentialRemoved(bytes32 indexed ccid, bytes32 indexed credentialTypeId)
func (_Credentialregistry *CredentialregistryFilterer) WatchCredentialRemoved(opts *bind.WatchOpts, sink chan<- *CredentialregistryCredentialRemoved, ccid [][32]byte, credentialTypeId [][32]byte) (event.Subscription, error) {

	var ccidRule []interface{}
	for _, ccidItem := range ccid {
		ccidRule = append(ccidRule, ccidItem)
	}
	var credentialTypeIdRule []interface{}
	for _, credentialTypeIdItem := range credentialTypeId {
		credentialTypeIdRule = append(credentialTypeIdRule, credentialTypeIdItem)
	}

	logs, sub, err := _Credentialregistry.contract.WatchLogs(opts, "CredentialRemoved", ccidRule, credentialTypeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CredentialregistryCredentialRemoved)
				if err := _Credentialregistry.contract.UnpackLog(event, "CredentialRemoved", log); err != nil {
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

// ParseCredentialRemoved is a log parse operation binding the contract event 0x9094e025ca0806e26dfbcc0f0784f043a038cfaeafcce2e05cbcfc6bb206835e.
//
// Solidity: event CredentialRemoved(bytes32 indexed ccid, bytes32 indexed credentialTypeId)
func (_Credentialregistry *CredentialregistryFilterer) ParseCredentialRemoved(log types.Log) (*CredentialregistryCredentialRemoved, error) {
	event := new(CredentialregistryCredentialRemoved)
	if err := _Credentialregistry.contract.UnpackLog(event, "CredentialRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CredentialregistryCredentialRenewedIterator is returned from FilterCredentialRenewed and is used to iterate over the raw logs and unpacked data for CredentialRenewed events raised by the Credentialregistry contract.
type CredentialregistryCredentialRenewedIterator struct {
	Event *CredentialregistryCredentialRenewed // Event containing the contract specifics and raw log

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
func (it *CredentialregistryCredentialRenewedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CredentialregistryCredentialRenewed)
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
		it.Event = new(CredentialregistryCredentialRenewed)
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
func (it *CredentialregistryCredentialRenewedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CredentialregistryCredentialRenewedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CredentialregistryCredentialRenewed represents a CredentialRenewed event raised by the Credentialregistry contract.
type CredentialregistryCredentialRenewed struct {
	Ccid             [32]byte
	CredentialTypeId [32]byte
	ExpiresAt        *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterCredentialRenewed is a free log retrieval operation binding the contract event 0xd47167880b9b6e85302ff993a0108cc563be2bc38369ca1bb83b9e685c29db59.
//
// Solidity: event CredentialRenewed(bytes32 indexed ccid, bytes32 indexed credentialTypeId, uint40 expiresAt)
func (_Credentialregistry *CredentialregistryFilterer) FilterCredentialRenewed(opts *bind.FilterOpts, ccid [][32]byte, credentialTypeId [][32]byte) (*CredentialregistryCredentialRenewedIterator, error) {

	var ccidRule []interface{}
	for _, ccidItem := range ccid {
		ccidRule = append(ccidRule, ccidItem)
	}
	var credentialTypeIdRule []interface{}
	for _, credentialTypeIdItem := range credentialTypeId {
		credentialTypeIdRule = append(credentialTypeIdRule, credentialTypeIdItem)
	}

	logs, sub, err := _Credentialregistry.contract.FilterLogs(opts, "CredentialRenewed", ccidRule, credentialTypeIdRule)
	if err != nil {
		return nil, err
	}
	return &CredentialregistryCredentialRenewedIterator{contract: _Credentialregistry.contract, event: "CredentialRenewed", logs: logs, sub: sub}, nil
}

// WatchCredentialRenewed is a free log subscription operation binding the contract event 0xd47167880b9b6e85302ff993a0108cc563be2bc38369ca1bb83b9e685c29db59.
//
// Solidity: event CredentialRenewed(bytes32 indexed ccid, bytes32 indexed credentialTypeId, uint40 expiresAt)
func (_Credentialregistry *CredentialregistryFilterer) WatchCredentialRenewed(opts *bind.WatchOpts, sink chan<- *CredentialregistryCredentialRenewed, ccid [][32]byte, credentialTypeId [][32]byte) (event.Subscription, error) {

	var ccidRule []interface{}
	for _, ccidItem := range ccid {
		ccidRule = append(ccidRule, ccidItem)
	}
	var credentialTypeIdRule []interface{}
	for _, credentialTypeIdItem := range credentialTypeId {
		credentialTypeIdRule = append(credentialTypeIdRule, credentialTypeIdItem)
	}

	logs, sub, err := _Credentialregistry.contract.WatchLogs(opts, "CredentialRenewed", ccidRule, credentialTypeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CredentialregistryCredentialRenewed)
				if err := _Credentialregistry.contract.UnpackLog(event, "CredentialRenewed", log); err != nil {
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

// ParseCredentialRenewed is a log parse operation binding the contract event 0xd47167880b9b6e85302ff993a0108cc563be2bc38369ca1bb83b9e685c29db59.
//
// Solidity: event CredentialRenewed(bytes32 indexed ccid, bytes32 indexed credentialTypeId, uint40 expiresAt)
func (_Credentialregistry *CredentialregistryFilterer) ParseCredentialRenewed(log types.Log) (*CredentialregistryCredentialRenewed, error) {
	event := new(CredentialregistryCredentialRenewed)
	if err := _Credentialregistry.contract.UnpackLog(event, "CredentialRenewed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CredentialregistryInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Credentialregistry contract.
type CredentialregistryInitializedIterator struct {
	Event *CredentialregistryInitialized // Event containing the contract specifics and raw log

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
func (it *CredentialregistryInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CredentialregistryInitialized)
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
		it.Event = new(CredentialregistryInitialized)
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
func (it *CredentialregistryInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CredentialregistryInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CredentialregistryInitialized represents a Initialized event raised by the Credentialregistry contract.
type CredentialregistryInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Credentialregistry *CredentialregistryFilterer) FilterInitialized(opts *bind.FilterOpts) (*CredentialregistryInitializedIterator, error) {

	logs, sub, err := _Credentialregistry.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &CredentialregistryInitializedIterator{contract: _Credentialregistry.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Credentialregistry *CredentialregistryFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *CredentialregistryInitialized) (event.Subscription, error) {

	logs, sub, err := _Credentialregistry.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CredentialregistryInitialized)
				if err := _Credentialregistry.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_Credentialregistry *CredentialregistryFilterer) ParseInitialized(log types.Log) (*CredentialregistryInitialized, error) {
	event := new(CredentialregistryInitialized)
	if err := _Credentialregistry.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CredentialregistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Credentialregistry contract.
type CredentialregistryOwnershipTransferredIterator struct {
	Event *CredentialregistryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *CredentialregistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CredentialregistryOwnershipTransferred)
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
		it.Event = new(CredentialregistryOwnershipTransferred)
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
func (it *CredentialregistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CredentialregistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CredentialregistryOwnershipTransferred represents a OwnershipTransferred event raised by the Credentialregistry contract.
type CredentialregistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Credentialregistry *CredentialregistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*CredentialregistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Credentialregistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &CredentialregistryOwnershipTransferredIterator{contract: _Credentialregistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Credentialregistry *CredentialregistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *CredentialregistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Credentialregistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CredentialregistryOwnershipTransferred)
				if err := _Credentialregistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Credentialregistry *CredentialregistryFilterer) ParseOwnershipTransferred(log types.Log) (*CredentialregistryOwnershipTransferred, error) {
	event := new(CredentialregistryOwnershipTransferred)
	if err := _Credentialregistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CredentialregistryPolicyEngineAttachedIterator is returned from FilterPolicyEngineAttached and is used to iterate over the raw logs and unpacked data for PolicyEngineAttached events raised by the Credentialregistry contract.
type CredentialregistryPolicyEngineAttachedIterator struct {
	Event *CredentialregistryPolicyEngineAttached // Event containing the contract specifics and raw log

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
func (it *CredentialregistryPolicyEngineAttachedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CredentialregistryPolicyEngineAttached)
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
		it.Event = new(CredentialregistryPolicyEngineAttached)
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
func (it *CredentialregistryPolicyEngineAttachedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CredentialregistryPolicyEngineAttachedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CredentialregistryPolicyEngineAttached represents a PolicyEngineAttached event raised by the Credentialregistry contract.
type CredentialregistryPolicyEngineAttached struct {
	PolicyEngine common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterPolicyEngineAttached is a free log retrieval operation binding the contract event 0x57d241970863a27bedbf58b705b45a0b267f76f9a3a7fd432e217a37e4173fac.
//
// Solidity: event PolicyEngineAttached(address indexed policyEngine)
func (_Credentialregistry *CredentialregistryFilterer) FilterPolicyEngineAttached(opts *bind.FilterOpts, policyEngine []common.Address) (*CredentialregistryPolicyEngineAttachedIterator, error) {

	var policyEngineRule []interface{}
	for _, policyEngineItem := range policyEngine {
		policyEngineRule = append(policyEngineRule, policyEngineItem)
	}

	logs, sub, err := _Credentialregistry.contract.FilterLogs(opts, "PolicyEngineAttached", policyEngineRule)
	if err != nil {
		return nil, err
	}
	return &CredentialregistryPolicyEngineAttachedIterator{contract: _Credentialregistry.contract, event: "PolicyEngineAttached", logs: logs, sub: sub}, nil
}

// WatchPolicyEngineAttached is a free log subscription operation binding the contract event 0x57d241970863a27bedbf58b705b45a0b267f76f9a3a7fd432e217a37e4173fac.
//
// Solidity: event PolicyEngineAttached(address indexed policyEngine)
func (_Credentialregistry *CredentialregistryFilterer) WatchPolicyEngineAttached(opts *bind.WatchOpts, sink chan<- *CredentialregistryPolicyEngineAttached, policyEngine []common.Address) (event.Subscription, error) {

	var policyEngineRule []interface{}
	for _, policyEngineItem := range policyEngine {
		policyEngineRule = append(policyEngineRule, policyEngineItem)
	}

	logs, sub, err := _Credentialregistry.contract.WatchLogs(opts, "PolicyEngineAttached", policyEngineRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CredentialregistryPolicyEngineAttached)
				if err := _Credentialregistry.contract.UnpackLog(event, "PolicyEngineAttached", log); err != nil {
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
func (_Credentialregistry *CredentialregistryFilterer) ParsePolicyEngineAttached(log types.Log) (*CredentialregistryPolicyEngineAttached, error) {
	event := new(CredentialregistryPolicyEngineAttached)
	if err := _Credentialregistry.contract.UnpackLog(event, "PolicyEngineAttached", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
