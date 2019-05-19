package eospark

import "github.com/eoscanada/eos-go/ecc"

/*
// inline
{
    "errno":0,
    "errmsg":"Success",
    "data":{
        "block_num":37539694,
        "block_time":"2019-01-15T09:47:42.500",
        "eospark_trx_type":"inline",
        "id":"578f2bb176d98254400b0185bf20799c726b863c024c89cdce1f6630c46a5edb",
        "last_irreversible_block":37550784,
        "timestamp":"2019-01-15T09:47:42.500",
        "traces":[
            {
                "account_ram_deltas":[
                    {
                        "account":"dongjie55555",
                        "delta":-268
                    }
                ],
                "act":{
                    "account":"eosio",
                    "authorization":[
                        {
                            "actor":"dongjie55555",
                            "permission":"active"
                        }
                    ],
                    "data":{
                        "owner":"dongjie55555"
                    },
                    "hex_data":"504a2945b9c7264d",
                    "name":"refund"
                },
                "block_num":37539694,
                "block_time":"2019-01-15T09:47:42.500",
                "console":"",
                "context_free":false,
                "elapsed":323,
                "except":null,
                "inline_traces":[
                    {
                        "account_ram_deltas":[

                        ],
                        "act":{
                            "account":"eosio.token",
                            "authorization":[
                                {
                                    "actor":"eosio.stake",
                                    "permission":"active"
                                },
                                {
                                    "actor":"dongjie55555",
                                    "permission":"active"
                                }
                            ],
                            "data":{
                                "from":"eosio.stake",
                                "memo":"unstake",
                                "quantity":"1.8100 EOS",
                                "to":"dongjie55555"
                            },
                            "hex_data":"0014341903ea3055504a2945b9c7264db44600000000000004454f530000000007756e7374616b65",
                            "name":"transfer"
                        },
                        "block_num":37539694,
                        "block_time":"2019-01-15T09:47:42.500",
                        "console":"",
                        "context_free":false,
                        "elapsed":173,
                        "except":null,
                        "inline_traces":[
                            {
                                "account_ram_deltas":[

                                ],
                                "act":{
                                    "account":"eosio.token",
                                    "authorization":[
                                        {
                                            "actor":"eosio.stake",
                                            "permission":"active"
                                        },
                                        {
                                            "actor":"dongjie55555",
                                            "permission":"active"
                                        }
                                    ],
                                    "data":{
                                        "from":"eosio.stake",
                                        "memo":"unstake",
                                        "quantity":"1.8100 EOS",
                                        "to":"dongjie55555"
                                    },
                                    "hex_data":"0014341903ea3055504a2945b9c7264db44600000000000004454f530000000007756e7374616b65",
                                    "name":"transfer"
                                },
                                "block_num":37539694,
                                "block_time":"2019-01-15T09:47:42.500",
                                "console":"",
                                "context_free":false,
                                "elapsed":7,
                                "except":null,
                                "inline_traces":[

                                ],
                                "producer_block_id":"023ccf6e2382bcceeb01d247234d44e9817db2b82cdf39d639120d6580a2f190",
                                "receipt":{
                                    "abi_sequence":2,
                                    "act_digest":"f639d73480395170668cd54b71ca90e9f2f166306dc88143adb161892c33df47",
                                    "auth_sequence":[
                                        [
                                            "dongjie55555",
                                            552
                                        ],
                                        [
                                            "eosio.stake",
                                            434763
                                        ]
                                    ],
                                    "code_sequence":2,
                                    "global_sequence":4085793036,
                                    "receiver":"eosio.stake",
                                    "recv_sequence":1998290
                                },
                                "trx_id":"578f2bb176d98254400b0185bf20799c726b863c024c89cdce1f6630c46a5edb"
                            },
                            {
                                "account_ram_deltas":[

                                ],
                                "act":{
                                    "account":"eosio.token",
                                    "authorization":[
                                        {
                                            "actor":"eosio.stake",
                                            "permission":"active"
                                        },
                                        {
                                            "actor":"dongjie55555",
                                            "permission":"active"
                                        }
                                    ],
                                    "data":{
                                        "from":"eosio.stake",
                                        "memo":"unstake",
                                        "quantity":"1.8100 EOS",
                                        "to":"dongjie55555"
                                    },
                                    "hex_data":"0014341903ea3055504a2945b9c7264db44600000000000004454f530000000007756e7374616b65",
                                    "name":"transfer"
                                },
                                "block_num":37539694,
                                "block_time":"2019-01-15T09:47:42.500",
                                "console":"",
                                "context_free":false,
                                "elapsed":6,
                                "except":null,
                                "inline_traces":[

                                ],
                                "producer_block_id":"023ccf6e2382bcceeb01d247234d44e9817db2b82cdf39d639120d6580a2f190",
                                "receipt":{
                                    "abi_sequence":2,
                                    "act_digest":"f639d73480395170668cd54b71ca90e9f2f166306dc88143adb161892c33df47",
                                    "auth_sequence":[
                                        [
                                            "dongjie55555",
                                            553
                                        ],
                                        [
                                            "eosio.stake",
                                            434764
                                        ]
                                    ],
                                    "code_sequence":2,
                                    "global_sequence":4085793037,
                                    "receiver":"dongjie55555",
                                    "recv_sequence":296
                                },
                                "trx_id":"578f2bb176d98254400b0185bf20799c726b863c024c89cdce1f6630c46a5edb"
                            }
                        ],
                        "producer_block_id":"023ccf6e2382bcceeb01d247234d44e9817db2b82cdf39d639120d6580a2f190",
                        "receipt":{
                            "abi_sequence":2,
                            "act_digest":"f639d73480395170668cd54b71ca90e9f2f166306dc88143adb161892c33df47",
                            "auth_sequence":[
                                [
                                    "dongjie55555",
                                    551
                                ],
                                [
                                    "eosio.stake",
                                    434762
                                ]
                            ],
                            "code_sequence":2,
                            "global_sequence":4085793035,
                            "receiver":"eosio.token",
                            "recv_sequence":678736644
                        },
                        "trx_id":"578f2bb176d98254400b0185bf20799c726b863c024c89cdce1f6630c46a5edb"
                    }
                ],
                "producer_block_id":"023ccf6e2382bcceeb01d247234d44e9817db2b82cdf39d639120d6580a2f190",
                "receipt":{
                    "abi_sequence":12,
                    "act_digest":"eecbc3f0f778aee53fbc5737da6fec57b11c1346e6b04c5d0cd43ae0365f9569",
                    "auth_sequence":[
                        [
                            "dongjie55555",
                            550
                        ]
                    ],
                    "code_sequence":11,
                    "global_sequence":4085793034,
                    "receiver":"eosio",
                    "recv_sequence":45288353
                },
                "trx_id":"578f2bb176d98254400b0185bf20799c726b863c024c89cdce1f6630c46a5edb"
            },
            {
                "account_ram_deltas":[

                ],
                "act":{
                    "account":"eosio.token",
                    "authorization":[
                        {
                            "actor":"eosio.stake",
                            "permission":"active"
                        },
                        {
                            "actor":"dongjie55555",
                            "permission":"active"
                        }
                    ],
                    "data":{
                        "from":"eosio.stake",
                        "memo":"unstake",
                        "quantity":"1.8100 EOS",
                        "to":"dongjie55555"
                    },
                    "hex_data":"0014341903ea3055504a2945b9c7264db44600000000000004454f530000000007756e7374616b65",
                    "name":"transfer"
                },
                "block_num":37539694,
                "block_time":"2019-01-15T09:47:42.500",
                "console":"",
                "context_free":false,
                "elapsed":173,
                "except":null,
                "inline_traces":[
                    {
                        "account_ram_deltas":[

                        ],
                        "act":{
                            "account":"eosio.token",
                            "authorization":[
                                {
                                    "actor":"eosio.stake",
                                    "permission":"active"
                                },
                                {
                                    "actor":"dongjie55555",
                                    "permission":"active"
                                }
                            ],
                            "data":{
                                "from":"eosio.stake",
                                "memo":"unstake",
                                "quantity":"1.8100 EOS",
                                "to":"dongjie55555"
                            },
                            "hex_data":"0014341903ea3055504a2945b9c7264db44600000000000004454f530000000007756e7374616b65",
                            "name":"transfer"
                        },
                        "block_num":37539694,
                        "block_time":"2019-01-15T09:47:42.500",
                        "console":"",
                        "context_free":false,
                        "elapsed":7,
                        "except":null,
                        "inline_traces":[

                        ],
                        "producer_block_id":"023ccf6e2382bcceeb01d247234d44e9817db2b82cdf39d639120d6580a2f190",
                        "receipt":{
                            "abi_sequence":2,
                            "act_digest":"f639d73480395170668cd54b71ca90e9f2f166306dc88143adb161892c33df47",
                            "auth_sequence":[
                                [
                                    "dongjie55555",
                                    552
                                ],
                                [
                                    "eosio.stake",
                                    434763
                                ]
                            ],
                            "code_sequence":2,
                            "global_sequence":4085793036,
                            "receiver":"eosio.stake",
                            "recv_sequence":1998290
                        },
                        "trx_id":"578f2bb176d98254400b0185bf20799c726b863c024c89cdce1f6630c46a5edb"
                    },
                    {
                        "account_ram_deltas":[

                        ],
                        "act":{
                            "account":"eosio.token",
                            "authorization":[
                                {
                                    "actor":"eosio.stake",
                                    "permission":"active"
                                },
                                {
                                    "actor":"dongjie55555",
                                    "permission":"active"
                                }
                            ],
                            "data":{
                                "from":"eosio.stake",
                                "memo":"unstake",
                                "quantity":"1.8100 EOS",
                                "to":"dongjie55555"
                            },
                            "hex_data":"0014341903ea3055504a2945b9c7264db44600000000000004454f530000000007756e7374616b65",
                            "name":"transfer"
                        },
                        "block_num":37539694,
                        "block_time":"2019-01-15T09:47:42.500",
                        "console":"",
                        "context_free":false,
                        "elapsed":6,
                        "except":null,
                        "inline_traces":[

                        ],
                        "producer_block_id":"023ccf6e2382bcceeb01d247234d44e9817db2b82cdf39d639120d6580a2f190",
                        "receipt":{
                            "abi_sequence":2,
                            "act_digest":"f639d73480395170668cd54b71ca90e9f2f166306dc88143adb161892c33df47",
                            "auth_sequence":[
                                [
                                    "dongjie55555",
                                    553
                                ],
                                [
                                    "eosio.stake",
                                    434764
                                ]
                            ],
                            "code_sequence":2,
                            "global_sequence":4085793037,
                            "receiver":"dongjie55555",
                            "recv_sequence":296
                        },
                        "trx_id":"578f2bb176d98254400b0185bf20799c726b863c024c89cdce1f6630c46a5edb"
                    }
                ],
                "producer_block_id":"023ccf6e2382bcceeb01d247234d44e9817db2b82cdf39d639120d6580a2f190",
                "receipt":{
                    "abi_sequence":2,
                    "act_digest":"f639d73480395170668cd54b71ca90e9f2f166306dc88143adb161892c33df47",
                    "auth_sequence":[
                        [
                            "dongjie55555",
                            551
                        ],
                        [
                            "eosio.stake",
                            434762
                        ]
                    ],
                    "code_sequence":2,
                    "global_sequence":4085793035,
                    "receiver":"eosio.token",
                    "recv_sequence":678736644
                },
                "trx_id":"578f2bb176d98254400b0185bf20799c726b863c024c89cdce1f6630c46a5edb"
            },
            {
                "account_ram_deltas":[

                ],
                "act":{
                    "account":"eosio.token",
                    "authorization":[
                        {
                            "actor":"eosio.stake",
                            "permission":"active"
                        },
                        {
                            "actor":"dongjie55555",
                            "permission":"active"
                        }
                    ],
                    "data":{
                        "from":"eosio.stake",
                        "memo":"unstake",
                        "quantity":"1.8100 EOS",
                        "to":"dongjie55555"
                    },
                    "hex_data":"0014341903ea3055504a2945b9c7264db44600000000000004454f530000000007756e7374616b65",
                    "name":"transfer"
                },
                "block_num":37539694,
                "block_time":"2019-01-15T09:47:42.500",
                "console":"",
                "context_free":false,
                "elapsed":7,
                "except":null,
                "inline_traces":[

                ],
                "producer_block_id":"023ccf6e2382bcceeb01d247234d44e9817db2b82cdf39d639120d6580a2f190",
                "receipt":{
                    "abi_sequence":2,
                    "act_digest":"f639d73480395170668cd54b71ca90e9f2f166306dc88143adb161892c33df47",
                    "auth_sequence":[
                        [
                            "dongjie55555",
                            552
                        ],
                        [
                            "eosio.stake",
                            434763
                        ]
                    ],
                    "code_sequence":2,
                    "global_sequence":4085793036,
                    "receiver":"eosio.stake",
                    "recv_sequence":1998290
                },
                "trx_id":"578f2bb176d98254400b0185bf20799c726b863c024c89cdce1f6630c46a5edb"
            },
            {
                "account_ram_deltas":[

                ],
                "act":{
                    "account":"eosio.token",
                    "authorization":[
                        {
                            "actor":"eosio.stake",
                            "permission":"active"
                        },
                        {
                            "actor":"dongjie55555",
                            "permission":"active"
                        }
                    ],
                    "data":{
                        "from":"eosio.stake",
                        "memo":"unstake",
                        "quantity":"1.8100 EOS",
                        "to":"dongjie55555"
                    },
                    "hex_data":"0014341903ea3055504a2945b9c7264db44600000000000004454f530000000007756e7374616b65",
                    "name":"transfer"
                },
                "block_num":37539694,
                "block_time":"2019-01-15T09:47:42.500",
                "console":"",
                "context_free":false,
                "elapsed":6,
                "except":null,
                "inline_traces":[

                ],
                "producer_block_id":"023ccf6e2382bcceeb01d247234d44e9817db2b82cdf39d639120d6580a2f190",
                "receipt":{
                    "abi_sequence":2,
                    "act_digest":"f639d73480395170668cd54b71ca90e9f2f166306dc88143adb161892c33df47",
                    "auth_sequence":[
                        [
                            "dongjie55555",
                            553
                        ],
                        [
                            "eosio.stake",
                            434764
                        ]
                    ],
                    "code_sequence":2,
                    "global_sequence":4085793037,
                    "receiver":"dongjie55555",
                    "recv_sequence":296
                },
                "trx_id":"578f2bb176d98254400b0185bf20799c726b863c024c89cdce1f6630c46a5edb"
            }
        ],
        "trx":{
            "receipt":{
                "cpu_usage_us":203,
                "net_usage_words":0,
                "status":"executed",
                "trx":[
                    0,
                    "578f2bb176d98254400b0185bf20799c726b863c024c89cdce1f6630c46a5edb"
                ]
            }
        }
    }
}


{
    "errno":0,
    "errmsg":"Success",
    "data":{
        "block_num":37385454,
        "cpu_usage_us":307,
        "eospark_trx_type":"ordinary",
        "net_usage_words":17,
        "status":"executed",
        "timestamp":"2019-01-14T12:21:14.500",
        "trx":{
            "compression":"none",
            "context_free_data":[

            ],
            "id":"2566639b83742ad914b2c50665ec4adea9110543509aa99c5ec1a79abc4f7d0e",
            "packed_context_free_data":"",
            "packed_trx":"d87e3c5cec74ed11b156000000000100a6823403ea3055000000572d3ccdcd01301d451e221d315500000000a8ed323228301d451e221d3155504a2945b9c7264daa0a00000000000004454f53000000000762616c616e636500",
            "signatures":[
                "SIG_K1_Kg8iv4hkNWJKFRdAFd6jwTofTQfHRYED3zNJiejxUZwchHVWrHQaN1Q7R6teBawSUwrzBiLxG4mwEGmwS7MDkkfWfE6Lks"
            ],
            "transaction":{
                "actions":[
                    {
                        "account":"eosio.token",
                        "authorization":[
                            {
                                "actor":"eosluckycoin",
                                "permission":"active"
                            }
                        ],
                        "data":{
                            "from":"eosluckycoin",
                            "memo":"balance",
                            "quantity":"0.2730 EOS",
                            "to":"dongjie55555"
                        },
                        "hex_data":"301d451e221d3155504a2945b9c7264daa0a00000000000004454f53000000000762616c616e6365",
                        "name":"transfer"
                    }
                ],
                "context_free_actions":[

                ],
                "delay_sec":0,
                "expiration":"2019-01-14T12:21:44",
                "max_cpu_usage_ms":0,
                "max_net_usage_words":0,
                "ref_block_num":29932,
                "ref_block_prefix":1454445037,
                "transaction_extensions":[

                ]
            }
        }
    }
}
*/

