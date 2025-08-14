// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package identityvalidator

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

// ICredentialRequirementsCredentialRequirement is an auto generated low-level Go binding around an user-defined struct.
type ICredentialRequirementsCredentialRequirement struct {
	CredentialTypeIds [][32]byte
	MinValidations    *big.Int
	Invert            bool
}

// ICredentialRequirementsCredentialRequirementInput is an auto generated low-level Go binding around an user-defined struct.
type ICredentialRequirementsCredentialRequirementInput struct {
	RequirementId     [32]byte
	CredentialTypeIds [][32]byte
	MinValidations    *big.Int
	Invert            bool
}

// ICredentialRequirementsCredentialSource is an auto generated low-level Go binding around an user-defined struct.
type ICredentialRequirementsCredentialSource struct {
	IdentityRegistry   common.Address
	CredentialRegistry common.Address
	DataValidator      common.Address
}

// ICredentialRequirementsCredentialSourceInput is an auto generated low-level Go binding around an user-defined struct.
type ICredentialRequirementsCredentialSourceInput struct {
	CredentialTypeId   [32]byte
	IdentityRegistry   common.Address
	CredentialRegistry common.Address
	DataValidator      common.Address
}

// IdentityvalidatorMetaData contains all meta data concerning the Identityvalidator contract.
var IdentityvalidatorMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"addCredentialRequirement\",\"inputs\":[{\"name\":\"input\",\"type\":\"tuple\",\"internalType\":\"structICredentialRequirements.CredentialRequirementInput\",\"components\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeIds\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"minValidations\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"invert\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"addCredentialSource\",\"inputs\":[{\"name\":\"input\",\"type\":\"tuple\",\"internalType\":\"structICredentialRequirements.CredentialSourceInput\",\"components\":[{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"identityRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"dataValidator\",\"type\":\"address\",\"internalType\":\"address\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getCredentialRequirement\",\"inputs\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structICredentialRequirements.CredentialRequirement\",\"components\":[{\"name\":\"credentialTypeIds\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"minValidations\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"invert\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCredentialRequirementIds\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCredentialSources\",\"inputs\":[{\"name\":\"credential\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structICredentialRequirements.CredentialSource[]\",\"components\":[{\"name\":\"identityRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"dataValidator\",\"type\":\"address\",\"internalType\":\"address\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"credentialSourceInputs\",\"type\":\"tuple[]\",\"internalType\":\"structICredentialRequirements.CredentialSourceInput[]\",\"components\":[{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"identityRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"dataValidator\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"name\":\"credentialRequirementInputs\",\"type\":\"tuple[]\",\"internalType\":\"structICredentialRequirements.CredentialRequirementInput[]\",\"components\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeIds\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"minValidations\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"invert\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"removeCredentialRequirement\",\"inputs\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"removeCredentialSource\",\"inputs\":[{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"identityRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"validate\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"context\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"CredentialRequirementAdded\",\"inputs\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeIds\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"minValidations\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CredentialRequirementRemoved\",\"inputs\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeIds\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"minValidations\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CredentialSourceAdded\",\"inputs\":[{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"identityRegistry\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"dataValidator\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CredentialSourceRemoved\",\"inputs\":[{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"identityRegistry\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"dataValidator\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"InvalidConfiguration\",\"inputs\":[{\"name\":\"errorReason\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"RequirementExists\",\"inputs\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"RequirementNotFound\",\"inputs\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"SourceExists\",\"inputs\":[{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"identityRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"SourceNotFound\",\"inputs\":[{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"identityRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"internalType\":\"address\"}]}]",
}

// IdentityvalidatorABI is the input ABI used to generate the binding from.
// Deprecated: Use IdentityvalidatorMetaData.ABI instead.
var IdentityvalidatorABI = IdentityvalidatorMetaData.ABI

// Identityvalidator is an auto generated Go binding around an Ethereum contract.
type Identityvalidator struct {
	IdentityvalidatorCaller     // Read-only binding to the contract
	IdentityvalidatorTransactor // Write-only binding to the contract
	IdentityvalidatorFilterer   // Log filterer for contract events
}

// IdentityvalidatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type IdentityvalidatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IdentityvalidatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IdentityvalidatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IdentityvalidatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IdentityvalidatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IdentityvalidatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IdentityvalidatorSession struct {
	Contract     *Identityvalidator // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// IdentityvalidatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IdentityvalidatorCallerSession struct {
	Contract *IdentityvalidatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// IdentityvalidatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IdentityvalidatorTransactorSession struct {
	Contract     *IdentityvalidatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// IdentityvalidatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type IdentityvalidatorRaw struct {
	Contract *Identityvalidator // Generic contract binding to access the raw methods on
}

// IdentityvalidatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IdentityvalidatorCallerRaw struct {
	Contract *IdentityvalidatorCaller // Generic read-only contract binding to access the raw methods on
}

// IdentityvalidatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IdentityvalidatorTransactorRaw struct {
	Contract *IdentityvalidatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIdentityvalidator creates a new instance of Identityvalidator, bound to a specific deployed contract.
func NewIdentityvalidator(address common.Address, backend bind.ContractBackend) (*Identityvalidator, error) {
	contract, err := bindIdentityvalidator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Identityvalidator{IdentityvalidatorCaller: IdentityvalidatorCaller{contract: contract}, IdentityvalidatorTransactor: IdentityvalidatorTransactor{contract: contract}, IdentityvalidatorFilterer: IdentityvalidatorFilterer{contract: contract}}, nil
}

