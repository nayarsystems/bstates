package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/nayarsystems/bstates"
)

func main() {
	schemaRaw := `{
  "decodedFields": [
    {
      "decoder": "IntMap",
      "name": "ACT",
      "params": {
        "from": "ACT_RAW",
        "mapId": "ACT_MAP"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "CPSI",
      "params": {
        "from": "CPSI_RAW"
      }
    },
    {
      "decoder": "IntMap",
      "name": "CPSI_OPMODE",
      "params": {
        "from": "CPSI_OPMODE_RAW",
        "mapId": "CPSI_OPMODE_MAP"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "CPSI_SYSTEMMODE",
      "params": {
        "from": "CPSI_SYSTEMMODE_RAW"
      }
    },
    {
      "decoder": "IntMap",
      "name": "CREG",
      "params": {
        "from": "CREG_RAW",
        "mapId": "CREG_MAP"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "ICC",
      "params": {
        "from": "ICC_RAW"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "IMSI",
      "params": {
        "from": "IMSI_RAW"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "OPERATOR",
      "params": {
        "from": "OPERATOR_RAW"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "P0_ERROR_STR",
      "params": {
        "from": "P0_ERROR_MSG"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "P0_PRODUCT_STR",
      "params": {
        "from": "P0_PRODUCT"
      }
    },
    {
      "decoder": "IntMap",
      "name": "P0_TYPE",
      "params": {
        "from": "P0_DEV_TYPE",
        "mapId": "DEVTYPE_MAP"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "P0_VERSION_STR",
      "params": {
        "from": "P0_VERSION"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "P1_ERROR_STR",
      "params": {
        "from": "P1_ERROR_MSG"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "P1_PRODUCT_STR",
      "params": {
        "from": "P1_PRODUCT"
      }
    },
    {
      "decoder": "IntMap",
      "name": "P1_TYPE",
      "params": {
        "from": "P1_DEV_TYPE",
        "mapId": "DEVTYPE_MAP"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "P1_VERSION_STR",
      "params": {
        "from": "P1_VERSION"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "P2_ERROR_STR",
      "params": {
        "from": "P2_ERROR_MSG"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "P2_PRODUCT_STR",
      "params": {
        "from": "P2_PRODUCT"
      }
    },
    {
      "decoder": "IntMap",
      "name": "P2_TYPE",
      "params": {
        "from": "P2_DEV_TYPE",
        "mapId": "DEVTYPE_MAP"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "P2_VERSION_STR",
      "params": {
        "from": "P2_VERSION"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "P3_ERROR_STR",
      "params": {
        "from": "P3_ERROR_MSG"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "P3_PRODUCT_STR",
      "params": {
        "from": "P3_PRODUCT"
      }
    },
    {
      "decoder": "IntMap",
      "name": "P3_TYPE",
      "params": {
        "from": "P3_DEV_TYPE",
        "mapId": "DEVTYPE_MAP"
      }
    },
    {
      "decoder": "BufferToString",
      "name": "P3_VERSION_STR",
      "params": {
        "from": "P3_VERSION"
      }
    },
    {
      "decoder": "NumberToUnixTsMs",
      "name": "TIMESTAMP_MS",
      "params": {
        "factor": 1,
        "from": "TIMESTAMP_RAW",
        "year": 2024
      }
    }
  ],
  "decoderIntMaps": {
    "ACT_MAP": {
      "0": "GSM",
      "1": "GSM Compact",
      "2": "UTRAN",
      "7": "EUTRAN",
      "8": "CDMA_HDR"
    },
    "CPSI_OPMODE_MAP": {
      "1": "Online",
      "2": "Offline",
      "3": "Factory Test Mode",
      "4": "Reset",
      "5": "Low Power Mode"
    },
    "CREG_MAP": {
      "0": "NOREG_NOSEARCH",
      "1": "REG_HOME_NET",
      "2": "NOREG_SEARCH",
      "3": "NOREG_DENIED",
      "5": "REG_ROAMING"
    },
    "DEVTYPE_MAP": {
      "1": "RP2040"
    }
  },
  "encoderPipeline": "t:z",
  "fields": [
    {
      "defaultValue": 0,
      "name": "TIMESTAMP_RAW",
      "size": 48,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "BOOT_COUNTER",
      "size": 16,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "EVENT_COUNTER",
      "size": 32,
      "type": "uint"
    },
    {
      "defaultValue": false,
      "name": "MENHIR_LINKUP",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": false,
      "name": "MENHIR_RESTARTING",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": false,
      "name": "MENHIR_READY",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": false,
      "name": "PRIORITY_EVENT",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": false,
      "name": "NOTHING_TO_SEND",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "IMSI_RAW",
      "size": 136,
      "type": "buffer"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "ICC_RAW",
      "size": 184,
      "type": "buffer"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "OPERATOR_RAW",
      "size": 208,
      "type": "buffer"
    },
    {
      "defaultValue": 0,
      "name": "CREG_RAW",
      "size": 4,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "ACT_RAW",
      "size": 4,
      "type": "uint"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "CPSI_RAW",
      "size": 400,
      "type": "buffer"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAA=",
      "name": "CPSI_SYSTEMMODE_RAW",
      "size": 112,
      "type": "buffer"
    },
    {
      "defaultValue": 0,
      "name": "CPSI_OPMODE_RAW",
      "size": 3,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "CPSI_MCC",
      "size": 32,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "CPSI_MNC",
      "size": 32,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "CPSI_LAC_TAC",
      "size": 32,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "CPSI_CELLID",
      "size": 32,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "CPSI_SCELLID",
      "size": 32,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "CPSI_PCELLID",
      "size": 32,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "CPSI_SID",
      "size": 32,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "CSQ",
      "size": 8,
      "type": "int"
    },
    {
      "defaultValue": false,
      "name": "PPP",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": false,
      "name": "CLOUD",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": 0,
      "name": "TUN_TX",
      "size": 32,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "TUN_RX",
      "size": 32,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "P0_DEV_TYPE",
      "size": 4,
      "type": "uint"
    },
    {
      "defaultValue": false,
      "name": "P0_PRESENT",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": false,
      "name": "P0_FLASHMODE",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": false,
      "name": "P0_UPDATING",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": 0,
      "name": "P0_PID",
      "size": 16,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "P0_VID",
      "size": 16,
      "type": "uint"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "P0_PRODUCT",
      "size": 400,
      "type": "buffer"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "P0_VERSION",
      "size": 400,
      "type": "buffer"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "P0_ERROR_MSG",
      "size": 400,
      "type": "buffer"
    },
    {
      "defaultValue": 0,
      "name": "P1_DEV_TYPE",
      "size": 4,
      "type": "uint"
    },
    {
      "defaultValue": false,
      "name": "P1_PRESENT",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": false,
      "name": "P1_FLASHMODE",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": false,
      "name": "P1_UPDATING",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": 0,
      "name": "P1_PID",
      "size": 16,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "P1_VID",
      "size": 16,
      "type": "uint"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "P1_PRODUCT",
      "size": 400,
      "type": "buffer"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "P1_VERSION",
      "size": 400,
      "type": "buffer"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "P1_ERROR_MSG",
      "size": 400,
      "type": "buffer"
    },
    {
      "defaultValue": 0,
      "name": "P2_DEV_TYPE",
      "size": 4,
      "type": "uint"
    },
    {
      "defaultValue": false,
      "name": "P2_PRESENT",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": false,
      "name": "P2_FLASHMODE",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": false,
      "name": "P2_UPDATING",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": 0,
      "name": "P2_PID",
      "size": 16,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "P2_VID",
      "size": 16,
      "type": "uint"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "P2_PRODUCT",
      "size": 400,
      "type": "buffer"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "P2_VERSION",
      "size": 400,
      "type": "buffer"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "P2_ERROR_MSG",
      "size": 400,
      "type": "buffer"
    },
    {
      "defaultValue": 0,
      "name": "P3_DEV_TYPE",
      "size": 4,
      "type": "uint"
    },
    {
      "defaultValue": false,
      "name": "P3_PRESENT",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": false,
      "name": "P3_FLASHMODE",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": false,
      "name": "P3_UPDATING",
      "size": 1,
      "type": "bool"
    },
    {
      "defaultValue": 0,
      "name": "P3_PID",
      "size": 16,
      "type": "uint"
    },
    {
      "defaultValue": 0,
      "name": "P3_VID",
      "size": 16,
      "type": "uint"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "P3_PRODUCT",
      "size": 400,
      "type": "buffer"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "P3_VERSION",
      "size": 400,
      "type": "buffer"
    },
    {
      "defaultValue": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
      "name": "P3_ERROR_MSG",
      "size": 400,
      "type": "buffer"
    }
  ],
  "version": "2.0"
}`

	// first we create an schema
	var schema bstates.StateSchema
	if err := json.Unmarshal([]byte(schemaRaw), &schema); err != nil {
		perrf("can't parse schema: %v\n", err)
	}
	// now we create an state queue
	queue := bstates.CreateStateQueue(&schema)

	// Decode the queue (in this case there is only one state in the queue)
	binaryBlob, _ := base64.StdEncoding.DecodeString("H4sIAAAAAAAC/2JgTQ961cjAy8DAwMhAMWD/wUBL4MAwCkYBfQEgAAD//0pshnshAwAA")
	queue.Decode(binaryBlob)

	// pop an event from the queu

	state, _ := queue.Pop()
	printState("Decoded state: ", state)

}

func printState(name string, state *bstates.State) {
	msiState, _ := state.ToMsi()
	msiStateStr, _ := json.MarshalIndent(msiState, "", "  ")

	fmt.Printf("%s: %v\n", name, string(msiStateStr))

	raw, _ := state.Encode()

	fmt.Printf("%s (RAW): [ ", name)
	for _, n := range raw {
		fmt.Printf("%08b ", n)
	}
	fmt.Printf("]\n---------------\n")
}

func perrf(msg string, a ...interface{}) {
	fmt.Printf(msg, a...)
	os.Exit(1)
}
