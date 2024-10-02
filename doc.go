/*
Package bstates implements a parser for Idefix event blobs.

# Overview

Each binary blob is an event queue which is composed of one or more state encoded using the same schema. States represent a system state in a point of time. Schemas are json files which define States binary formats. Here there is an example schema which defines a 102bit [State] composed of 4 fields:

		{
			"version": "2.0",
			"encoderPipeline": "t:z",

			"decodedFields":
			{
				{
					"name": "MESSAGE"
					"from": "MESSAGE_BUFFER",
					"decoder": "BufferToString"

				},
				{
					"name": "STATE",
					"decoder": "IntMap",
					"params": {
						"from": "STATE_CODE",
						"mapId": "STATE_MAP"
				}
			},

			"decoderIntMaps":
				{
					"STATE_MAP": {
						"0" : "IDLE",
						"1" : "STOPPED",
						"2" : "RUNNING"
					}
				},

		    "fields": [
		        {
		            "name": "3BITS_INT",
		            "type": "int",
		            "size": 3
		        },
		        {
		            "name": "STATE_CODE",
		            "type": "uint",
		            "size": 2
		        },
		        {
		            "name": "BOOL",
		            "type": "bool"
		        },
		        {
		            "name": "MESSAGE_BUFFER",
		            "type": "buffer",
		            "size": 96
		        }
		    ]
	}

Decoders allow for complex codification of data, those saving space. In the above example there is a decodedField MESAGE which allows us to read MESSAGE_BUFFER as a string. Decoders work by executing the "decoder" parameter on the "from" field. Currently there are three types of [Decoder].

States are the objects which hold data. They are composed by a [frame.Frame] which holds the data and a [StateSchema] that specifies a codification. [StateField] and [DecodedStateField] both can be retrieved using the [State.Get] function.

All objects in this library, can be serialized as JSON or as MSI map[string]interface{}
*/
package bstates
