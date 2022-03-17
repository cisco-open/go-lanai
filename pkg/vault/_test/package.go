// Blackbox tests. The reason this file exists is to workaround an issue existed since go 1.3:
// https://github.com/golang/go/issues/8279
// Note: this issue has been fixed in 1.17
// These tests cannot be placed in "vault" package because of import cycle between "vault/init" and "vault"

package vault
