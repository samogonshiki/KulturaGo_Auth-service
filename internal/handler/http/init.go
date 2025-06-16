package http

import (
	"os"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/vk"
	"github.com/markbates/goth/providers/yandex"
)

func init() {
	base := os.Getenv("OAUTH_REDIRECT")

	goth.UseProviders(

		vk.New(
			os.Getenv("VK_CLIENT_ID"),
			os.Getenv("VK_CLIENT_SECRET"),
			base+"/vk/callback",
			"email",
		),

		yandex.New(
			os.Getenv("YA_CLIENT_ID"),
			os.Getenv("YA_CLIENT_SECRET"),
			base+"/yandex/callback",
			"login:email",
		),

		// apple.New(
		// 	os.Getenv("APPLE_CLIENT_ID"),
		// 	os.Getenv("APPLE_TEAM_ID"),
		// 	os.Getenv("APPLE_KEY_ID"),
		// 	nil,
		// 	decode(os.Getenv("APPLE_PRIVATE_KEY")),
		// 	base+"/apple/callback",
		// 	apple.ScopeEmail,
		// 	apple.ScopeName,
		// ),
		//! FIX: add apple provider
	)
}

// func decode(b64 string) []byte {
// 	d, _ := base64.StdEncoding.DecodeString(b64)
// 	return d
// }
