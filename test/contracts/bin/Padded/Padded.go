// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package Padded

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

// PaddedMetaData contains all meta data concerning the Padded contract.
var PaddedMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"Test0\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"emitEvent\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6080604052348015600f57600080fd5b5060b48061001e6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c80637b0cb83914602d575b600080fd5b60336035565b005b6040805180820190915260038082526201020360e81b602083019081527f1ea5ab62888366c058d9c8bd1be1081b32b39c8fbc3f91500bbd8a434be3936991828183a15050505056fea26469706673582212205a5b5d5045ac67ec2b85ed9aa44e4c8ee6dc50064b944bf0aad243077972598564736f6c634300080c0033",
}

// PaddedABI is the input ABI used to generate the binding from.
// Deprecated: Use PaddedMetaData.ABI instead.
var PaddedABI = PaddedMetaData.ABI

// PaddedBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use PaddedMetaData.Bin instead.
var PaddedBin = PaddedMetaData.Bin

// DeployPadded deploys a new Ethereum contract, binding an instance of Padded to it.
func DeployPadded(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Padded, error) {
	parsed, err := PaddedMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(PaddedBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Padded{PaddedCaller: PaddedCaller{contract: contract}, PaddedTransactor: PaddedTransactor{contract: contract}, PaddedFilterer: PaddedFilterer{contract: contract}}, nil
}

// Padded is an auto generated Go binding around an Ethereum contract.
type Padded struct {
	PaddedCaller     // Read-only binding to the contract
	PaddedTransactor // Write-only binding to the contract
	PaddedFilterer   // Log filterer for contract events
}

// PaddedCaller is an auto generated read-only Go binding around an Ethereum contract.
type PaddedCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PaddedTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PaddedTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PaddedFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PaddedFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PaddedSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PaddedSession struct {
	Contract     *Padded           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PaddedCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PaddedCallerSession struct {
	Contract *PaddedCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// PaddedTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PaddedTransactorSession struct {
	Contract     *PaddedTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PaddedRaw is an auto generated low-level Go binding around an Ethereum contract.
type PaddedRaw struct {
	Contract *Padded // Generic contract binding to access the raw methods on
}

// PaddedCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PaddedCallerRaw struct {
	Contract *PaddedCaller // Generic read-only contract binding to access the raw methods on
}

// PaddedTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PaddedTransactorRaw struct {
	Contract *PaddedTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPadded creates a new instance of Padded, bound to a specific deployed contract.
func NewPadded(address common.Address, backend bind.ContractBackend) (*Padded, error) {
	contract, err := bindPadded(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Padded{PaddedCaller: PaddedCaller{contract: contract}, PaddedTransactor: PaddedTransactor{contract: contract}, PaddedFilterer: PaddedFilterer{contract: contract}}, nil
}

// NewPaddedCaller creates a new read-only instance of Padded, bound to a specific deployed contract.
func NewPaddedCaller(address common.Address, caller bind.ContractCaller) (*PaddedCaller, error) {
	contract, err := bindPadded(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PaddedCaller{contract: contract}, nil
}

// NewPaddedTransactor creates a new write-only instance of Padded, bound to a specific deployed contract.
func NewPaddedTransactor(address common.Address, transactor bind.ContractTransactor) (*PaddedTransactor, error) {
	contract, err := bindPadded(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PaddedTransactor{contract: contract}, nil
}

// NewPaddedFilterer creates a new log filterer instance of Padded, bound to a specific deployed contract.
func NewPaddedFilterer(address common.Address, filterer bind.ContractFilterer) (*PaddedFilterer, error) {
	contract, err := bindPadded(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PaddedFilterer{contract: contract}, nil
}

// bindPadded binds a generic wrapper to an already deployed contract.
func bindPadded(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PaddedMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Padded *PaddedRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Padded.Contract.PaddedCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Padded *PaddedRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Padded.Contract.PaddedTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Padded *PaddedRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Padded.Contract.PaddedTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Padded *PaddedCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Padded.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Padded *PaddedTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Padded.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Padded *PaddedTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Padded.Contract.contract.Transact(opts, method, params...)
}

// EmitEvent is a paid mutator transaction binding the contract method 0x7b0cb839.
//
// Solidity: function emitEvent() returns()
func (_Padded *PaddedTransactor) EmitEvent(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Padded.contract.Transact(opts, "emitEvent")
}

// EmitEvent is a paid mutator transaction binding the contract method 0x7b0cb839.
//
// Solidity: function emitEvent() returns()
func (_Padded *PaddedSession) EmitEvent() (*types.Transaction, error) {
	return _Padded.Contract.EmitEvent(&_Padded.TransactOpts)
}

// EmitEvent is a paid mutator transaction binding the contract method 0x7b0cb839.
//
// Solidity: function emitEvent() returns()
func (_Padded *PaddedTransactorSession) EmitEvent() (*types.Transaction, error) {
	return _Padded.Contract.EmitEvent(&_Padded.TransactOpts)
}

// PaddedTest0Iterator is returned from FilterTest0 and is used to iterate over the raw logs and unpacked data for Test0 events raised by the Padded contract.
type PaddedTest0Iterator struct {
	Event *PaddedTest0 // Event containing the contract specifics and raw log

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
func (it *PaddedTest0Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaddedTest0)
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
		it.Event = new(PaddedTest0)
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
func (it *PaddedTest0Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaddedTest0Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaddedTest0 represents a Test0 event raised by the Padded contract.
type PaddedTest0 struct {
	Data []byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterTest0 is a free log retrieval operation binding the contract event 0x1ea5ab62888366c058d9c8bd1be1081b32b39c8fbc3f91500bbd8a434be39369.
//
// Solidity: event Test0(bytes data)
func (_Padded *PaddedFilterer) FilterTest0(opts *bind.FilterOpts) (*PaddedTest0Iterator, error) {

	logs, sub, err := _Padded.contract.FilterLogs(opts, "Test0")
	if err != nil {
		return nil, err
	}
	return &PaddedTest0Iterator{contract: _Padded.contract, event: "Test0", logs: logs, sub: sub}, nil
}

// WatchTest0 is a free log subscription operation binding the contract event 0x1ea5ab62888366c058d9c8bd1be1081b32b39c8fbc3f91500bbd8a434be39369.
//
// Solidity: event Test0(bytes data)
func (_Padded *PaddedFilterer) WatchTest0(opts *bind.WatchOpts, sink chan<- *PaddedTest0) (event.Subscription, error) {

	logs, sub, err := _Padded.contract.WatchLogs(opts, "Test0")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaddedTest0)
				if err := _Padded.contract.UnpackLog(event, "Test0", log); err != nil {
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

// ParseTest0 is a log parse operation binding the contract event 0x1ea5ab62888366c058d9c8bd1be1081b32b39c8fbc3f91500bbd8a434be39369.
//
// Solidity: event Test0(bytes data)
func (_Padded *PaddedFilterer) ParseTest0(log types.Log) (*PaddedTest0, error) {
	event := new(PaddedTest0)
	if err := _Padded.contract.UnpackLog(event, "Test0", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
