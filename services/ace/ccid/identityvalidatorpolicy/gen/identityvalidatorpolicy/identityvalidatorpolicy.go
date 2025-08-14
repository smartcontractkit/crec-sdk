// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package identityvalidatorpolicy

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

// IdentityvalidatorpolicyMetaData contains all meta data concerning the Identityvalidatorpolicy contract.
var IdentityvalidatorpolicyMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"addCredentialRequirement\",\"inputs\":[{\"name\":\"input\",\"type\":\"tuple\",\"internalType\":\"structICredentialRequirements.CredentialRequirementInput\",\"components\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeIds\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"minValidations\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"invert\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"addCredentialSource\",\"inputs\":[{\"name\":\"input\",\"type\":\"tuple\",\"internalType\":\"structICredentialRequirements.CredentialSourceInput\",\"components\":[{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"identityRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"dataValidator\",\"type\":\"address\",\"internalType\":\"address\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getCredentialRequirement\",\"inputs\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structICredentialRequirements.CredentialRequirement\",\"components\":[{\"name\":\"credentialTypeIds\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"minValidations\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"invert\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCredentialRequirementIds\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCredentialSources\",\"inputs\":[{\"name\":\"credential\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structICredentialRequirements.CredentialSource[]\",\"components\":[{\"name\":\"identityRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"dataValidator\",\"type\":\"address\",\"internalType\":\"address\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"credentialSourceInputs\",\"type\":\"tuple[]\",\"internalType\":\"structICredentialRequirements.CredentialSourceInput[]\",\"components\":[{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"identityRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"dataValidator\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"name\":\"credentialRequirementInputs\",\"type\":\"tuple[]\",\"internalType\":\"structICredentialRequirements.CredentialRequirementInput[]\",\"components\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeIds\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"minValidations\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"invert\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"policyEngine\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"initialOwner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"configParams\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"onInstall\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"onUninstall\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"postRun\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"},{\"name\":\"\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"removeCredentialRequirement\",\"inputs\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"removeCredentialSource\",\"inputs\":[{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"identityRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"run\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"},{\"name\":\"parameters\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"},{\"name\":\"context\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumIPolicyEngine.PolicyResult\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"validate\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"context\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"CredentialRequirementAdded\",\"inputs\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeIds\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"minValidations\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CredentialRequirementRemoved\",\"inputs\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"credentialTypeIds\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"minValidations\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CredentialSourceAdded\",\"inputs\":[{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"identityRegistry\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"dataValidator\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CredentialSourceRemoved\",\"inputs\":[{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"identityRegistry\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"dataValidator\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"InvalidConfiguration\",\"inputs\":[{\"name\":\"errorReason\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"InvalidConfiguration\",\"inputs\":[{\"name\":\"errorReason\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"RequirementExists\",\"inputs\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"RequirementNotFound\",\"inputs\":[{\"name\":\"requirementId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"SourceExists\",\"inputs\":[{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"identityRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"SourceNotFound\",\"inputs\":[{\"name\":\"credentialTypeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"identityRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"credentialRegistry\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"Unauthorized\",\"inputs\":[]}]",
}

// IdentityvalidatorpolicyABI is the input ABI used to generate the binding from.
// Deprecated: Use IdentityvalidatorpolicyMetaData.ABI instead.
var IdentityvalidatorpolicyABI = IdentityvalidatorpolicyMetaData.ABI

// Identityvalidatorpolicy is an auto generated Go binding around an Ethereum contract.
type Identityvalidatorpolicy struct {
	IdentityvalidatorpolicyCaller     // Read-only binding to the contract
	IdentityvalidatorpolicyTransactor // Write-only binding to the contract
	IdentityvalidatorpolicyFilterer   // Log filterer for contract events
}

// IdentityvalidatorpolicyCaller is an auto generated read-only Go binding around an Ethereum contract.
type IdentityvalidatorpolicyCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IdentityvalidatorpolicyTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IdentityvalidatorpolicyTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IdentityvalidatorpolicyFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IdentityvalidatorpolicyFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IdentityvalidatorpolicySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IdentityvalidatorpolicySession struct {
	Contract     *Identityvalidatorpolicy // Generic contract binding to set the session for
	CallOpts     bind.CallOpts            // Call options to use throughout this session
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// IdentityvalidatorpolicyCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IdentityvalidatorpolicyCallerSession struct {
	Contract *IdentityvalidatorpolicyCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                  // Call options to use throughout this session
}

// IdentityvalidatorpolicyTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IdentityvalidatorpolicyTransactorSession struct {
	Contract     *IdentityvalidatorpolicyTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                  // Transaction auth options to use throughout this session
}

// IdentityvalidatorpolicyRaw is an auto generated low-level Go binding around an Ethereum contract.
type IdentityvalidatorpolicyRaw struct {
	Contract *Identityvalidatorpolicy // Generic contract binding to access the raw methods on
}

// IdentityvalidatorpolicyCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IdentityvalidatorpolicyCallerRaw struct {
	Contract *IdentityvalidatorpolicyCaller // Generic read-only contract binding to access the raw methods on
}

// IdentityvalidatorpolicyTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IdentityvalidatorpolicyTransactorRaw struct {
	Contract *IdentityvalidatorpolicyTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIdentityvalidatorpolicy creates a new instance of Identityvalidatorpolicy, bound to a specific deployed contract.