/* the eospark api get_account_related_trx_info response struct*/
type GetAccountRelatedTrxInfoResp struct {
	TraceCount		int32 		`json:"trace_count"`
	TraceList	    []*Trace   	`json:"trace_list"`
}

type Trace struct {
	TrxId			string			`json:"trx_id"`
	Timestamp		JSONTime		`json:"timestamp"`
	Receiver		AccountName		`json:"receiver"`
	Sender			AccountName		`json:"sender"`
	Code			string			`json:"code"`
	Quantity		string			`json:"quantity"`
	Memo			string			`json:"memo"`
	Symbol			string			`json:"symbol"`
	Status			string			`json:"status"`
	BlockNum		uint32			`json:"block_num"`
}

/* the eospark api get_token_list response struct*/
type GetTokenListResp struct {
	SymbolList []*Symbol `json:"symbol_list"`
}

type Symbol struct {
	Symbol 		string		`json:"symbol"`
	Code 		string		`json:"code"`
	Balance 	string		`json:"balance"`
}

/* the eospark api get_account_resource_info response struct*/
type GetAccountResourceInfoResp struct {
	Ram 		AccountResourceLimit	`json:"ram"`
	Net 		AccountResourceLimit	`json:"net"`
	Cpu 		AccountResourceLimit	`json:"cpu"`
	Staked 		AccountStaked			`json:"staked"`
	UnStaked 	AccountUnStaked			`json:"un_staked"`
}

