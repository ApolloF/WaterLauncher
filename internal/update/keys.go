package update

// ReleaseKeys verify release signatures (see signature.go). The private
// key never touches GitHub: it's on the maintainer's PC, with an offline
// backup (tools/release). To rotate, add the new key here in a release
// signed with the old one, and remove the old one a release later.
var ReleaseKeys = mustKeys(
	"9Guduk/WirDFaB1HJi/YrMQDIYvLiPl8teZun9K/jWw=", // made 2026-09-26
)