func NewIdentityvalidatorpolicy(address common.Address, backend bind.ContractBackend) (*Identityvalidatorpolicy, error) {
	contract, err := bindIdentityvalidatorpolicy(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Identityvalidatorpolicy{IdentityvalidatorpolicyCaller: IdentityvalidatorpolicyCaller{contract: contract}, IdentityvalidatorpolicyTransactor: IdentityvalidatorpolicyTransactor{contract: contract}, IdentityvalidatorpolicyFilterer: IdentityvalidatorpolicyFilterer{contract: contract}}, nil
}

// NewIdentityvalidatorpolicyCaller creates a new read-only instance of Identityvalidatorpolicy, bound to a specific deployed contract.
func NewIdentityvalidatorpolicyCaller(address common.Address, caller bind.ContractCaller) (*IdentityvalidatorpolicyCaller, error) {
	contract, err := bindIdentityvalidatorpolicy(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorpolicyCaller{contract: contract}, nil
}

// NewIdentityvalidatorpolicyTransactor creates a new write-only instance of Identityvalidatorpolicy, bound to a specific deployed contract.
func NewIdentityvalidatorpolicyTransactor(address common.Address, transactor bind.ContractTransactor) (*IdentityvalidatorpolicyTransactor, error) {
	contract, err := bindIdentityvalidatorpolicy(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorpolicyTransactor{contract: contract}, nil
}

// NewIdentityvalidatorpolicyFilterer creates a new log filterer instance of Identityvalidatorpolicy, bound to a specific deployed contract.
func NewIdentityvalidatorpolicyFilterer(address common.Address, filterer bind.ContractFilterer) (*IdentityvalidatorpolicyFilterer, error) {
	contract, err := bindIdentityvalidatorpolicy(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorpolicyFilterer{contract: contract}, nil
}

// bindIdentityvalidatorpolicy binds a generic wrapper to an already deployed contract.
func bindIdentityvalidatorpolicy(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IdentityvalidatorpolicyMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Identityvalidatorpolicy.Contract.IdentityvalidatorpolicyCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.IdentityvalidatorpolicyTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.IdentityvalidatorpolicyTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Identityvalidatorpolicy.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.contract.Transact(opts, method, params...)
}

// GetCredentialRequirement is a free data retrieval call binding the contract method 0x27fbdc39.
//
// Solidity: function getCredentialRequirement(bytes32 requirementId) view returns((bytes32[],uint256,bool))
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCaller) GetCredentialRequirement(opts *bind.CallOpts, requirementId [32]byte) (ICredentialRequirementsCredentialRequirement, error) {
	var out []interface{}
	err := _Identityvalidatorpolicy.contract.Call(opts, &out, "getCredentialRequirement", requirementId)

	if err != nil {
		return *new(ICredentialRequirementsCredentialRequirement), err
	}

	out0 := *abi.ConvertType(out[0], new(ICredentialRequirementsCredentialRequirement)).(*ICredentialRequirementsCredentialRequirement)

	return out0, err

}

// GetCredentialRequirement is a free data retrieval call binding the contract method 0x27fbdc39.
//
// Solidity: function getCredentialRequirement(bytes32 requirementId) view returns((bytes32[],uint256,bool))
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) GetCredentialRequirement(requirementId [32]byte) (ICredentialRequirementsCredentialRequirement, error) {
	return _Identityvalidatorpolicy.Contract.GetCredentialRequirement(&_Identityvalidatorpolicy.CallOpts, requirementId)
}

// GetCredentialRequirement is a free data retrieval call binding the contract method 0x27fbdc39.
//
// Solidity: function getCredentialRequirement(bytes32 requirementId) view returns((bytes32[],uint256,bool))
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCallerSession) GetCredentialRequirement(requirementId [32]byte) (ICredentialRequirementsCredentialRequirement, error) {
	return _Identityvalidatorpolicy.Contract.GetCredentialRequirement(&_Identityvalidatorpolicy.CallOpts, requirementId)
}

// GetCredentialRequirementIds is a free data retrieval call binding the contract method 0x7fab8711.
//
// Solidity: function getCredentialRequirementIds() view returns(bytes32[])
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCaller) GetCredentialRequirementIds(opts *bind.CallOpts) ([][32]byte, error) {
	var out []interface{}
	err := _Identityvalidatorpolicy.contract.Call(opts, &out, "getCredentialRequirementIds")

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetCredentialRequirementIds is a free data retrieval call binding the contract method 0x7fab8711.
//
// Solidity: function getCredentialRequirementIds() view returns(bytes32[])
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) GetCredentialRequirementIds() ([][32]byte, error) {
	return _Identityvalidatorpolicy.Contract.GetCredentialRequirementIds(&_Identityvalidatorpolicy.CallOpts)
}

// GetCredentialRequirementIds is a free data retrieval call binding the contract method 0x7fab8711.
//
// Solidity: function getCredentialRequirementIds() view returns(bytes32[])
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCallerSession) GetCredentialRequirementIds() ([][32]byte, error) {
	return _Identityvalidatorpolicy.Contract.GetCredentialRequirementIds(&_Identityvalidatorpolicy.CallOpts)
}

