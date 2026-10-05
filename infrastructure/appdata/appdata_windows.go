package appdata

// variable is the environment variable naming the user's roaming application data folder.
const variable = "APPDATA"

// base answers the roaming application data folder; false when the environment names none.
func base(lookup func(string) (string, bool)) (string, bool) {
	folder, ok := lookup(variable)
	return folder, ok && folder != ""
}