/* the eospark api get_transaction_detail_info response struct*/
type GetTransactionDetailInfoResp struct {
	BlockNum					uint32					`json:"block_num"`
	EosparkTrxType				string					`json:"eospark_trx_type"`
	Timestamp					JSONTime				`json:"timestamp"`
	Trx 						Trx						`json:"trx,omitempty"`
	// eospark_trx_type为inline时
	BlockTime 					JSONTime				`json:"block_time"`
	ID							Checksum256				`json:"id,omitempty"`
	LastIrreversibleBlock		uint32					`json:"last_irreversible_block,omitempty"`
	Traces						[]*TransactionTrace		`json:"traces"`
	// eospark_trx_type为ordinary时
	CpuUsageUs 					uint32					`json:"cpu_usage_us,omitempty"`
	NetUsageWords				uint32					`json:"net_usage_words,omitempty"`
	Status 						TransactionStatus		`json:"status,omitempty"`
}

type Trx struct {
	// eospark_trx_type为inline时
	Receipt 					TrxReceipt				`json:"receipt,omitempty"`
	// eospark_trx_type为ordinary时
	Compression 				string					`json:"compression,omitempty"`
	ContextFreeData 			[]HexBytes  			`json:"context_free_data,omitempty"`
	ID 							Checksum256				`json:"id,omitempty"`
	PackedContextFreeData		HexBytes				`json:"packed_context_free_data,omitempty"`
	PackedTrx					string					`json:"packed_trx,omitempty"`
	Signatures      			[]ecc.Signature 		`json:"signatures,omitempty"`
	Transaction					TrxTransaction			`json:"transaction,omitempty"`
}

