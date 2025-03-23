package processor

var __usage string

func getUsage() string {
	if __usage != "" {
		return __usage
	}

	__usage = "/create vpn\n"
	__usage += "/create peer <peer_ip> <public_key>\n"
	__usage += "/list vpn\n"
	__usage += "/list peers\n"
	__usage += "/delete vpn <vpn_id>\n"
	__usage += "/delete peer <peer_ip>\n"
	__usage += "/settings\n"
	__usage += "/help\n"

	return __usage
}
