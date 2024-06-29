# Signal Bot _ReplicatorGo_

## Requirements
1. Docker
2. https://github.com/bbernhard/signal-cli-rest-api container run and configured properly (see https://github.com/bbernhard/signal-cli-rest-api/blob/master/README.md)

## Bot configuration
File config.json should be located inside the source root directory to be exported inside container.

### Param's description:

`cli_address` -- http address (with port) of your signal-cli-rest-api.  
`self_number` -- phone number of the sender (you, bot)  
`logs_receiver_number` -- phone number who will receive bot logs (not implemented yet)  
`ignore_older_messages` -- time in ms, older messages, than that, will be ignored (it is for a reason, when messages syncs after service started)  
`is_sending_enabled` -- disables/enables real messages send. It should be false on first service start  
`is_print_messages` -- to enable/disable messages printing to a log  
`enable_debug_messages` -- to enable/disable debug messages printing to a log (which and why ignored, etc.)  
`forwarding` -- array of forwarding groups:   
 >`group_id` -- which group to process messages from  
 >`is_enabled` -- this flag is for disable/enable processing this particular forwarding group  
 >`forwarding_mode` -- can be "__attachments__"/"__messages__"/"__all__" which content we should forward  
 >`receivers_group_ids` -- which groups list will receive forwarded message  
 >`bot_special_addon_msg` -- is applied only in "__attachments__" mode, means which message bot will add to the attachments  
 >`reaction_mark` -- which reaction (should be a smile utf-8 like ➕)  
 >`sender_names` -- forward messages only from given senders names (not recommend to use)  
 >`sender_uuids` -- forward messages only from given senders uuids (recommended to use)  
 >`starts_with` -- forward messages only that starts with given string  
 >`contains` -- forward messages only that contains given string  

### Config example:
```json
{
  "cli_address": "localhost:8080",
  "self_number": "+380123456789",
  "logs_receiver_number": "+380999999999",
  "ignore_older_messages": 240000,
  "is_sending_enabled": true,
  "is_print_messages": true,
  "enable_debug_messages": false,
  "forwarding": [
    {
      "group_id": "Z3JvdXAwX2dyb3VwMF9ncm91cDBfZ3JvdXAwX19fXw==",
      "is_enabled": false,
      "forwarding_mode": "attachments",
      "receivers_group_ids": [
        "VHA3azVUbkJYOGxJenMyN2llTXBjTHBIbVhLakdseA==",
       "dXlQNWNLcDBPYUh4bm4wVHFxTE5PZkdaWVlxTmdMOQ=="
      ],
      "bot_special_addon_msg": " ",
      "reaction_mark": "➕"
    },
    {
      "group_id": "RHVPMTVNVUxXdTI2a3MzenhXb1NZVzNFcEZ0YU9ERg==",
      "is_enabled": true,
      "forwarding_mode": "messages",
      "receivers_group_ids": [
        "b3pwUGtPNDE4a2RZREhNSTZ5OXQxYnJkT01nOHFoSw=="
      ],
      "sender_uuids": [
        "7758825a-09bd-4217-a4a6-fcea0212dfd2"
      ],
      "starts_with": [
        "\uD83D\uDFE2 GOOD BYE ",
        "\uD83D\uDD34Hello",
        "\uD83D\uDD34 HELLO",
        "\uD83D\uDD34HELLO"
      ],
      "contains": [
       "sun"
      ]
    }
  ]
}
```

## Installation instruction
- After signal-cli and REST API container runs properly, go to http://localhost:8080/v1/qrcodelink?device_name=signal-bot and connect your account via phone and QR-Code.  
- Next, we should run this docker-container with command `.\run.cmd` and __disabled sending config__ (disabled messages sending and all forwarding groups processing)
- To get all groups ids, we should wait for some messages in that groups appears (to receive it in Signal), and we can go to http://localhost:8181/groups    
  Notice: if you receive something like "error: Expected a row in result set, but none found.", just restart your containers
- In this page you can see all (known to bot Signal client) groups with it's `name` and `internal_id`. You can use that `internal_id` for your config forwarding params.
- After you finish your configuration, save it and restart ReplicatorGo container.