type TrxTransaction struct {
	Actions        				[]*Action     				`json:"actions"`
	ContextFreeActions 			[]*Action    				`json:"context_free_actions"`
	DelaySec					uint32						`json:"delay_sec"`
	Expiration					string						`json:"expiration"`
	TransactionExtensions      	[]*TransactionExtension 	`json:"transaction_extensions"`
	MaxCpuUsageMs				uint32						`json:"max_cpu_usage_ms"`
	MaxNetUsageWords			uint32						`json:"max_net_usage_words"`
	RefBlockNum					uint32						`json:"ref_block_num"`
	RefBlockPrefix				uint32 						`json:"ref_block_prefix"`
}

type TrxReceipt struct {
	CpuUsageUs		int32					`json:"cpu_usage_us"`
	NetUsageWords	int32					`json:"net_usage_words"`
	Status 			TransactionStatus		`json:"status"`
	Trx 			[]interface{}			`json:"trx"`
}

type Action struct {
	Account       	AccountName       	`json:"account"`
	Name          	ActionName        	`json:"name"`
	Authorization 	[]*PermissionLevel 	`json:"authorization,omitempty"`
	HexData  		HexBytes    		`json:"hex_data,omitempty"`
	Data     		interface{} 		`json:"data,omitempty"`
}

