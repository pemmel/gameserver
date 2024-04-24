package auth

// Constants representing various identity providers (IDPs).
const (
	ANONYMOUS_IDP_STRING  string = "anonymous"
	APPLE_IDP_STRING      string = "apple"
	EPIC_GAMES_IDP_STRING string = "epic_games"
	FACEBOOK_IDP_STRING   string = "facebook"
	GOOGLE_IDP_STRING     string = "google"
	MICROSOFT_IDP_STRING  string = "microsoft"
	NINTENTDO_IDP_STRING  string = "nintendo"
	PSN_IDP_STRING        string = "psn"
	STEAM_IDP_STRING      string = "steam"
)

// mapIdp is a map that stores the association between identity providers (IDPs)
// and their corresponding registered application IDs.
var mapIdp map[string]string

// init initializes the mapIdp with the registered application IDs for specific
// identity providers.
func init() {
	mapIdp = make(map[string]string)
	mapIdp[ANONYMOUS_IDP_STRING] = ""
	mapIdp[STEAM_IDP_STRING] = "480"
}

// VerifyServerMatch verifies whether the provided identity provider (IDP) and
// application ID match the registered values.
// It returns true if the provided IDP is registered and the provided application
// ID matches the registered application ID for that IDP, otherwise it returns false.
//
// Parameters:
//
//	idp (string): The identity provider (IDP) string.
//	appid (string): The application ID string.
//
// Returns:
//
//	bool: True if the provided IDP is registered and the provided application
//	ID matches the registered application ID for that IDP, otherwise false.
func VerifyServerMatch(idp string, appid string) bool {
	registeredAppId, ok := mapIdp[idp]
	return ok && appid == registeredAppId
}