// GetCredentialSources is a free data retrieval call binding the contract method 0x04628d5d.
//
// Solidity: function getCredentialSources(bytes32 credential) view returns((address,address,address)[])
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCaller) GetCredentialSources(opts *bind.CallOpts, credential [32]byte) ([]ICredentialRequirementsCredentialSource, error) {
	var out []interface{}
	err := _Identityvalidatorpolicy.contract.Call(opts, &out, "getCredentialSources", credential)

	if err != nil {
		return *new([]ICredentialRequirementsCredentialSource), err
	}

	out0 := *abi.ConvertType(out[0], new([]ICredentialRequirementsCredentialSource)).(*[]ICredentialRequirementsCredentialSource)

	return out0, err

}

// GetCredentialSources is a free data retrieval call binding the contract method 0x04628d5d.
//
// Solidity: function getCredentialSources(bytes32 credential) view returns((address,address,address)[])
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) GetCredentialSources(credential [32]byte) ([]ICredentialRequirementsCredentialSource, error) {
	return _Identityvalidatorpolicy.Contract.GetCredentialSources(&_Identityvalidatorpolicy.CallOpts, credential)
}

// GetCredentialSources is a free data retrieval call binding the contract method 0x04628d5d.
//
// Solidity: function getCredentialSources(bytes32 credential) view returns((address,address,address)[])
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCallerSession) GetCredentialSources(credential [32]byte) ([]ICredentialRequirementsCredentialSource, error) {
	return _Identityvalidatorpolicy.Contract.GetCredentialSources(&_Identityvalidatorpolicy.CallOpts, credential)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Identityvalidatorpolicy.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) Owner() (common.Address, error) {
	return _Identityvalidatorpolicy.Contract.Owner(&_Identityvalidatorpolicy.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCallerSession) Owner() (common.Address, error) {
	return _Identityvalidatorpolicy.Contract.Owner(&_Identityvalidatorpolicy.CallOpts)
}

// Run is a free data retrieval call binding the contract method 0xe3e3094e.
//
// Solidity: function run(address , address , bytes4 , bytes[] parameters, bytes context) view returns(uint8)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCaller) Run(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address, arg2 [4]byte, parameters [][]byte, context []byte) (uint8, error) {
	var out []interface{}
	err := _Identityvalidatorpolicy.contract.Call(opts, &out, "run", arg0, arg1, arg2, parameters, context)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Run is a free data retrieval call binding the contract method 0xe3e3094e.
//
// Solidity: function run(address , address , bytes4 , bytes[] parameters, bytes context) view returns(uint8)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) Run(arg0 common.Address, arg1 common.Address, arg2 [4]byte, parameters [][]byte, context []byte) (uint8, error) {
	return _Identityvalidatorpolicy.Contract.Run(&_Identityvalidatorpolicy.CallOpts, arg0, arg1, arg2, parameters, context)
}

// Run is a free data retrieval call binding the contract method 0xe3e3094e.
//
// Solidity: function run(address , address , bytes4 , bytes[] parameters, bytes context) view returns(uint8)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCallerSession) Run(arg0 common.Address, arg1 common.Address, arg2 [4]byte, parameters [][]byte, context []byte) (uint8, error) {
	return _Identityvalidatorpolicy.Contract.Run(&_Identityvalidatorpolicy.CallOpts, arg0, arg1, arg2, parameters, context)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Identityvalidatorpolicy.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Identityvalidatorpolicy.Contract.SupportsInterface(&_Identityvalidatorpolicy.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Identityvalidatorpolicy.Contract.SupportsInterface(&_Identityvalidatorpolicy.CallOpts, interfaceId)
}

// Validate is a free data retrieval call binding the contract method 0xcaf92785.
//
// Solidity: function validate(address account, bytes context) view returns(bool)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCaller) Validate(opts *bind.CallOpts, account common.Address, context []byte) (bool, error) {
	var out []interface{}
	err := _Identityvalidatorpolicy.contract.Call(opts, &out, "validate", account, context)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Validate is a free data retrieval call binding the contract method 0xcaf92785.
//
// Solidity: function validate(address account, bytes context) view returns(bool)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) Validate(account common.Address, context []byte) (bool, error) {
	return _Identityvalidatorpolicy.Contract.Validate(&_Identityvalidatorpolicy.CallOpts, account, context)
}

// Validate is a free data retrieval call binding the contract method 0xcaf92785.
//
// Solidity: function validate(address account, bytes context) view returns(bool)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyCallerSession) Validate(account common.Address, context []byte) (bool, error) {
	return _Identityvalidatorpolicy.Contract.Validate(&_Identityvalidatorpolicy.CallOpts, account, context)
}

// AddCredentialRequirement is a paid mutator transaction binding the contract method 0x4afbe01e.
//
// Solidity: function addCredentialRequirement((bytes32,bytes32[],uint256,bool) input) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactor) AddCredentialRequirement(opts *bind.TransactOpts, input ICredentialRequirementsCredentialRequirementInput) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.contract.Transact(opts, "addCredentialRequirement", input)
}

// AddCredentialRequirement is a paid mutator transaction binding the contract method 0x4afbe01e.
//
// Solidity: function addCredentialRequirement((bytes32,bytes32[],uint256,bool) input) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) AddCredentialRequirement(input ICredentialRequirementsCredentialRequirementInput) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.AddCredentialRequirement(&_Identityvalidatorpolicy.TransactOpts, input)
}

