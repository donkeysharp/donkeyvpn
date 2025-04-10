package processor

var __usage string

func getUsage() string {
	if __usage != "" {
		return __usage
	}

	__usage = "/create vpn\n"
	__usage += "/create peer <peer\\_ip> <public\\_key>\n"
	__usage += "/list vpn\n"
	__usage += "/list peers\n"
	__usage += "/delete vpn <vpn\\_id or all>\n"
	__usage += "/delete peer <peer\\_ip>\n"
	__usage += "/settings\n"
	__usage += "/docs\n"
	__usage += "/help\n"

	return __usage
}