// NewIdentityvalidatorCaller creates a new read-only instance of Identityvalidator, bound to a specific deployed contract.
func NewIdentityvalidatorCaller(address common.Address, caller bind.ContractCaller) (*IdentityvalidatorCaller, error) {
	contract, err := bindIdentityvalidator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorCaller{contract: contract}, nil
}

// NewIdentityvalidatorTransactor creates a new write-only instance of Identityvalidator, bound to a specific deployed contract.
func NewIdentityvalidatorTransactor(address common.Address, transactor bind.ContractTransactor) (*IdentityvalidatorTransactor, error) {
	contract, err := bindIdentityvalidator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorTransactor{contract: contract}, nil
}

// NewIdentityvalidatorFilterer creates a new log filterer instance of Identityvalidator, bound to a specific deployed contract.
func NewIdentityvalidatorFilterer(address common.Address, filterer bind.ContractFilterer) (*IdentityvalidatorFilterer, error) {
	contract, err := bindIdentityvalidator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorFilterer{contract: contract}, nil
}

// bindIdentityvalidator binds a generic wrapper to an already deployed contract.
func bindIdentityvalidator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IdentityvalidatorMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Identityvalidator *IdentityvalidatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Identityvalidator.Contract.IdentityvalidatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Identityvalidator *IdentityvalidatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Identityvalidator.Contract.IdentityvalidatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Identityvalidator *IdentityvalidatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Identityvalidator.Contract.IdentityvalidatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Identityvalidator *IdentityvalidatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Identityvalidator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Identityvalidator *IdentityvalidatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Identityvalidator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Identityvalidator *IdentityvalidatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Identityvalidator.Contract.contract.Transact(opts, method, params...)
}

// GetCredentialRequirement is a free data retrieval call binding the contract method 0x27fbdc39.
//
// Solidity: function getCredentialRequirement(bytes32 requirementId) view returns((bytes32[],uint256,bool))
func (_Identityvalidator *IdentityvalidatorCaller) GetCredentialRequirement(opts *bind.CallOpts, requirementId [32]byte) (ICredentialRequirementsCredentialRequirement, error) {
	var out []interface{}
	err := _Identityvalidator.contract.Call(opts, &out, "getCredentialRequirement", requirementId)

	if err != nil {
		return *new(ICredentialRequirementsCredentialRequirement), err
	}

	out0 := *abi.ConvertType(out[0], new(ICredentialRequirementsCredentialRequirement)).(*ICredentialRequirementsCredentialRequirement)

	return out0, err

}

// GetCredentialRequirement is a free data retrieval call binding the contract method 0x27fbdc39.
//
// Solidity: function getCredentialRequirement(bytes32 requirementId) view returns((bytes32[],uint256,bool))
func (_Identityvalidator *IdentityvalidatorSession) GetCredentialRequirement(requirementId [32]byte) (ICredentialRequirementsCredentialRequirement, error) {
	return _Identityvalidator.Contract.GetCredentialRequirement(&_Identityvalidator.CallOpts, requirementId)
}

// GetCredentialRequirement is a free data retrieval call binding the contract method 0x27fbdc39.
//
// Solidity: function getCredentialRequirement(bytes32 requirementId) view returns((bytes32[],uint256,bool))
func (_Identityvalidator *IdentityvalidatorCallerSession) GetCredentialRequirement(requirementId [32]byte) (ICredentialRequirementsCredentialRequirement, error) {
	return _Identityvalidator.Contract.GetCredentialRequirement(&_Identityvalidator.CallOpts, requirementId)
}

