// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package holdmanager

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

// IHoldManagerHold is an auto generated low-level Go binding around an user-defined struct.
type IHoldManagerHold struct {
	Token     common.Address
	Sender    common.Address
	Recipient common.Address
	Executor  common.Address
	ExpiresAt *big.Int
	Value     *big.Int
}

// HoldmanagerMetaData contains all meta data concerning the Holdmanager contract.
var HoldmanagerMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"cancelHold\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"holdId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"createHold\",\"inputs\":[{\"name\":\"holdId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"hold\",\"type\":\"tuple\",\"internalType\":\"structIHoldManager.Hold\",\"components\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"recipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"executor\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"expiresAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"executeHold\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"holdId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"extendHold\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"holdId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"expiresAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getHold\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"holdId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIHoldManager.Hold\",\"components\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"recipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"executor\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"expiresAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getHoldStatus\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"holdId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumIHoldManager.HoldStatus\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"heldBalanceOf\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isExpired\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"holdId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"releaseHold\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"holdId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"HoldCanceled\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"holdId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"HoldCreated\",\"inputs\":[{\"name\":\"holdId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"hold\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structIHoldManager.Hold\",\"components\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"recipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"executor\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"expiresAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"HoldExecuted\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"holdId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"HoldExtended\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"holdId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"expiresAt\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"HoldReleased\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"holdId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false}]",
}

// Holdmanager is an auto generated Go binding around an Ethereum contract.
type Holdmanager struct {
	HoldmanagerCaller     // Read-only binding to the contract
	HoldmanagerTransactor // Write-only binding to the contract
	HoldmanagerFilterer   // Log filterer for contract events
}

// HoldmanagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type HoldmanagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HoldmanagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type HoldmanagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HoldmanagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type HoldmanagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HoldmanagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type HoldmanagerSession struct {
	Contract     *Holdmanager      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HoldmanagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type HoldmanagerCallerSession struct {
	Contract *HoldmanagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// HoldmanagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type HoldmanagerTransactorSession struct {
	Contract     *HoldmanagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// HoldmanagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type HoldmanagerRaw struct {
	Contract *Holdmanager // Generic contract binding to access the raw methods on
}

// HoldmanagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type HoldmanagerCallerRaw struct {
	Contract *HoldmanagerCaller // Generic read-only contract binding to access the raw methods on
}

// HoldmanagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type HoldmanagerTransactorRaw struct {
	Contract *HoldmanagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewHoldmanager creates a new instance of Holdmanager, bound to a specific deployed contract.
func NewHoldmanager(address common.Address, backend bind.ContractBackend) (*Holdmanager, error) {
	contract, err := bindHoldmanager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Holdmanager{HoldmanagerCaller: HoldmanagerCaller{contract: contract}, HoldmanagerTransactor: HoldmanagerTransactor{contract: contract}, HoldmanagerFilterer: HoldmanagerFilterer{contract: contract}}, nil
}

// NewHoldmanagerCaller creates a new read-only instance of Holdmanager, bound to a specific deployed contract.
func NewHoldmanagerCaller(address common.Address, caller bind.ContractCaller) (*HoldmanagerCaller, error) {
	contract, err := bindHoldmanager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &HoldmanagerCaller{contract: contract}, nil
}

// NewHoldmanagerTransactor creates a new write-only instance of Holdmanager, bound to a specific deployed contract.
func NewHoldmanagerTransactor(address common.Address, transactor bind.ContractTransactor) (*HoldmanagerTransactor, error) {
	contract, err := bindHoldmanager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &HoldmanagerTransactor{contract: contract}, nil
}

// NewHoldmanagerFilterer creates a new log filterer instance of Holdmanager, bound to a specific deployed contract.
func NewHoldmanagerFilterer(address common.Address, filterer bind.ContractFilterer) (*HoldmanagerFilterer, error) {
	contract, err := bindHoldmanager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &HoldmanagerFilterer{contract: contract}, nil
}

