package plugin

import (
	"strconv"
)

func panelOperator(operator string, args []string) any {
	send := struct {
		PostType    string   `json:"post_type"`
		MessageType string   `json:"message_type"`
		SubType     string   `json:"sub_type"`
		CommandArgs []string `json:"command_args"`
	}{
		PostType:    "operator",
		MessageType: "return",
	}
	switch operator {
	case "get_self_id":
		send.SubType = "self_id"
		for selfId == 0 {
		}
		send.CommandArgs = []string{strconv.FormatInt(selfId, 10)}
	}
	return send
}