// GetCredentialRequirementIds is a free data retrieval call binding the contract method 0x7fab8711.
//
// Solidity: function getCredentialRequirementIds() view returns(bytes32[])
func (_Identityvalidator *IdentityvalidatorCaller) GetCredentialRequirementIds(opts *bind.CallOpts) ([][32]byte, error) {
	var out []interface{}
	err := _Identityvalidator.contract.Call(opts, &out, "getCredentialRequirementIds")

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetCredentialRequirementIds is a free data retrieval call binding the contract method 0x7fab8711.
//
// Solidity: function getCredentialRequirementIds() view returns(bytes32[])
func (_Identityvalidator *IdentityvalidatorSession) GetCredentialRequirementIds() ([][32]byte, error) {
	return _Identityvalidator.Contract.GetCredentialRequirementIds(&_Identityvalidator.CallOpts)
}

// GetCredentialRequirementIds is a free data retrieval call binding the contract method 0x7fab8711.
//
// Solidity: function getCredentialRequirementIds() view returns(bytes32[])
func (_Identityvalidator *IdentityvalidatorCallerSession) GetCredentialRequirementIds() ([][32]byte, error) {
	return _Identityvalidator.Contract.GetCredentialRequirementIds(&_Identityvalidator.CallOpts)
}

// GetCredentialSources is a free data retrieval call binding the contract method 0x04628d5d.
//
// Solidity: function getCredentialSources(bytes32 credential) view returns((address,address,address)[])
func (_Identityvalidator *IdentityvalidatorCaller) GetCredentialSources(opts *bind.CallOpts, credential [32]byte) ([]ICredentialRequirementsCredentialSource, error) {
	var out []interface{}
	err := _Identityvalidator.contract.Call(opts, &out, "getCredentialSources", credential)

	if err != nil {
		return *new([]ICredentialRequirementsCredentialSource), err
	}

	out0 := *abi.ConvertType(out[0], new([]ICredentialRequirementsCredentialSource)).(*[]ICredentialRequirementsCredentialSource)

	return out0, err

}

// GetCredentialSources is a free data retrieval call binding the contract method 0x04628d5d.
//
// Solidity: function getCredentialSources(bytes32 credential) view returns((address,address,address)[])
func (_Identityvalidator *IdentityvalidatorSession) GetCredentialSources(credential [32]byte) ([]ICredentialRequirementsCredentialSource, error) {
	return _Identityvalidator.Contract.GetCredentialSources(&_Identityvalidator.CallOpts, credential)
}

// GetCredentialSources is a free data retrieval call binding the contract method 0x04628d5d.
//
// Solidity: function getCredentialSources(bytes32 credential) view returns((address,address,address)[])
func (_Identityvalidator *IdentityvalidatorCallerSession) GetCredentialSources(credential [32]byte) ([]ICredentialRequirementsCredentialSource, error) {
	return _Identityvalidator.Contract.GetCredentialSources(&_Identityvalidator.CallOpts, credential)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Identityvalidator *IdentityvalidatorCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Identityvalidator.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Identityvalidator *IdentityvalidatorSession) Owner() (common.Address, error) {
	return _Identityvalidator.Contract.Owner(&_Identityvalidator.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Identityvalidator *IdentityvalidatorCallerSession) Owner() (common.Address, error) {
	return _Identityvalidator.Contract.Owner(&_Identityvalidator.CallOpts)
}

// Validate is a free data retrieval call binding the contract method 0xcaf92785.
//
// Solidity: function validate(address account, bytes context) view returns(bool)
func (_Identityvalidator *IdentityvalidatorCaller) Validate(opts *bind.CallOpts, account common.Address, context []byte) (bool, error) {
	var out []interface{}
	err := _Identityvalidator.contract.Call(opts, &out, "validate", account, context)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Validate is a free data retrieval call binding the contract method 0xcaf92785.
//
// Solidity: function validate(address account, bytes context) view returns(bool)
func (_Identityvalidator *IdentityvalidatorSession) Validate(account common.Address, context []byte) (bool, error) {
	return _Identityvalidator.Contract.Validate(&_Identityvalidator.CallOpts, account, context)
}

// Validate is a free data retrieval call binding the contract method 0xcaf92785.
//
// Solidity: function validate(address account, bytes context) view returns(bool)
func (_Identityvalidator *IdentityvalidatorCallerSession) Validate(account common.Address, context []byte) (bool, error) {
	return _Identityvalidator.Contract.Validate(&_Identityvalidator.CallOpts, account, context)
}

// AddCredentialRequirement is a paid mutator transaction binding the contract method 0x4afbe01e.
//
// Solidity: function addCredentialRequirement((bytes32,bytes32[],uint256,bool) input) returns()
func (_Identityvalidator *IdentityvalidatorTransactor) AddCredentialRequirement(opts *bind.TransactOpts, input ICredentialRequirementsCredentialRequirementInput) (*types.Transaction, error) {
	return _Identityvalidator.contract.Transact(opts, "addCredentialRequirement", input)
}

// AddCredentialRequirement is a paid mutator transaction binding the contract method 0x4afbe01e.
//
// Solidity: function addCredentialRequirement((bytes32,bytes32[],uint256,bool) input) returns()
func (_Identityvalidator *IdentityvalidatorSession) AddCredentialRequirement(input ICredentialRequirementsCredentialRequirementInput) (*types.Transaction, error) {
	return _Identityvalidator.Contract.AddCredentialRequirement(&_Identityvalidator.TransactOpts, input)
}

// AddCredentialRequirement is a paid mutator transaction binding the contract method 0x4afbe01e.
//
// Solidity: function addCredentialRequirement((bytes32,bytes32[],uint256,bool) input) returns()
func (_Identityvalidator *IdentityvalidatorTransactorSession) AddCredentialRequirement(input ICredentialRequirementsCredentialRequirementInput) (*types.Transaction, error) {
	return _Identityvalidator.Contract.AddCredentialRequirement(&_Identityvalidator.TransactOpts, input)
}

// AddCredentialSource is a paid mutator transaction binding the contract method 0xe09a5423.
//
// Solidity: function addCredentialSource((bytes32,address,address,address) input) returns()
func (_Identityvalidator *IdentityvalidatorTransactor) AddCredentialSource(opts *bind.TransactOpts, input ICredentialRequirementsCredentialSourceInput) (*types.Transaction, error) {
	return _Identityvalidator.contract.Transact(opts, "addCredentialSource", input)
}

// AddCredentialSource is a paid mutator transaction binding the contract method 0xe09a5423.
//
// Solidity: function addCredentialSource((bytes32,address,address,address) input) returns()
func (_Identityvalidator *IdentityvalidatorSession) AddCredentialSource(input ICredentialRequirementsCredentialSourceInput) (*types.Transaction, error) {
	return _Identityvalidator.Contract.AddCredentialSource(&_Identityvalidator.TransactOpts, input)
}

// AddCredentialSource is a paid mutator transaction binding the contract method 0xe09a5423.
//
// Solidity: function addCredentialSource((bytes32,address,address,address) input) returns()
func (_Identityvalidator *IdentityvalidatorTransactorSession) AddCredentialSource(input ICredentialRequirementsCredentialSourceInput) (*types.Transaction, error) {
	return _Identityvalidator.Contract.AddCredentialSource(&_Identityvalidator.TransactOpts, input)
}

// Initialize is a paid mutator transaction binding the contract method 0x0fe13086.
//
// Solidity: function initialize((bytes32,address,address,address)[] credentialSourceInputs, (bytes32,bytes32[],uint256,bool)[] credentialRequirementInputs) returns()
func (_Identityvalidator *IdentityvalidatorTransactor) Initialize(opts *bind.TransactOpts, credentialSourceInputs []ICredentialRequirementsCredentialSourceInput, credentialRequirementInputs []ICredentialRequirementsCredentialRequirementInput) (*types.Transaction, error) {
	return _Identityvalidator.contract.Transact(opts, "initialize", credentialSourceInputs, credentialRequirementInputs)
}

// Initialize is a paid mutator transaction binding the contract method 0x0fe13086.
//
// Solidity: function initialize((bytes32,address,address,address)[] credentialSourceInputs, (bytes32,bytes32[],uint256,bool)[] credentialRequirementInputs) returns()
func (_Identityvalidator *IdentityvalidatorSession) Initialize(credentialSourceInputs []ICredentialRequirementsCredentialSourceInput, credentialRequirementInputs []ICredentialRequirementsCredentialRequirementInput) (*types.Transaction, error) {
	return _Identityvalidator.Contract.Initialize(&_Identityvalidator.TransactOpts, credentialSourceInputs, credentialRequirementInputs)
}

// Initialize is a paid mutator transaction binding the contract method 0x0fe13086.
//
// Solidity: function initialize((bytes32,address,address,address)[] credentialSourceInputs, (bytes32,bytes32[],uint256,bool)[] credentialRequirementInputs) returns()
func (_Identityvalidator *IdentityvalidatorTransactorSession) Initialize(credentialSourceInputs []ICredentialRequirementsCredentialSourceInput, credentialRequirementInputs []ICredentialRequirementsCredentialRequirementInput) (*types.Transaction, error) {
	return _Identityvalidator.Contract.Initialize(&_Identityvalidator.TransactOpts, credentialSourceInputs, credentialRequirementInputs)
}

// RemoveCredentialRequirement is a paid mutator transaction binding the contract method 0x6c904c42.
//
// Solidity: function removeCredentialRequirement(bytes32 requirementId) returns()
func (_Identityvalidator *IdentityvalidatorTransactor) RemoveCredentialRequirement(opts *bind.TransactOpts, requirementId [32]byte) (*types.Transaction, error) {
	return _Identityvalidator.contract.Transact(opts, "removeCredentialRequirement", requirementId)
}

// RemoveCredentialRequirement is a paid mutator transaction binding the contract method 0x6c904c42.
//
// Solidity: function removeCredentialRequirement(bytes32 requirementId) returns()
func (_Identityvalidator *IdentityvalidatorSession) RemoveCredentialRequirement(requirementId [32]byte) (*types.Transaction, error) {
	return _Identityvalidator.Contract.RemoveCredentialRequirement(&_Identityvalidator.TransactOpts, requirementId)
}

// RemoveCredentialRequirement is a paid mutator transaction binding the contract method 0x6c904c42.
//
// Solidity: function removeCredentialRequirement(bytes32 requirementId) returns()
func (_Identityvalidator *IdentityvalidatorTransactorSession) RemoveCredentialRequirement(requirementId [32]byte) (*types.Transaction, error) {
	return _Identityvalidator.Contract.RemoveCredentialRequirement(&_Identityvalidator.TransactOpts, requirementId)
}

// RemoveCredentialSource is a paid mutator transaction binding the contract method 0x3578b00d.
//
// Solidity: function removeCredentialSource(bytes32 credentialTypeId, address identityRegistry, address credentialRegistry) returns()
func (_Identityvalidator *IdentityvalidatorTransactor) RemoveCredentialSource(opts *bind.TransactOpts, credentialTypeId [32]byte, identityRegistry common.Address, credentialRegistry common.Address) (*types.Transaction, error) {
	return _Identityvalidator.contract.Transact(opts, "removeCredentialSource", credentialTypeId, identityRegistry, credentialRegistry)
}

// RemoveCredentialSource is a paid mutator transaction binding the contract method 0x3578b00d.
//
// Solidity: function removeCredentialSource(bytes32 credentialTypeId, address identityRegistry, address credentialRegistry) returns()
func (_Identityvalidator *IdentityvalidatorSession) RemoveCredentialSource(credentialTypeId [32]byte, identityRegistry common.Address, credentialRegistry common.Address) (*types.Transaction, error) {
	return _Identityvalidator.Contract.RemoveCredentialSource(&_Identityvalidator.TransactOpts, credentialTypeId, identityRegistry, credentialRegistry)
}

// RemoveCredentialSource is a paid mutator transaction binding the contract method 0x3578b00d.
//
// Solidity: function removeCredentialSource(bytes32 credentialTypeId, address identityRegistry, address credentialRegistry) returns()
func (_Identityvalidator *IdentityvalidatorTransactorSession) RemoveCredentialSource(credentialTypeId [32]byte, identityRegistry common.Address, credentialRegistry common.Address) (*types.Transaction, error) {
	return _Identityvalidator.Contract.RemoveCredentialSource(&_Identityvalidator.TransactOpts, credentialTypeId, identityRegistry, credentialRegistry)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Identityvalidator *IdentityvalidatorTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Identityvalidator.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Identityvalidator *IdentityvalidatorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Identityvalidator.Contract.RenounceOwnership(&_Identityvalidator.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Identityvalidator *IdentityvalidatorTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Identityvalidator.Contract.RenounceOwnership(&_Identityvalidator.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Identityvalidator *IdentityvalidatorTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Identityvalidator.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Identityvalidator *IdentityvalidatorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Identityvalidator.Contract.TransferOwnership(&_Identityvalidator.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Identityvalidator *IdentityvalidatorTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Identityvalidator.Contract.TransferOwnership(&_Identityvalidator.TransactOpts, newOwner)
}

// IdentityvalidatorCredentialRequirementAddedIterator is returned from FilterCredentialRequirementAdded and is used to iterate over the raw logs and unpacked data for CredentialRequirementAdded events raised by the Identityvalidator contract.
type IdentityvalidatorCredentialRequirementAddedIterator struct {
	Event *IdentityvalidatorCredentialRequirementAdded // Event containing the contract specifics and raw log

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
func (it *IdentityvalidatorCredentialRequirementAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityvalidatorCredentialRequirementAdded)
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
		it.Event = new(IdentityvalidatorCredentialRequirementAdded)
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
func (it *IdentityvalidatorCredentialRequirementAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityvalidatorCredentialRequirementAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityvalidatorCredentialRequirementAdded represents a CredentialRequirementAdded event raised by the Identityvalidator contract.
type IdentityvalidatorCredentialRequirementAdded struct {
	RequirementId     [32]byte
	CredentialTypeIds [][32]byte
	MinValidations    *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterCredentialRequirementAdded is a free log retrieval operation binding the contract event 0x5e62bddfd91274dfd4cdbfe7ca9f7dbd770b0eb6c57e2b39b5c412a6b44fb010.
//
// Solidity: event CredentialRequirementAdded(bytes32 indexed requirementId, bytes32[] credentialTypeIds, uint256 minValidations)
func (_Identityvalidator *IdentityvalidatorFilterer) FilterCredentialRequirementAdded(opts *bind.FilterOpts, requirementId [][32]byte) (*IdentityvalidatorCredentialRequirementAddedIterator, error) {

	var requirementIdRule []interface{}
	for _, requirementIdItem := range requirementId {
		requirementIdRule = append(requirementIdRule, requirementIdItem)
	}

	logs, sub, err := _Identityvalidator.contract.FilterLogs(opts, "CredentialRequirementAdded", requirementIdRule)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorCredentialRequirementAddedIterator{contract: _Identityvalidator.contract, event: "CredentialRequirementAdded", logs: logs, sub: sub}, nil
}

// WatchCredentialRequirementAdded is a free log subscription operation binding the contract event 0x5e62bddfd91274dfd4cdbfe7ca9f7dbd770b0eb6c57e2b39b5c412a6b44fb010.
//
// Solidity: event CredentialRequirementAdded(bytes32 indexed requirementId, bytes32[] credentialTypeIds, uint256 minValidations)
func (_Identityvalidator *IdentityvalidatorFilterer) WatchCredentialRequirementAdded(opts *bind.WatchOpts, sink chan<- *IdentityvalidatorCredentialRequirementAdded, requirementId [][32]byte) (event.Subscription, error) {

	var requirementIdRule []interface{}
	for _, requirementIdItem := range requirementId {
		requirementIdRule = append(requirementIdRule, requirementIdItem)
	}

	logs, sub, err := _Identityvalidator.contract.WatchLogs(opts, "CredentialRequirementAdded", requirementIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityvalidatorCredentialRequirementAdded)
				if err := _Identityvalidator.contract.UnpackLog(event, "CredentialRequirementAdded", log); err != nil {
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

// ParseCredentialRequirementAdded is a log parse operation binding the contract event 0x5e62bddfd91274dfd4cdbfe7ca9f7dbd770b0eb6c57e2b39b5c412a6b44fb010.
//
// Solidity: event CredentialRequirementAdded(bytes32 indexed requirementId, bytes32[] credentialTypeIds, uint256 minValidations)
func (_Identityvalidator *IdentityvalidatorFilterer) ParseCredentialRequirementAdded(log types.Log) (*IdentityvalidatorCredentialRequirementAdded, error) {
	event := new(IdentityvalidatorCredentialRequirementAdded)
	if err := _Identityvalidator.contract.UnpackLog(event, "CredentialRequirementAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IdentityvalidatorCredentialRequirementRemovedIterator is returned from FilterCredentialRequirementRemoved and is used to iterate over the raw logs and unpacked data for CredentialRequirementRemoved events raised by the Identityvalidator contract.
type IdentityvalidatorCredentialRequirementRemovedIterator struct {
	Event *IdentityvalidatorCredentialRequirementRemoved // Event containing the contract specifics and raw log

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
func (it *IdentityvalidatorCredentialRequirementRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityvalidatorCredentialRequirementRemoved)
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
		it.Event = new(IdentityvalidatorCredentialRequirementRemoved)
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
func (it *IdentityvalidatorCredentialRequirementRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityvalidatorCredentialRequirementRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityvalidatorCredentialRequirementRemoved represents a CredentialRequirementRemoved event raised by the Identityvalidator contract.
type IdentityvalidatorCredentialRequirementRemoved struct {
	RequirementId     [32]byte
	CredentialTypeIds [][32]byte
	MinValidations    *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterCredentialRequirementRemoved is a free log retrieval operation binding the contract event 0xd31ba7aeb0964621ef60b025c72e30a8816a36274b78ffd549764f87e453e112.
//
// Solidity: event CredentialRequirementRemoved(bytes32 indexed requirementId, bytes32[] credentialTypeIds, uint256 minValidations)
func (_Identityvalidator *IdentityvalidatorFilterer) FilterCredentialRequirementRemoved(opts *bind.FilterOpts, requirementId [][32]byte) (*IdentityvalidatorCredentialRequirementRemovedIterator, error) {

	var requirementIdRule []interface{}
	for _, requirementIdItem := range requirementId {
		requirementIdRule = append(requirementIdRule, requirementIdItem)
	}

	logs, sub, err := _Identityvalidator.contract.FilterLogs(opts, "CredentialRequirementRemoved", requirementIdRule)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorCredentialRequirementRemovedIterator{contract: _Identityvalidator.contract, event: "CredentialRequirementRemoved", logs: logs, sub: sub}, nil
}

// WatchCredentialRequirementRemoved is a free log subscription operation binding the contract event 0xd31ba7aeb0964621ef60b025c72e30a8816a36274b78ffd549764f87e453e112.
//
// Solidity: event CredentialRequirementRemoved(bytes32 indexed requirementId, bytes32[] credentialTypeIds, uint256 minValidations)
func (_Identityvalidator *IdentityvalidatorFilterer) WatchCredentialRequirementRemoved(opts *bind.WatchOpts, sink chan<- *IdentityvalidatorCredentialRequirementRemoved, requirementId [][32]byte) (event.Subscription, error) {

	var requirementIdRule []interface{}
	for _, requirementIdItem := range requirementId {
		requirementIdRule = append(requirementIdRule, requirementIdItem)
	}

	logs, sub, err := _Identityvalidator.contract.WatchLogs(opts, "CredentialRequirementRemoved", requirementIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityvalidatorCredentialRequirementRemoved)
				if err := _Identityvalidator.contract.UnpackLog(event, "CredentialRequirementRemoved", log); err != nil {
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

// ParseCredentialRequirementRemoved is a log parse operation binding the contract event 0xd31ba7aeb0964621ef60b025c72e30a8816a36274b78ffd549764f87e453e112.
//
// Solidity: event CredentialRequirementRemoved(bytes32 indexed requirementId, bytes32[] credentialTypeIds, uint256 minValidations)
func (_Identityvalidator *IdentityvalidatorFilterer) ParseCredentialRequirementRemoved(log types.Log) (*IdentityvalidatorCredentialRequirementRemoved, error) {
	event := new(IdentityvalidatorCredentialRequirementRemoved)
	if err := _Identityvalidator.contract.UnpackLog(event, "CredentialRequirementRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IdentityvalidatorCredentialSourceAddedIterator is returned from FilterCredentialSourceAdded and is used to iterate over the raw logs and unpacked data for CredentialSourceAdded events raised by the Identityvalidator contract.
type IdentityvalidatorCredentialSourceAddedIterator struct {
	Event *IdentityvalidatorCredentialSourceAdded // Event containing the contract specifics and raw log

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
func (it *IdentityvalidatorCredentialSourceAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityvalidatorCredentialSourceAdded)
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
		it.Event = new(IdentityvalidatorCredentialSourceAdded)
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
func (it *IdentityvalidatorCredentialSourceAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityvalidatorCredentialSourceAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityvalidatorCredentialSourceAdded represents a CredentialSourceAdded event raised by the Identityvalidator contract.
type IdentityvalidatorCredentialSourceAdded struct {
	CredentialTypeId   [32]byte
	IdentityRegistry   common.Address
	CredentialRegistry common.Address
	DataValidator      common.Address
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterCredentialSourceAdded is a free log retrieval operation binding the contract event 0x62bbab76853dfe54818ecdeaca8c73b87b7d3266439a265109815d909ed277c9.
//
// Solidity: event CredentialSourceAdded(bytes32 indexed credentialTypeId, address indexed identityRegistry, address indexed credentialRegistry, address dataValidator)
func (_Identityvalidator *IdentityvalidatorFilterer) FilterCredentialSourceAdded(opts *bind.FilterOpts, credentialTypeId [][32]byte, identityRegistry []common.Address, credentialRegistry []common.Address) (*IdentityvalidatorCredentialSourceAddedIterator, error) {

	var credentialTypeIdRule []interface{}
	for _, credentialTypeIdItem := range credentialTypeId {
		credentialTypeIdRule = append(credentialTypeIdRule, credentialTypeIdItem)
	}
	var identityRegistryRule []interface{}
	for _, identityRegistryItem := range identityRegistry {
		identityRegistryRule = append(identityRegistryRule, identityRegistryItem)
	}
	var credentialRegistryRule []interface{}
	for _, credentialRegistryItem := range credentialRegistry {
		credentialRegistryRule = append(credentialRegistryRule, credentialRegistryItem)
	}

	logs, sub, err := _Identityvalidator.contract.FilterLogs(opts, "CredentialSourceAdded", credentialTypeIdRule, identityRegistryRule, credentialRegistryRule)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorCredentialSourceAddedIterator{contract: _Identityvalidator.contract, event: "CredentialSourceAdded", logs: logs, sub: sub}, nil
}

// WatchCredentialSourceAdded is a free log subscription operation binding the contract event 0x62bbab76853dfe54818ecdeaca8c73b87b7d3266439a265109815d909ed277c9.
//
// Solidity: event CredentialSourceAdded(bytes32 indexed credentialTypeId, address indexed identityRegistry, address indexed credentialRegistry, address dataValidator)
func (_Identityvalidator *IdentityvalidatorFilterer) WatchCredentialSourceAdded(opts *bind.WatchOpts, sink chan<- *IdentityvalidatorCredentialSourceAdded, credentialTypeId [][32]byte, identityRegistry []common.Address, credentialRegistry []common.Address) (event.Subscription, error) {

	var credentialTypeIdRule []interface{}
	for _, credentialTypeIdItem := range credentialTypeId {
		credentialTypeIdRule = append(credentialTypeIdRule, credentialTypeIdItem)
	}
	var identityRegistryRule []interface{}
	for _, identityRegistryItem := range identityRegistry {
		identityRegistryRule = append(identityRegistryRule, identityRegistryItem)
	}
	var credentialRegistryRule []interface{}
	for _, credentialRegistryItem := range credentialRegistry {
		credentialRegistryRule = append(credentialRegistryRule, credentialRegistryItem)
	}

	logs, sub, err := _Identityvalidator.contract.WatchLogs(opts, "CredentialSourceAdded", credentialTypeIdRule, identityRegistryRule, credentialRegistryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityvalidatorCredentialSourceAdded)
				if err := _Identityvalidator.contract.UnpackLog(event, "CredentialSourceAdded", log); err != nil {
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

// ParseCredentialSourceAdded is a log parse operation binding the contract event 0x62bbab76853dfe54818ecdeaca8c73b87b7d3266439a265109815d909ed277c9.
//
// Solidity: event CredentialSourceAdded(bytes32 indexed credentialTypeId, address indexed identityRegistry, address indexed credentialRegistry, address dataValidator)
func (_Identityvalidator *IdentityvalidatorFilterer) ParseCredentialSourceAdded(log types.Log) (*IdentityvalidatorCredentialSourceAdded, error) {
	event := new(IdentityvalidatorCredentialSourceAdded)
	if err := _Identityvalidator.contract.UnpackLog(event, "CredentialSourceAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IdentityvalidatorCredentialSourceRemovedIterator is returned from FilterCredentialSourceRemoved and is used to iterate over the raw logs and unpacked data for CredentialSourceRemoved events raised by the Identityvalidator contract.
type IdentityvalidatorCredentialSourceRemovedIterator struct {
	Event *IdentityvalidatorCredentialSourceRemoved // Event containing the contract specifics and raw log

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
func (it *IdentityvalidatorCredentialSourceRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityvalidatorCredentialSourceRemoved)
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
		it.Event = new(IdentityvalidatorCredentialSourceRemoved)
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
func (it *IdentityvalidatorCredentialSourceRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityvalidatorCredentialSourceRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityvalidatorCredentialSourceRemoved represents a CredentialSourceRemoved event raised by the Identityvalidator contract.
type IdentityvalidatorCredentialSourceRemoved struct {
	CredentialTypeId   [32]byte
	IdentityRegistry   common.Address
	CredentialRegistry common.Address
	DataValidator      common.Address
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterCredentialSourceRemoved is a free log retrieval operation binding the contract event 0x9c6217b18044e22e3ba28f6713b159e1fa7b0dc0340eb48512432fdfc4650705.
//
// Solidity: event CredentialSourceRemoved(bytes32 indexed credentialTypeId, address indexed identityRegistry, address indexed credentialRegistry, address dataValidator)
func (_Identityvalidator *IdentityvalidatorFilterer) FilterCredentialSourceRemoved(opts *bind.FilterOpts, credentialTypeId [][32]byte, identityRegistry []common.Address, credentialRegistry []common.Address) (*IdentityvalidatorCredentialSourceRemovedIterator, error) {

	var credentialTypeIdRule []interface{}
	for _, credentialTypeIdItem := range credentialTypeId {
		credentialTypeIdRule = append(credentialTypeIdRule, credentialTypeIdItem)
	}
	var identityRegistryRule []interface{}
	for _, identityRegistryItem := range identityRegistry {
		identityRegistryRule = append(identityRegistryRule, identityRegistryItem)
	}
	var credentialRegistryRule []interface{}
	for _, credentialRegistryItem := range credentialRegistry {
		credentialRegistryRule = append(credentialRegistryRule, credentialRegistryItem)
	}

	logs, sub, err := _Identityvalidator.contract.FilterLogs(opts, "CredentialSourceRemoved", credentialTypeIdRule, identityRegistryRule, credentialRegistryRule)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorCredentialSourceRemovedIterator{contract: _Identityvalidator.contract, event: "CredentialSourceRemoved", logs: logs, sub: sub}, nil
}

// WatchCredentialSourceRemoved is a free log subscription operation binding the contract event 0x9c6217b18044e22e3ba28f6713b159e1fa7b0dc0340eb48512432fdfc4650705.
//
// Solidity: event CredentialSourceRemoved(bytes32 indexed credentialTypeId, address indexed identityRegistry, address indexed credentialRegistry, address dataValidator)
func (_Identityvalidator *IdentityvalidatorFilterer) WatchCredentialSourceRemoved(opts *bind.WatchOpts, sink chan<- *IdentityvalidatorCredentialSourceRemoved, credentialTypeId [][32]byte, identityRegistry []common.Address, credentialRegistry []common.Address) (event.Subscription, error) {

	var credentialTypeIdRule []interface{}
	for _, credentialTypeIdItem := range credentialTypeId {
		credentialTypeIdRule = append(credentialTypeIdRule, credentialTypeIdItem)
	}
	var identityRegistryRule []interface{}
	for _, identityRegistryItem := range identityRegistry {
		identityRegistryRule = append(identityRegistryRule, identityRegistryItem)
	}
	var credentialRegistryRule []interface{}
	for _, credentialRegistryItem := range credentialRegistry {
		credentialRegistryRule = append(credentialRegistryRule, credentialRegistryItem)
	}

	logs, sub, err := _Identityvalidator.contract.WatchLogs(opts, "CredentialSourceRemoved", credentialTypeIdRule, identityRegistryRule, credentialRegistryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityvalidatorCredentialSourceRemoved)
				if err := _Identityvalidator.contract.UnpackLog(event, "CredentialSourceRemoved", log); err != nil {
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

// ParseCredentialSourceRemoved is a log parse operation binding the contract event 0x9c6217b18044e22e3ba28f6713b159e1fa7b0dc0340eb48512432fdfc4650705.
//
// Solidity: event CredentialSourceRemoved(bytes32 indexed credentialTypeId, address indexed identityRegistry, address indexed credentialRegistry, address dataValidator)
func (_Identityvalidator *IdentityvalidatorFilterer) ParseCredentialSourceRemoved(log types.Log) (*IdentityvalidatorCredentialSourceRemoved, error) {
	event := new(IdentityvalidatorCredentialSourceRemoved)
	if err := _Identityvalidator.contract.UnpackLog(event, "CredentialSourceRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IdentityvalidatorInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Identityvalidator contract.
type IdentityvalidatorInitializedIterator struct {
	Event *IdentityvalidatorInitialized // Event containing the contract specifics and raw log

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
func (it *IdentityvalidatorInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityvalidatorInitialized)
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
		it.Event = new(IdentityvalidatorInitialized)
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
func (it *IdentityvalidatorInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityvalidatorInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityvalidatorInitialized represents a Initialized event raised by the Identityvalidator contract.
type IdentityvalidatorInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Identityvalidator *IdentityvalidatorFilterer) FilterInitialized(opts *bind.FilterOpts) (*IdentityvalidatorInitializedIterator, error) {

	logs, sub, err := _Identityvalidator.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorInitializedIterator{contract: _Identityvalidator.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Identityvalidator *IdentityvalidatorFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *IdentityvalidatorInitialized) (event.Subscription, error) {

	logs, sub, err := _Identityvalidator.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityvalidatorInitialized)
				if err := _Identityvalidator.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_Identityvalidator *IdentityvalidatorFilterer) ParseInitialized(log types.Log) (*IdentityvalidatorInitialized, error) {
	event := new(IdentityvalidatorInitialized)
	if err := _Identityvalidator.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IdentityvalidatorOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Identityvalidator contract.
type IdentityvalidatorOwnershipTransferredIterator struct {
	Event *IdentityvalidatorOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *IdentityvalidatorOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityvalidatorOwnershipTransferred)
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
		it.Event = new(IdentityvalidatorOwnershipTransferred)
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
func (it *IdentityvalidatorOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityvalidatorOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityvalidatorOwnershipTransferred represents a OwnershipTransferred event raised by the Identityvalidator contract.
type IdentityvalidatorOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Identityvalidator *IdentityvalidatorFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*IdentityvalidatorOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Identityvalidator.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorOwnershipTransferredIterator{contract: _Identityvalidator.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Identityvalidator *IdentityvalidatorFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *IdentityvalidatorOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Identityvalidator.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityvalidatorOwnershipTransferred)
				if err := _Identityvalidator.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Identityvalidator *IdentityvalidatorFilterer) ParseOwnershipTransferred(log types.Log) (*IdentityvalidatorOwnershipTransferred, error) {
	event := new(IdentityvalidatorOwnershipTransferred)
	if err := _Identityvalidator.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