// bindHoldmanager binds a generic wrapper to an already deployed contract.
func bindHoldmanager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := HoldmanagerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Holdmanager *HoldmanagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Holdmanager.Contract.HoldmanagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Holdmanager *HoldmanagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Holdmanager.Contract.HoldmanagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Holdmanager *HoldmanagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Holdmanager.Contract.HoldmanagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Holdmanager *HoldmanagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Holdmanager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Holdmanager *HoldmanagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Holdmanager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Holdmanager *HoldmanagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Holdmanager.Contract.contract.Transact(opts, method, params...)
}

// GetHold is a free data retrieval call binding the contract method 0x003a30a6.
//
// Solidity: function getHold(address sender, bytes32 holdId) view returns((address,address,address,address,uint256,uint256))
func (_Holdmanager *HoldmanagerCaller) GetHold(opts *bind.CallOpts, sender common.Address, holdId [32]byte) (IHoldManagerHold, error) {
	var out []interface{}
	err := _Holdmanager.contract.Call(opts, &out, "getHold", sender, holdId)

	if err != nil {
		return *new(IHoldManagerHold), err
	}

	out0 := *abi.ConvertType(out[0], new(IHoldManagerHold)).(*IHoldManagerHold)

	return out0, err

}

// GetHold is a free data retrieval call binding the contract method 0x003a30a6.
//
// Solidity: function getHold(address sender, bytes32 holdId) view returns((address,address,address,address,uint256,uint256))
func (_Holdmanager *HoldmanagerSession) GetHold(sender common.Address, holdId [32]byte) (IHoldManagerHold, error) {
	return _Holdmanager.Contract.GetHold(&_Holdmanager.CallOpts, sender, holdId)
}

// GetHold is a free data retrieval call binding the contract method 0x003a30a6.
//
// Solidity: function getHold(address sender, bytes32 holdId) view returns((address,address,address,address,uint256,uint256))
func (_Holdmanager *HoldmanagerCallerSession) GetHold(sender common.Address, holdId [32]byte) (IHoldManagerHold, error) {
	return _Holdmanager.Contract.GetHold(&_Holdmanager.CallOpts, sender, holdId)
}