// AddCredentialRequirement is a paid mutator transaction binding the contract method 0x4afbe01e.
//
// Solidity: function addCredentialRequirement((bytes32,bytes32[],uint256,bool) input) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactorSession) AddCredentialRequirement(input ICredentialRequirementsCredentialRequirementInput) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.AddCredentialRequirement(&_Identityvalidatorpolicy.TransactOpts, input)
}

// AddCredentialSource is a paid mutator transaction binding the contract method 0xe09a5423.
//
// Solidity: function addCredentialSource((bytes32,address,address,address) input) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactor) AddCredentialSource(opts *bind.TransactOpts, input ICredentialRequirementsCredentialSourceInput) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.contract.Transact(opts, "addCredentialSource", input)
}

// AddCredentialSource is a paid mutator transaction binding the contract method 0xe09a5423.
//
// Solidity: function addCredentialSource((bytes32,address,address,address) input) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) AddCredentialSource(input ICredentialRequirementsCredentialSourceInput) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.AddCredentialSource(&_Identityvalidatorpolicy.TransactOpts, input)
}

// AddCredentialSource is a paid mutator transaction binding the contract method 0xe09a5423.
//
// Solidity: function addCredentialSource((bytes32,address,address,address) input) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactorSession) AddCredentialSource(input ICredentialRequirementsCredentialSourceInput) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.AddCredentialSource(&_Identityvalidatorpolicy.TransactOpts, input)
}

// Initialize is a paid mutator transaction binding the contract method 0x0fe13086.
//
// Solidity: function initialize((bytes32,address,address,address)[] credentialSourceInputs, (bytes32,bytes32[],uint256,bool)[] credentialRequirementInputs) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactor) Initialize(opts *bind.TransactOpts, credentialSourceInputs []ICredentialRequirementsCredentialSourceInput, credentialRequirementInputs []ICredentialRequirementsCredentialRequirementInput) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.contract.Transact(opts, "initialize", credentialSourceInputs, credentialRequirementInputs)
}

// Initialize is a paid mutator transaction binding the contract method 0x0fe13086.
//
// Solidity: function initialize((bytes32,address,address,address)[] credentialSourceInputs, (bytes32,bytes32[],uint256,bool)[] credentialRequirementInputs) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) Initialize(credentialSourceInputs []ICredentialRequirementsCredentialSourceInput, credentialRequirementInputs []ICredentialRequirementsCredentialRequirementInput) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.Initialize(&_Identityvalidatorpolicy.TransactOpts, credentialSourceInputs, credentialRequirementInputs)
}

// Initialize is a paid mutator transaction binding the contract method 0x0fe13086.
//
// Solidity: function initialize((bytes32,address,address,address)[] credentialSourceInputs, (bytes32,bytes32[],uint256,bool)[] credentialRequirementInputs) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactorSession) Initialize(credentialSourceInputs []ICredentialRequirementsCredentialSourceInput, credentialRequirementInputs []ICredentialRequirementsCredentialRequirementInput) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.Initialize(&_Identityvalidatorpolicy.TransactOpts, credentialSourceInputs, credentialRequirementInputs)
}

// Initialize0 is a paid mutator transaction binding the contract method 0xcf7a1d77.
//
// Solidity: function initialize(address policyEngine, address initialOwner, bytes configParams) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactor) Initialize0(opts *bind.TransactOpts, policyEngine common.Address, initialOwner common.Address, configParams []byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.contract.Transact(opts, "initialize0", policyEngine, initialOwner, configParams)
}

// Initialize0 is a paid mutator transaction binding the contract method 0xcf7a1d77.
//
// Solidity: function initialize(address policyEngine, address initialOwner, bytes configParams) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) Initialize0(policyEngine common.Address, initialOwner common.Address, configParams []byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.Initialize0(&_Identityvalidatorpolicy.TransactOpts, policyEngine, initialOwner, configParams)
}

// Initialize0 is a paid mutator transaction binding the contract method 0xcf7a1d77.
//
// Solidity: function initialize(address policyEngine, address initialOwner, bytes configParams) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactorSession) Initialize0(policyEngine common.Address, initialOwner common.Address, configParams []byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.Initialize0(&_Identityvalidatorpolicy.TransactOpts, policyEngine, initialOwner, configParams)
}

// OnInstall is a paid mutator transaction binding the contract method 0x3aaf77c7.
//
// Solidity: function onInstall(bytes4 ) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactor) OnInstall(opts *bind.TransactOpts, arg0 [4]byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.contract.Transact(opts, "onInstall", arg0)
}

// OnInstall is a paid mutator transaction binding the contract method 0x3aaf77c7.
//
// Solidity: function onInstall(bytes4 ) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) OnInstall(arg0 [4]byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.OnInstall(&_Identityvalidatorpolicy.TransactOpts, arg0)
}

// OnInstall is a paid mutator transaction binding the contract method 0x3aaf77c7.
//
// Solidity: function onInstall(bytes4 ) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactorSession) OnInstall(arg0 [4]byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.OnInstall(&_Identityvalidatorpolicy.TransactOpts, arg0)
}

// OnUninstall is a paid mutator transaction binding the contract method 0x4e302170.
//
// Solidity: function onUninstall(bytes4 ) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactor) OnUninstall(opts *bind.TransactOpts, arg0 [4]byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.contract.Transact(opts, "onUninstall", arg0)
}

