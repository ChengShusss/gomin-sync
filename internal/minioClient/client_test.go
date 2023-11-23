package minioClient

import (
	"gomin-sync/internal/config"
	"testing"
)

func TestGetTimeOffset(t *testing.T) {
	// Please use your own config
	config.Config.AccessUser = "user"
	config.Config.AccessPassword = "passwd"
	config.Config.Bucket = "clip"
	config.Config.EndPoint = "example.com"
	config.Config.Prefix = "prefix"
	config.Config.UseSSL = true
	config.Debug = false

	of, err := getTimeOffset()
	if err != nil {
		t.Fatalf("failed to get offset, err: %v\n", err)
	}

	t.Logf("Offset: %v\n", of)

}