// GetHoldStatus is a free data retrieval call binding the contract method 0x5cd2aadc.
//
// Solidity: function getHoldStatus(address sender, bytes32 holdId) view returns(uint8)
func (_Holdmanager *HoldmanagerCaller) GetHoldStatus(opts *bind.CallOpts, sender common.Address, holdId [32]byte) (uint8, error) {
	var out []interface{}
	err := _Holdmanager.contract.Call(opts, &out, "getHoldStatus", sender, holdId)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetHoldStatus is a free data retrieval call binding the contract method 0x5cd2aadc.
//
// Solidity: function getHoldStatus(address sender, bytes32 holdId) view returns(uint8)
func (_Holdmanager *HoldmanagerSession) GetHoldStatus(sender common.Address, holdId [32]byte) (uint8, error) {
	return _Holdmanager.Contract.GetHoldStatus(&_Holdmanager.CallOpts, sender, holdId)
}

// GetHoldStatus is a free data retrieval call binding the contract method 0x5cd2aadc.
//
// Solidity: function getHoldStatus(address sender, bytes32 holdId) view returns(uint8)
func (_Holdmanager *HoldmanagerCallerSession) GetHoldStatus(sender common.Address, holdId [32]byte) (uint8, error) {
	return _Holdmanager.Contract.GetHoldStatus(&_Holdmanager.CallOpts, sender, holdId)
}

// HeldBalanceOf is a free data retrieval call binding the contract method 0xbe9c1add.
//
// Solidity: function heldBalanceOf(address sender) view returns(uint256)
func (_Holdmanager *HoldmanagerCaller) HeldBalanceOf(opts *bind.CallOpts, sender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Holdmanager.contract.Call(opts, &out, "heldBalanceOf", sender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// HeldBalanceOf is a free data retrieval call binding the contract method 0xbe9c1add.
//
// Solidity: function heldBalanceOf(address sender) view returns(uint256)
func (_Holdmanager *HoldmanagerSession) HeldBalanceOf(sender common.Address) (*big.Int, error) {
	return _Holdmanager.Contract.HeldBalanceOf(&_Holdmanager.CallOpts, sender)
}

// HeldBalanceOf is a free data retrieval call binding the contract method 0xbe9c1add.
//
// Solidity: function heldBalanceOf(address sender) view returns(uint256)
func (_Holdmanager *HoldmanagerCallerSession) HeldBalanceOf(sender common.Address) (*big.Int, error) {
	return _Holdmanager.Contract.HeldBalanceOf(&_Holdmanager.CallOpts, sender)
}

// IsExpired is a free data retrieval call binding the contract method 0x00f304db.
//
// Solidity: function isExpired(address sender, bytes32 holdId) view returns(bool)
func (_Holdmanager *HoldmanagerCaller) IsExpired(opts *bind.CallOpts, sender common.Address, holdId [32]byte) (bool, error) {
	var out []interface{}
	err := _Holdmanager.contract.Call(opts, &out, "isExpired", sender, holdId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsExpired is a free data retrieval call binding the contract method 0x00f304db.
//
// Solidity: function isExpired(address sender, bytes32 holdId) view returns(bool)
func (_Holdmanager *HoldmanagerSession) IsExpired(sender common.Address, holdId [32]byte) (bool, error) {
	return _Holdmanager.Contract.IsExpired(&_Holdmanager.CallOpts, sender, holdId)
}

// IsExpired is a free data retrieval call binding the contract method 0x00f304db.
//
// Solidity: function isExpired(address sender, bytes32 holdId) view returns(bool)
func (_Holdmanager *HoldmanagerCallerSession) IsExpired(sender common.Address, holdId [32]byte) (bool, error) {
	return _Holdmanager.Contract.IsExpired(&_Holdmanager.CallOpts, sender, holdId)
}

// CancelHold is a paid mutator transaction binding the contract method 0xeee2068b.
//
// Solidity: function cancelHold(address sender, bytes32 holdId) returns()
func (_Holdmanager *HoldmanagerTransactor) CancelHold(opts *bind.TransactOpts, sender common.Address, holdId [32]byte) (*types.Transaction, error) {
	return _Holdmanager.contract.Transact(opts, "cancelHold", sender, holdId)
}

// CancelHold is a paid mutator transaction binding the contract method 0xeee2068b.
//
// Solidity: function cancelHold(address sender, bytes32 holdId) returns()
func (_Holdmanager *HoldmanagerSession) CancelHold(sender common.Address, holdId [32]byte) (*types.Transaction, error) {
	return _Holdmanager.Contract.CancelHold(&_Holdmanager.TransactOpts, sender, holdId)
}

// CancelHold is a paid mutator transaction binding the contract method 0xeee2068b.
//
// Solidity: function cancelHold(address sender, bytes32 holdId) returns()
func (_Holdmanager *HoldmanagerTransactorSession) CancelHold(sender common.Address, holdId [32]byte) (*types.Transaction, error) {
	return _Holdmanager.Contract.CancelHold(&_Holdmanager.TransactOpts, sender, holdId)
}

// CreateHold is a paid mutator transaction binding the contract method 0xae01e0bd.
//
// Solidity: function createHold(bytes32 holdId, (address,address,address,address,uint256,uint256) hold) returns()
func (_Holdmanager *HoldmanagerTransactor) CreateHold(opts *bind.TransactOpts, holdId [32]byte, hold IHoldManagerHold) (*types.Transaction, error) {
	return _Holdmanager.contract.Transact(opts, "createHold", holdId, hold)
}

// CreateHold is a paid mutator transaction binding the contract method 0xae01e0bd.
//
// Solidity: function createHold(bytes32 holdId, (address,address,address,address,uint256,uint256) hold) returns()
func (_Holdmanager *HoldmanagerSession) CreateHold(holdId [32]byte, hold IHoldManagerHold) (*types.Transaction, error) {
	return _Holdmanager.Contract.CreateHold(&_Holdmanager.TransactOpts, holdId, hold)
}

// CreateHold is a paid mutator transaction binding the contract method 0xae01e0bd.
//
// Solidity: function createHold(bytes32 holdId, (address,address,address,address,uint256,uint256) hold) returns()
func (_Holdmanager *HoldmanagerTransactorSession) CreateHold(holdId [32]byte, hold IHoldManagerHold) (*types.Transaction, error) {
	return _Holdmanager.Contract.CreateHold(&_Holdmanager.TransactOpts, holdId, hold)
}

// ExecuteHold is a paid mutator transaction binding the contract method 0x0e432dec.
//
// Solidity: function executeHold(address sender, bytes32 holdId) returns()
func (_Holdmanager *HoldmanagerTransactor) ExecuteHold(opts *bind.TransactOpts, sender common.Address, holdId [32]byte) (*types.Transaction, error) {
	return _Holdmanager.contract.Transact(opts, "executeHold", sender, holdId)
}

// ExecuteHold is a paid mutator transaction binding the contract method 0x0e432dec.
//
// Solidity: function executeHold(address sender, bytes32 holdId) returns()
func (_Holdmanager *HoldmanagerSession) ExecuteHold(sender common.Address, holdId [32]byte) (*types.Transaction, error) {
	return _Holdmanager.Contract.ExecuteHold(&_Holdmanager.TransactOpts, sender, holdId)
}

// ExecuteHold is a paid mutator transaction binding the contract method 0x0e432dec.
//
// Solidity: function executeHold(address sender, bytes32 holdId) returns()
func (_Holdmanager *HoldmanagerTransactorSession) ExecuteHold(sender common.Address, holdId [32]byte) (*types.Transaction, error) {
	return _Holdmanager.Contract.ExecuteHold(&_Holdmanager.TransactOpts, sender, holdId)
}

// ExtendHold is a paid mutator transaction binding the contract method 0x29251264.
//
// Solidity: function extendHold(address sender, bytes32 holdId, uint256 expiresAt) returns()
func (_Holdmanager *HoldmanagerTransactor) ExtendHold(opts *bind.TransactOpts, sender common.Address, holdId [32]byte, expiresAt *big.Int) (*types.Transaction, error) {
	return _Holdmanager.contract.Transact(opts, "extendHold", sender, holdId, expiresAt)
}

// ExtendHold is a paid mutator transaction binding the contract method 0x29251264.
//
// Solidity: function extendHold(address sender, bytes32 holdId, uint256 expiresAt) returns()
func (_Holdmanager *HoldmanagerSession) ExtendHold(sender common.Address, holdId [32]byte, expiresAt *big.Int) (*types.Transaction, error) {
	return _Holdmanager.Contract.ExtendHold(&_Holdmanager.TransactOpts, sender, holdId, expiresAt)
}

// ExtendHold is a paid mutator transaction binding the contract method 0x29251264.
//
// Solidity: function extendHold(address sender, bytes32 holdId, uint256 expiresAt) returns()
func (_Holdmanager *HoldmanagerTransactorSession) ExtendHold(sender common.Address, holdId [32]byte, expiresAt *big.Int) (*types.Transaction, error) {
	return _Holdmanager.Contract.ExtendHold(&_Holdmanager.TransactOpts, sender, holdId, expiresAt)
}

// ReleaseHold is a paid mutator transaction binding the contract method 0xc1c60fe2.
//
// Solidity: function releaseHold(address sender, bytes32 holdId) returns()
func (_Holdmanager *HoldmanagerTransactor) ReleaseHold(opts *bind.TransactOpts, sender common.Address, holdId [32]byte) (*types.Transaction, error) {
	return _Holdmanager.contract.Transact(opts, "releaseHold", sender, holdId)
}

// ReleaseHold is a paid mutator transaction binding the contract method 0xc1c60fe2.
//
// Solidity: function releaseHold(address sender, bytes32 holdId) returns()
func (_Holdmanager *HoldmanagerSession) ReleaseHold(sender common.Address, holdId [32]byte) (*types.Transaction, error) {
	return _Holdmanager.Contract.ReleaseHold(&_Holdmanager.TransactOpts, sender, holdId)
}

// ReleaseHold is a paid mutator transaction binding the contract method 0xc1c60fe2.
//
// Solidity: function releaseHold(address sender, bytes32 holdId) returns()
func (_Holdmanager *HoldmanagerTransactorSession) ReleaseHold(sender common.Address, holdId [32]byte) (*types.Transaction, error) {
	return _Holdmanager.Contract.ReleaseHold(&_Holdmanager.TransactOpts, sender, holdId)
}

// HoldmanagerHoldCanceledIterator is returned from FilterHoldCanceled and is used to iterate over the raw logs and unpacked data for HoldCanceled events raised by the Holdmanager contract.
type HoldmanagerHoldCanceledIterator struct {
	Event *HoldmanagerHoldCanceled // Event containing the contract specifics and raw log

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
func (it *HoldmanagerHoldCanceledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HoldmanagerHoldCanceled)
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
		it.Event = new(HoldmanagerHoldCanceled)
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
func (it *HoldmanagerHoldCanceledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HoldmanagerHoldCanceledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HoldmanagerHoldCanceled represents a HoldCanceled event raised by the Holdmanager contract.
type HoldmanagerHoldCanceled struct {
	Sender common.Address
	HoldId [32]byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterHoldCanceled is a free log retrieval operation binding the contract event 0x2fd5592c5d0335e054db1164b19cf94f92ed8f707baa92cc90f565ba7e53b196.
//
// Solidity: event HoldCanceled(address indexed sender, bytes32 indexed holdId)
func (_Holdmanager *HoldmanagerFilterer) FilterHoldCanceled(opts *bind.FilterOpts, sender []common.Address, holdId [][32]byte) (*HoldmanagerHoldCanceledIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var holdIdRule []interface{}
	for _, holdIdItem := range holdId {
		holdIdRule = append(holdIdRule, holdIdItem)
	}

	logs, sub, err := _Holdmanager.contract.FilterLogs(opts, "HoldCanceled", senderRule, holdIdRule)
	if err != nil {
		return nil, err
	}
	return &HoldmanagerHoldCanceledIterator{contract: _Holdmanager.contract, event: "HoldCanceled", logs: logs, sub: sub}, nil
}

// WatchHoldCanceled is a free log subscription operation binding the contract event 0x2fd5592c5d0335e054db1164b19cf94f92ed8f707baa92cc90f565ba7e53b196.
//
// Solidity: event HoldCanceled(address indexed sender, bytes32 indexed holdId)
func (_Holdmanager *HoldmanagerFilterer) WatchHoldCanceled(opts *bind.WatchOpts, sink chan<- *HoldmanagerHoldCanceled, sender []common.Address, holdId [][32]byte) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var holdIdRule []interface{}
	for _, holdIdItem := range holdId {
		holdIdRule = append(holdIdRule, holdIdItem)
	}

	logs, sub, err := _Holdmanager.contract.WatchLogs(opts, "HoldCanceled", senderRule, holdIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HoldmanagerHoldCanceled)
				if err := _Holdmanager.contract.UnpackLog(event, "HoldCanceled", log); err != nil {
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

// ParseHoldCanceled is a log parse operation binding the contract event 0x2fd5592c5d0335e054db1164b19cf94f92ed8f707baa92cc90f565ba7e53b196.
//
// Solidity: event HoldCanceled(address indexed sender, bytes32 indexed holdId)
func (_Holdmanager *HoldmanagerFilterer) ParseHoldCanceled(log types.Log) (*HoldmanagerHoldCanceled, error) {
	event := new(HoldmanagerHoldCanceled)
	if err := _Holdmanager.contract.UnpackLog(event, "HoldCanceled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HoldmanagerHoldCreatedIterator is returned from FilterHoldCreated and is used to iterate over the raw logs and unpacked data for HoldCreated events raised by the Holdmanager contract.
type HoldmanagerHoldCreatedIterator struct {
	Event *HoldmanagerHoldCreated // Event containing the contract specifics and raw log

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
func (it *HoldmanagerHoldCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HoldmanagerHoldCreated)
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
		it.Event = new(HoldmanagerHoldCreated)
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
func (it *HoldmanagerHoldCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HoldmanagerHoldCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HoldmanagerHoldCreated represents a HoldCreated event raised by the Holdmanager contract.
type HoldmanagerHoldCreated struct {
	HoldId [32]byte
	Hold   IHoldManagerHold
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterHoldCreated is a free log retrieval operation binding the contract event 0xea975310e945455e3e9ac7b1b0678439bbd01bdbeba38c4a8542da22dc85b95a.
//
// Solidity: event HoldCreated(bytes32 indexed holdId, (address,address,address,address,uint256,uint256) hold)
func (_Holdmanager *HoldmanagerFilterer) FilterHoldCreated(opts *bind.FilterOpts, holdId [][32]byte) (*HoldmanagerHoldCreatedIterator, error) {

	var holdIdRule []interface{}
	for _, holdIdItem := range holdId {
		holdIdRule = append(holdIdRule, holdIdItem)
	}

	logs, sub, err := _Holdmanager.contract.FilterLogs(opts, "HoldCreated", holdIdRule)
	if err != nil {
		return nil, err
	}
	return &HoldmanagerHoldCreatedIterator{contract: _Holdmanager.contract, event: "HoldCreated", logs: logs, sub: sub}, nil
}

// WatchHoldCreated is a free log subscription operation binding the contract event 0xea975310e945455e3e9ac7b1b0678439bbd01bdbeba38c4a8542da22dc85b95a.
//
// Solidity: event HoldCreated(bytes32 indexed holdId, (address,address,address,address,uint256,uint256) hold)
func (_Holdmanager *HoldmanagerFilterer) WatchHoldCreated(opts *bind.WatchOpts, sink chan<- *HoldmanagerHoldCreated, holdId [][32]byte) (event.Subscription, error) {

	var holdIdRule []interface{}
	for _, holdIdItem := range holdId {
		holdIdRule = append(holdIdRule, holdIdItem)
	}

	logs, sub, err := _Holdmanager.contract.WatchLogs(opts, "HoldCreated", holdIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HoldmanagerHoldCreated)
				if err := _Holdmanager.contract.UnpackLog(event, "HoldCreated", log); err != nil {
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

// ParseHoldCreated is a log parse operation binding the contract event 0xea975310e945455e3e9ac7b1b0678439bbd01bdbeba38c4a8542da22dc85b95a.
//
// Solidity: event HoldCreated(bytes32 indexed holdId, (address,address,address,address,uint256,uint256) hold)
func (_Holdmanager *HoldmanagerFilterer) ParseHoldCreated(log types.Log) (*HoldmanagerHoldCreated, error) {
	event := new(HoldmanagerHoldCreated)
	if err := _Holdmanager.contract.UnpackLog(event, "HoldCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HoldmanagerHoldExecutedIterator is returned from FilterHoldExecuted and is used to iterate over the raw logs and unpacked data for HoldExecuted events raised by the Holdmanager contract.
type HoldmanagerHoldExecutedIterator struct {
	Event *HoldmanagerHoldExecuted // Event containing the contract specifics and raw log

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
func (it *HoldmanagerHoldExecutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HoldmanagerHoldExecuted)
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
		it.Event = new(HoldmanagerHoldExecuted)
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
func (it *HoldmanagerHoldExecutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HoldmanagerHoldExecutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HoldmanagerHoldExecuted represents a HoldExecuted event raised by the Holdmanager contract.
type HoldmanagerHoldExecuted struct {
	Sender common.Address
	HoldId [32]byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterHoldExecuted is a free log retrieval operation binding the contract event 0xed3cc5a786d717fbf71ddf9b128c44ac87cd0fb35c999373b2f938d14c12066f.
//
// Solidity: event HoldExecuted(address indexed sender, bytes32 indexed holdId)
func (_Holdmanager *HoldmanagerFilterer) FilterHoldExecuted(opts *bind.FilterOpts, sender []common.Address, holdId [][32]byte) (*HoldmanagerHoldExecutedIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var holdIdRule []interface{}
	for _, holdIdItem := range holdId {
		holdIdRule = append(holdIdRule, holdIdItem)
	}

	logs, sub, err := _Holdmanager.contract.FilterLogs(opts, "HoldExecuted", senderRule, holdIdRule)
	if err != nil {
		return nil, err
	}
	return &HoldmanagerHoldExecutedIterator{contract: _Holdmanager.contract, event: "HoldExecuted", logs: logs, sub: sub}, nil
}

// WatchHoldExecuted is a free log subscription operation binding the contract event 0xed3cc5a786d717fbf71ddf9b128c44ac87cd0fb35c999373b2f938d14c12066f.
//
// Solidity: event HoldExecuted(address indexed sender, bytes32 indexed holdId)
func (_Holdmanager *HoldmanagerFilterer) WatchHoldExecuted(opts *bind.WatchOpts, sink chan<- *HoldmanagerHoldExecuted, sender []common.Address, holdId [][32]byte) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var holdIdRule []interface{}
	for _, holdIdItem := range holdId {
		holdIdRule = append(holdIdRule, holdIdItem)
	}

	logs, sub, err := _Holdmanager.contract.WatchLogs(opts, "HoldExecuted", senderRule, holdIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HoldmanagerHoldExecuted)
				if err := _Holdmanager.contract.UnpackLog(event, "HoldExecuted", log); err != nil {
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

// ParseHoldExecuted is a log parse operation binding the contract event 0xed3cc5a786d717fbf71ddf9b128c44ac87cd0fb35c999373b2f938d14c12066f.
//
// Solidity: event HoldExecuted(address indexed sender, bytes32 indexed holdId)
func (_Holdmanager *HoldmanagerFilterer) ParseHoldExecuted(log types.Log) (*HoldmanagerHoldExecuted, error) {
	event := new(HoldmanagerHoldExecuted)
	if err := _Holdmanager.contract.UnpackLog(event, "HoldExecuted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HoldmanagerHoldExtendedIterator is returned from FilterHoldExtended and is used to iterate over the raw logs and unpacked data for HoldExtended events raised by the Holdmanager contract.
type HoldmanagerHoldExtendedIterator struct {
	Event *HoldmanagerHoldExtended // Event containing the contract specifics and raw log

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
func (it *HoldmanagerHoldExtendedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HoldmanagerHoldExtended)
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
		it.Event = new(HoldmanagerHoldExtended)
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
func (it *HoldmanagerHoldExtendedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HoldmanagerHoldExtendedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HoldmanagerHoldExtended represents a HoldExtended event raised by the Holdmanager contract.
type HoldmanagerHoldExtended struct {
	Sender    common.Address
	HoldId    [32]byte
	ExpiresAt *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterHoldExtended is a free log retrieval operation binding the contract event 0x33daaf142cd1f192d2598aa040236d4c3c21f9fc800f7219bca5d8c0a96fbeee.
//
// Solidity: event HoldExtended(address indexed sender, bytes32 indexed holdId, uint256 expiresAt)
func (_Holdmanager *HoldmanagerFilterer) FilterHoldExtended(opts *bind.FilterOpts, sender []common.Address, holdId [][32]byte) (*HoldmanagerHoldExtendedIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var holdIdRule []interface{}
	for _, holdIdItem := range holdId {
		holdIdRule = append(holdIdRule, holdIdItem)
	}

	logs, sub, err := _Holdmanager.contract.FilterLogs(opts, "HoldExtended", senderRule, holdIdRule)
	if err != nil {
		return nil, err
	}
	return &HoldmanagerHoldExtendedIterator{contract: _Holdmanager.contract, event: "HoldExtended", logs: logs, sub: sub}, nil
}

// WatchHoldExtended is a free log subscription operation binding the contract event 0x33daaf142cd1f192d2598aa040236d4c3c21f9fc800f7219bca5d8c0a96fbeee.
//
// Solidity: event HoldExtended(address indexed sender, bytes32 indexed holdId, uint256 expiresAt)
func (_Holdmanager *HoldmanagerFilterer) WatchHoldExtended(opts *bind.WatchOpts, sink chan<- *HoldmanagerHoldExtended, sender []common.Address, holdId [][32]byte) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var holdIdRule []interface{}
	for _, holdIdItem := range holdId {
		holdIdRule = append(holdIdRule, holdIdItem)
	}

	logs, sub, err := _Holdmanager.contract.WatchLogs(opts, "HoldExtended", senderRule, holdIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HoldmanagerHoldExtended)
				if err := _Holdmanager.contract.UnpackLog(event, "HoldExtended", log); err != nil {
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

// ParseHoldExtended is a log parse operation binding the contract event 0x33daaf142cd1f192d2598aa040236d4c3c21f9fc800f7219bca5d8c0a96fbeee.
//
// Solidity: event HoldExtended(address indexed sender, bytes32 indexed holdId, uint256 expiresAt)
func (_Holdmanager *HoldmanagerFilterer) ParseHoldExtended(log types.Log) (*HoldmanagerHoldExtended, error) {
	event := new(HoldmanagerHoldExtended)
	if err := _Holdmanager.contract.UnpackLog(event, "HoldExtended", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HoldmanagerHoldReleasedIterator is returned from FilterHoldReleased and is used to iterate over the raw logs and unpacked data for HoldReleased events raised by the Holdmanager contract.
type HoldmanagerHoldReleasedIterator struct {
	Event *HoldmanagerHoldReleased // Event containing the contract specifics and raw log

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
func (it *HoldmanagerHoldReleasedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HoldmanagerHoldReleased)
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
		it.Event = new(HoldmanagerHoldReleased)
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
func (it *HoldmanagerHoldReleasedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HoldmanagerHoldReleasedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HoldmanagerHoldReleased represents a HoldReleased event raised by the Holdmanager contract.
type HoldmanagerHoldReleased struct {
	Sender common.Address
	HoldId [32]byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterHoldReleased is a free log retrieval operation binding the contract event 0xab2a9a25f2d6f297f59ab9852e62b3cd1623db40094924f7da14222f59da7d3e.
//
// Solidity: event HoldReleased(address indexed sender, bytes32 indexed holdId)
func (_Holdmanager *HoldmanagerFilterer) FilterHoldReleased(opts *bind.FilterOpts, sender []common.Address, holdId [][32]byte) (*HoldmanagerHoldReleasedIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var holdIdRule []interface{}
	for _, holdIdItem := range holdId {
		holdIdRule = append(holdIdRule, holdIdItem)
	}

	logs, sub, err := _Holdmanager.contract.FilterLogs(opts, "HoldReleased", senderRule, holdIdRule)
	if err != nil {
		return nil, err
	}
	return &HoldmanagerHoldReleasedIterator{contract: _Holdmanager.contract, event: "HoldReleased", logs: logs, sub: sub}, nil
}

// WatchHoldReleased is a free log subscription operation binding the contract event 0xab2a9a25f2d6f297f59ab9852e62b3cd1623db40094924f7da14222f59da7d3e.
//
// Solidity: event HoldReleased(address indexed sender, bytes32 indexed holdId)
func (_Holdmanager *HoldmanagerFilterer) WatchHoldReleased(opts *bind.WatchOpts, sink chan<- *HoldmanagerHoldReleased, sender []common.Address, holdId [][32]byte) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var holdIdRule []interface{}
	for _, holdIdItem := range holdId {
		holdIdRule = append(holdIdRule, holdIdItem)
	}

	logs, sub, err := _Holdmanager.contract.WatchLogs(opts, "HoldReleased", senderRule, holdIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HoldmanagerHoldReleased)
				if err := _Holdmanager.contract.UnpackLog(event, "HoldReleased", log); err != nil {
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

// ParseHoldReleased is a log parse operation binding the contract event 0xab2a9a25f2d6f297f59ab9852e62b3cd1623db40094924f7da14222f59da7d3e.
//
// Solidity: event HoldReleased(address indexed sender, bytes32 indexed holdId)
func (_Holdmanager *HoldmanagerFilterer) ParseHoldReleased(log types.Log) (*HoldmanagerHoldReleased, error) {
	event := new(HoldmanagerHoldReleased)
	if err := _Holdmanager.contract.UnpackLog(event, "HoldReleased", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
