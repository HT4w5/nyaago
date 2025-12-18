package meta

import (
	"fmt"
	"testing"
)

func TestMotd(t *testing.T) {
	fmt.Println(getMotd().Lines())
}

func TestBanner(t *testing.T) {
	fmt.Println(getBanner().Lines())
}

func TestBuildInfo(t *testing.T) {
	fmt.Println(getBuildInfo().Lines())
}

func TestGetMetadataMultiline(t *testing.T) {
	fmt.Println(GetMetadataMultiline())
}