// OnUninstall is a paid mutator transaction binding the contract method 0x4e302170.
//
// Solidity: function onUninstall(bytes4 ) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) OnUninstall(arg0 [4]byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.OnUninstall(&_Identityvalidatorpolicy.TransactOpts, arg0)
}

// OnUninstall is a paid mutator transaction binding the contract method 0x4e302170.
//
// Solidity: function onUninstall(bytes4 ) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactorSession) OnUninstall(arg0 [4]byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.OnUninstall(&_Identityvalidatorpolicy.TransactOpts, arg0)
}

// PostRun is a paid mutator transaction binding the contract method 0x43747414.
//
// Solidity: function postRun(address , address , bytes4 , bytes[] , bytes ) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactor) PostRun(opts *bind.TransactOpts, arg0 common.Address, arg1 common.Address, arg2 [4]byte, arg3 [][]byte, arg4 []byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.contract.Transact(opts, "postRun", arg0, arg1, arg2, arg3, arg4)
}

// PostRun is a paid mutator transaction binding the contract method 0x43747414.
//
// Solidity: function postRun(address , address , bytes4 , bytes[] , bytes ) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) PostRun(arg0 common.Address, arg1 common.Address, arg2 [4]byte, arg3 [][]byte, arg4 []byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.PostRun(&_Identityvalidatorpolicy.TransactOpts, arg0, arg1, arg2, arg3, arg4)
}

// PostRun is a paid mutator transaction binding the contract method 0x43747414.
//
// Solidity: function postRun(address , address , bytes4 , bytes[] , bytes ) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactorSession) PostRun(arg0 common.Address, arg1 common.Address, arg2 [4]byte, arg3 [][]byte, arg4 []byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.PostRun(&_Identityvalidatorpolicy.TransactOpts, arg0, arg1, arg2, arg3, arg4)
}

// RemoveCredentialRequirement is a paid mutator transaction binding the contract method 0x6c904c42.
//
// Solidity: function removeCredentialRequirement(bytes32 requirementId) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactor) RemoveCredentialRequirement(opts *bind.TransactOpts, requirementId [32]byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.contract.Transact(opts, "removeCredentialRequirement", requirementId)
}

// RemoveCredentialRequirement is a paid mutator transaction binding the contract method 0x6c904c42.
//
// Solidity: function removeCredentialRequirement(bytes32 requirementId) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) RemoveCredentialRequirement(requirementId [32]byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.RemoveCredentialRequirement(&_Identityvalidatorpolicy.TransactOpts, requirementId)
}

// RemoveCredentialRequirement is a paid mutator transaction binding the contract method 0x6c904c42.
//
// Solidity: function removeCredentialRequirement(bytes32 requirementId) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactorSession) RemoveCredentialRequirement(requirementId [32]byte) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.RemoveCredentialRequirement(&_Identityvalidatorpolicy.TransactOpts, requirementId)
}

// RemoveCredentialSource is a paid mutator transaction binding the contract method 0x3578b00d.
//
// Solidity: function removeCredentialSource(bytes32 credentialTypeId, address identityRegistry, address credentialRegistry) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactor) RemoveCredentialSource(opts *bind.TransactOpts, credentialTypeId [32]byte, identityRegistry common.Address, credentialRegistry common.Address) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.contract.Transact(opts, "removeCredentialSource", credentialTypeId, identityRegistry, credentialRegistry)
}

// RemoveCredentialSource is a paid mutator transaction binding the contract method 0x3578b00d.
//
// Solidity: function removeCredentialSource(bytes32 credentialTypeId, address identityRegistry, address credentialRegistry) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) RemoveCredentialSource(credentialTypeId [32]byte, identityRegistry common.Address, credentialRegistry common.Address) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.RemoveCredentialSource(&_Identityvalidatorpolicy.TransactOpts, credentialTypeId, identityRegistry, credentialRegistry)
}

// RemoveCredentialSource is a paid mutator transaction binding the contract method 0x3578b00d.
//
// Solidity: function removeCredentialSource(bytes32 credentialTypeId, address identityRegistry, address credentialRegistry) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactorSession) RemoveCredentialSource(credentialTypeId [32]byte, identityRegistry common.Address, credentialRegistry common.Address) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.RemoveCredentialSource(&_Identityvalidatorpolicy.TransactOpts, credentialTypeId, identityRegistry, credentialRegistry)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) RenounceOwnership() (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.RenounceOwnership(&_Identityvalidatorpolicy.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.RenounceOwnership(&_Identityvalidatorpolicy.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.TransferOwnership(&_Identityvalidatorpolicy.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Identityvalidatorpolicy.Contract.TransferOwnership(&_Identityvalidatorpolicy.TransactOpts, newOwner)
}

