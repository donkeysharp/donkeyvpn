package processor

var __usage string

func getUsage() string {
	if __usage != "" {
		return __usage
	}

	__usage = "/create vpn\n"
	__usage += "/create peer <vpn_ip> <public_key>\n"
	__usage += "/list vpn\n"
	__usage += "/list peers\n"
	__usage += "/delete vpn\n"
	__usage += "/delete peer <vpn_ip>\n"

	return __usage
}