type TransactionExtension struct {
	Type 	uint16   	`json:"type"`
	Data 	HexBytes 	`json:"data"`
}

type TransactionTrace struct {
	AccountRamDeltas 		[]*TransactionAccountRamDelta		`json:"account_ram_deltas"`
	Action 					Action								`json:"act"`
	BlockNum				uint32								`json:"block_num"`
	BlockTime 				JSONTime							`json:"block_time"`
	Console 				string								`json:"console"`
	ContextFree 			bool								`json:"context_free"`
	Elapsed 				int32								`json:"elapsed"`
	Except					interface{}							`json:"except"`
	InlineTraces 			[]*TransactionTrace					`json:"inline_traces"`
	ProducerBlockId			string								`json:"producer_block_id"`
	Receipt 				TransactionReceipt 					`json:"receipt"`
	TrxId					Checksum256							`json:"trx_id"`
}

type TransactionAccountRamDelta struct {
	Account 	AccountName		`json:"account"`
	Delta		int32			`json:"delta"`
}

type TransactionReceipt struct {
	AbiSequence			int32			`json:"abi_sequence"`
	ActDigest			string			`json:"act_digest"`
	AuthSequence		[]interface{}	`json:"auth_sequence"`
	CodeSequence		int32			`json:"code_sequence"`
	GlobalSequence		uint32			`json:"global_sequence"`
	Receiver 			Name			`json:"receiver"`
	RecvSequence		uint32			`json:"recv_sequence"`
}