// IdentityvalidatorpolicyCredentialRequirementAddedIterator is returned from FilterCredentialRequirementAdded and is used to iterate over the raw logs and unpacked data for CredentialRequirementAdded events raised by the Identityvalidatorpolicy contract.
type IdentityvalidatorpolicyCredentialRequirementAddedIterator struct {
	Event *IdentityvalidatorpolicyCredentialRequirementAdded // Event containing the contract specifics and raw log

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
func (it *IdentityvalidatorpolicyCredentialRequirementAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityvalidatorpolicyCredentialRequirementAdded)
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
		it.Event = new(IdentityvalidatorpolicyCredentialRequirementAdded)
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
func (it *IdentityvalidatorpolicyCredentialRequirementAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityvalidatorpolicyCredentialRequirementAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityvalidatorpolicyCredentialRequirementAdded represents a CredentialRequirementAdded event raised by the Identityvalidatorpolicy contract.
type IdentityvalidatorpolicyCredentialRequirementAdded struct {
	RequirementId     [32]byte
	CredentialTypeIds [][32]byte
	MinValidations    *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterCredentialRequirementAdded is a free log retrieval operation binding the contract event 0x5e62bddfd91274dfd4cdbfe7ca9f7dbd770b0eb6c57e2b39b5c412a6b44fb010.
//
// Solidity: event CredentialRequirementAdded(bytes32 indexed requirementId, bytes32[] credentialTypeIds, uint256 minValidations)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) FilterCredentialRequirementAdded(opts *bind.FilterOpts, requirementId [][32]byte) (*IdentityvalidatorpolicyCredentialRequirementAddedIterator, error) {

	var requirementIdRule []interface{}
	for _, requirementIdItem := range requirementId {
		requirementIdRule = append(requirementIdRule, requirementIdItem)
	}

	logs, sub, err := _Identityvalidatorpolicy.contract.FilterLogs(opts, "CredentialRequirementAdded", requirementIdRule)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorpolicyCredentialRequirementAddedIterator{contract: _Identityvalidatorpolicy.contract, event: "CredentialRequirementAdded", logs: logs, sub: sub}, nil
}

// WatchCredentialRequirementAdded is a free log subscription operation binding the contract event 0x5e62bddfd91274dfd4cdbfe7ca9f7dbd770b0eb6c57e2b39b5c412a6b44fb010.
//
// Solidity: event CredentialRequirementAdded(bytes32 indexed requirementId, bytes32[] credentialTypeIds, uint256 minValidations)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) WatchCredentialRequirementAdded(opts *bind.WatchOpts, sink chan<- *IdentityvalidatorpolicyCredentialRequirementAdded, requirementId [][32]byte) (event.Subscription, error) {

	var requirementIdRule []interface{}
	for _, requirementIdItem := range requirementId {
		requirementIdRule = append(requirementIdRule, requirementIdItem)
	}

	logs, sub, err := _Identityvalidatorpolicy.contract.WatchLogs(opts, "CredentialRequirementAdded", requirementIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityvalidatorpolicyCredentialRequirementAdded)
				if err := _Identityvalidatorpolicy.contract.UnpackLog(event, "CredentialRequirementAdded", log); err != nil {
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
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) ParseCredentialRequirementAdded(log types.Log) (*IdentityvalidatorpolicyCredentialRequirementAdded, error) {
	event := new(IdentityvalidatorpolicyCredentialRequirementAdded)
	if err := _Identityvalidatorpolicy.contract.UnpackLog(event, "CredentialRequirementAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IdentityvalidatorpolicyCredentialRequirementRemovedIterator is returned from FilterCredentialRequirementRemoved and is used to iterate over the raw logs and unpacked data for CredentialRequirementRemoved events raised by the Identityvalidatorpolicy contract.
type IdentityvalidatorpolicyCredentialRequirementRemovedIterator struct {
	Event *IdentityvalidatorpolicyCredentialRequirementRemoved // Event containing the contract specifics and raw log

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
func (it *IdentityvalidatorpolicyCredentialRequirementRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityvalidatorpolicyCredentialRequirementRemoved)
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
		it.Event = new(IdentityvalidatorpolicyCredentialRequirementRemoved)
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
func (it *IdentityvalidatorpolicyCredentialRequirementRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityvalidatorpolicyCredentialRequirementRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityvalidatorpolicyCredentialRequirementRemoved represents a CredentialRequirementRemoved event raised by the Identityvalidatorpolicy contract.
type IdentityvalidatorpolicyCredentialRequirementRemoved struct {
	RequirementId     [32]byte
	CredentialTypeIds [][32]byte
	MinValidations    *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterCredentialRequirementRemoved is a free log retrieval operation binding the contract event 0xd31ba7aeb0964621ef60b025c72e30a8816a36274b78ffd549764f87e453e112.
//
// Solidity: event CredentialRequirementRemoved(bytes32 indexed requirementId, bytes32[] credentialTypeIds, uint256 minValidations)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) FilterCredentialRequirementRemoved(opts *bind.FilterOpts, requirementId [][32]byte) (*IdentityvalidatorpolicyCredentialRequirementRemovedIterator, error) {

	var requirementIdRule []interface{}
	for _, requirementIdItem := range requirementId {
		requirementIdRule = append(requirementIdRule, requirementIdItem)
	}

	logs, sub, err := _Identityvalidatorpolicy.contract.FilterLogs(opts, "CredentialRequirementRemoved", requirementIdRule)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorpolicyCredentialRequirementRemovedIterator{contract: _Identityvalidatorpolicy.contract, event: "CredentialRequirementRemoved", logs: logs, sub: sub}, nil
}

// WatchCredentialRequirementRemoved is a free log subscription operation binding the contract event 0xd31ba7aeb0964621ef60b025c72e30a8816a36274b78ffd549764f87e453e112.
//
// Solidity: event CredentialRequirementRemoved(bytes32 indexed requirementId, bytes32[] credentialTypeIds, uint256 minValidations)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) WatchCredentialRequirementRemoved(opts *bind.WatchOpts, sink chan<- *IdentityvalidatorpolicyCredentialRequirementRemoved, requirementId [][32]byte) (event.Subscription, error) {

	var requirementIdRule []interface{}
	for _, requirementIdItem := range requirementId {
		requirementIdRule = append(requirementIdRule, requirementIdItem)
	}

	logs, sub, err := _Identityvalidatorpolicy.contract.WatchLogs(opts, "CredentialRequirementRemoved", requirementIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityvalidatorpolicyCredentialRequirementRemoved)
				if err := _Identityvalidatorpolicy.contract.UnpackLog(event, "CredentialRequirementRemoved", log); err != nil {
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
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) ParseCredentialRequirementRemoved(log types.Log) (*IdentityvalidatorpolicyCredentialRequirementRemoved, error) {
	event := new(IdentityvalidatorpolicyCredentialRequirementRemoved)
	if err := _Identityvalidatorpolicy.contract.UnpackLog(event, "CredentialRequirementRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IdentityvalidatorpolicyCredentialSourceAddedIterator is returned from FilterCredentialSourceAdded and is used to iterate over the raw logs and unpacked data for CredentialSourceAdded events raised by the Identityvalidatorpolicy contract.
type IdentityvalidatorpolicyCredentialSourceAddedIterator struct {
	Event *IdentityvalidatorpolicyCredentialSourceAdded // Event containing the contract specifics and raw log

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
func (it *IdentityvalidatorpolicyCredentialSourceAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityvalidatorpolicyCredentialSourceAdded)
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
		it.Event = new(IdentityvalidatorpolicyCredentialSourceAdded)
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
func (it *IdentityvalidatorpolicyCredentialSourceAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityvalidatorpolicyCredentialSourceAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityvalidatorpolicyCredentialSourceAdded represents a CredentialSourceAdded event raised by the Identityvalidatorpolicy contract.
type IdentityvalidatorpolicyCredentialSourceAdded struct {
	CredentialTypeId   [32]byte
	IdentityRegistry   common.Address
	CredentialRegistry common.Address
	DataValidator      common.Address
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterCredentialSourceAdded is a free log retrieval operation binding the contract event 0x62bbab76853dfe54818ecdeaca8c73b87b7d3266439a265109815d909ed277c9.
//
// Solidity: event CredentialSourceAdded(bytes32 indexed credentialTypeId, address indexed identityRegistry, address indexed credentialRegistry, address dataValidator)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) FilterCredentialSourceAdded(opts *bind.FilterOpts, credentialTypeId [][32]byte, identityRegistry []common.Address, credentialRegistry []common.Address) (*IdentityvalidatorpolicyCredentialSourceAddedIterator, error) {

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

	logs, sub, err := _Identityvalidatorpolicy.contract.FilterLogs(opts, "CredentialSourceAdded", credentialTypeIdRule, identityRegistryRule, credentialRegistryRule)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorpolicyCredentialSourceAddedIterator{contract: _Identityvalidatorpolicy.contract, event: "CredentialSourceAdded", logs: logs, sub: sub}, nil
}

// WatchCredentialSourceAdded is a free log subscription operation binding the contract event 0x62bbab76853dfe54818ecdeaca8c73b87b7d3266439a265109815d909ed277c9.
//
// Solidity: event CredentialSourceAdded(bytes32 indexed credentialTypeId, address indexed identityRegistry, address indexed credentialRegistry, address dataValidator)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) WatchCredentialSourceAdded(opts *bind.WatchOpts, sink chan<- *IdentityvalidatorpolicyCredentialSourceAdded, credentialTypeId [][32]byte, identityRegistry []common.Address, credentialRegistry []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _Identityvalidatorpolicy.contract.WatchLogs(opts, "CredentialSourceAdded", credentialTypeIdRule, identityRegistryRule, credentialRegistryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityvalidatorpolicyCredentialSourceAdded)
				if err := _Identityvalidatorpolicy.contract.UnpackLog(event, "CredentialSourceAdded", log); err != nil {
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
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) ParseCredentialSourceAdded(log types.Log) (*IdentityvalidatorpolicyCredentialSourceAdded, error) {
	event := new(IdentityvalidatorpolicyCredentialSourceAdded)
	if err := _Identityvalidatorpolicy.contract.UnpackLog(event, "CredentialSourceAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IdentityvalidatorpolicyCredentialSourceRemovedIterator is returned from FilterCredentialSourceRemoved and is used to iterate over the raw logs and unpacked data for CredentialSourceRemoved events raised by the Identityvalidatorpolicy contract.
type IdentityvalidatorpolicyCredentialSourceRemovedIterator struct {
	Event *IdentityvalidatorpolicyCredentialSourceRemoved // Event containing the contract specifics and raw log

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
func (it *IdentityvalidatorpolicyCredentialSourceRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityvalidatorpolicyCredentialSourceRemoved)
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
		it.Event = new(IdentityvalidatorpolicyCredentialSourceRemoved)
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
func (it *IdentityvalidatorpolicyCredentialSourceRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityvalidatorpolicyCredentialSourceRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityvalidatorpolicyCredentialSourceRemoved represents a CredentialSourceRemoved event raised by the Identityvalidatorpolicy contract.
type IdentityvalidatorpolicyCredentialSourceRemoved struct {
	CredentialTypeId   [32]byte
	IdentityRegistry   common.Address
	CredentialRegistry common.Address
	DataValidator      common.Address
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterCredentialSourceRemoved is a free log retrieval operation binding the contract event 0x9c6217b18044e22e3ba28f6713b159e1fa7b0dc0340eb48512432fdfc4650705.
//
// Solidity: event CredentialSourceRemoved(bytes32 indexed credentialTypeId, address indexed identityRegistry, address indexed credentialRegistry, address dataValidator)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) FilterCredentialSourceRemoved(opts *bind.FilterOpts, credentialTypeId [][32]byte, identityRegistry []common.Address, credentialRegistry []common.Address) (*IdentityvalidatorpolicyCredentialSourceRemovedIterator, error) {

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

	logs, sub, err := _Identityvalidatorpolicy.contract.FilterLogs(opts, "CredentialSourceRemoved", credentialTypeIdRule, identityRegistryRule, credentialRegistryRule)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorpolicyCredentialSourceRemovedIterator{contract: _Identityvalidatorpolicy.contract, event: "CredentialSourceRemoved", logs: logs, sub: sub}, nil
}

// WatchCredentialSourceRemoved is a free log subscription operation binding the contract event 0x9c6217b18044e22e3ba28f6713b159e1fa7b0dc0340eb48512432fdfc4650705.
//
// Solidity: event CredentialSourceRemoved(bytes32 indexed credentialTypeId, address indexed identityRegistry, address indexed credentialRegistry, address dataValidator)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) WatchCredentialSourceRemoved(opts *bind.WatchOpts, sink chan<- *IdentityvalidatorpolicyCredentialSourceRemoved, credentialTypeId [][32]byte, identityRegistry []common.Address, credentialRegistry []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _Identityvalidatorpolicy.contract.WatchLogs(opts, "CredentialSourceRemoved", credentialTypeIdRule, identityRegistryRule, credentialRegistryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityvalidatorpolicyCredentialSourceRemoved)
				if err := _Identityvalidatorpolicy.contract.UnpackLog(event, "CredentialSourceRemoved", log); err != nil {
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
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) ParseCredentialSourceRemoved(log types.Log) (*IdentityvalidatorpolicyCredentialSourceRemoved, error) {
	event := new(IdentityvalidatorpolicyCredentialSourceRemoved)
	if err := _Identityvalidatorpolicy.contract.UnpackLog(event, "CredentialSourceRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IdentityvalidatorpolicyInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Identityvalidatorpolicy contract.
type IdentityvalidatorpolicyInitializedIterator struct {
	Event *IdentityvalidatorpolicyInitialized // Event containing the contract specifics and raw log

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
func (it *IdentityvalidatorpolicyInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityvalidatorpolicyInitialized)
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
		it.Event = new(IdentityvalidatorpolicyInitialized)
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
func (it *IdentityvalidatorpolicyInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityvalidatorpolicyInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityvalidatorpolicyInitialized represents a Initialized event raised by the Identityvalidatorpolicy contract.
type IdentityvalidatorpolicyInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) FilterInitialized(opts *bind.FilterOpts) (*IdentityvalidatorpolicyInitializedIterator, error) {

	logs, sub, err := _Identityvalidatorpolicy.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorpolicyInitializedIterator{contract: _Identityvalidatorpolicy.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *IdentityvalidatorpolicyInitialized) (event.Subscription, error) {

	logs, sub, err := _Identityvalidatorpolicy.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityvalidatorpolicyInitialized)
				if err := _Identityvalidatorpolicy.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) ParseInitialized(log types.Log) (*IdentityvalidatorpolicyInitialized, error) {
	event := new(IdentityvalidatorpolicyInitialized)
	if err := _Identityvalidatorpolicy.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IdentityvalidatorpolicyOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Identityvalidatorpolicy contract.
type IdentityvalidatorpolicyOwnershipTransferredIterator struct {
	Event *IdentityvalidatorpolicyOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *IdentityvalidatorpolicyOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityvalidatorpolicyOwnershipTransferred)
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
		it.Event = new(IdentityvalidatorpolicyOwnershipTransferred)
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
func (it *IdentityvalidatorpolicyOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityvalidatorpolicyOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityvalidatorpolicyOwnershipTransferred represents a OwnershipTransferred event raised by the Identityvalidatorpolicy contract.
type IdentityvalidatorpolicyOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*IdentityvalidatorpolicyOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Identityvalidatorpolicy.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &IdentityvalidatorpolicyOwnershipTransferredIterator{contract: _Identityvalidatorpolicy.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *IdentityvalidatorpolicyOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Identityvalidatorpolicy.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityvalidatorpolicyOwnershipTransferred)
				if err := _Identityvalidatorpolicy.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Identityvalidatorpolicy *IdentityvalidatorpolicyFilterer) ParseOwnershipTransferred(log types.Log) (*IdentityvalidatorpolicyOwnershipTransferred, error) {
	event := new(IdentityvalidatorpolicyOwnershipTransferred)
	if err := _Identityvalidatorpolicy.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
