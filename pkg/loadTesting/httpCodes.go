package loadTesting

// The contents of the map follow the func.
type codeTable struct {
	descr  string
	create bool
}

// codeDescr returns a short description for an http code and a flag
// for mkLoadTestFiles
// FIXME, the flag is usefull, but the string is unneeded
func codeDescr(errorValue int) (string, bool) {
	val, present := codeMap[errorValue]
	if !present {
		return string(errorValue) + " not defined", false
	}
	return string(errorValue) + " " + val.descr, val.create
}

var codeMap = map[int]codeTable{
	200: {"OK", true},
	201: {"created", false},
	202: {"accepted", false},
	203: {"non-authoritative information", true},
	204: {"no content", false},
	205: {"reset content", false},
	206: {"partial content delivered to you", true},

	304: {"not modified since you asked last", true},
	305: {"use proxy", false},
	306: {"switch proxy", false},

	400: {"bad request", false},
	401: {"unauthorized", false},
	402: {"payment required", false},
	403: {"forbidden", false},
	404: {"not found", false},
	405: {"method not allowed", false},
	406: {"not acceptable", false},
	407: {"proxy authentication required", false},
	408: {"timed out", true},
	409: {"conflict", false},
	410: {"gone", false},
	411: {"length required", false},
	412: {"precondition failed", false},
	413: {"payload too large", false},
	414: {"uri too long (", false},
	415: {"unsupported media type", false},
	416: {"range not satisfiable", false},
	417: {"expectation failed", false},
	418: {"I'm a teapot", false},
	420: {"method failure (spring framework), enhance your calm (twitter)", false},
	421: {"misdirected request", false},
	422: {"unprocessable entity", false},
	423: {"locked", false},
	424: {"failed dependency", false},
	426: {"upgrade required", false},
	428: {"precondition required", false},
	429: {"too many requests", false},
	431: {"request header fields too large", false},
	440: {"login time-out", false},
	444: {"no response", false},
	449: {" retry with", false},
	451: {"redirect, unavailable for legal reasons", false},
	450: {"blocked by windows parental controls (microsoft)", false},
	495: {"ssl certificate error", false},
	496: {" ssl certificate required", false},
	497: {"http request sent to https port", false},
	498: {"invalid token (esri)", false},
	499: {"client closed and gave up (nginx)", true},

	500: {"internal server error", true},
	501: {"not implemented", false},
	502: {"bad gateway", false},
	503: {"service unavailable", false},
	504: {"gateway time-out", false},
	505: {"http version not supported", false},
	506: {"variant also negotiates (negotiation loop)", false},
	507: {"insufficient storage", false},
	508: {"loop detected", false},
	509: {" bandwidth limit exceeded (apache web server/cpanel)", false},
	510: {"not extended", false},
	511: {"network authentication required", false},
	520: {" unknown error", false},
	521: {" web server is down", false},
	522: {" connection timed out", false},
	523: {" origin is unreachable", false},
	524: {" a Timeout occurred", false},
	525: {" ssl handshake failed", false},
	526: {" invalid ssl certificate", false},
	527: {" railgun error", false},
	530: {" site is frozen", false},
	598: {" (informal convention) network read Timeout error", false},
	599: {" (informal convention) network connect Timeout error", false},
}
