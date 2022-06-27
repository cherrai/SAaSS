package conf

var SocketRouterEventNames = map[string]string{
	"AddFriend":           "AddFriend",
	"AgreeFriend":         "AgreeFriend",
	"DisagreeFriend":      "DisagreeFriend",
	"DeleteFriend":        "DeleteFriend",
	"Error":               "Error",
	"ChatMessage":         "ChatMessage",
	"StartCallingMessage": "StartCallingMessage",
	"LeaveRoom":           "LeaveRoom",
	"JoinRoom":            "JoinRoom",
	"OnAnonymousMessage":  "OnAnonymousMessage",
}
var SocketRouterNamespace = map[string]string{
	"Chat": "/chat",
}
