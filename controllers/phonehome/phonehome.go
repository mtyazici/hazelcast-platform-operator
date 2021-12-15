package phonehome

import (
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Metrics struct {
	CreatedClusters map[types.UID]bool
}

const (
	CreatedClusterCount = "ccc"
)

func Start(m *Metrics) {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			phoneHome(m)
		}
	}()
}

func phoneHome(m *Metrics) {
	phUrl := "https://phonehome.hazelcast.com/pingOp"

	props := url.Values{
		CreatedClusterCount: {strconv.Itoa(len(m.CreatedClusters))},
	}

	req, err := http.NewRequest("POST", phUrl, strings.NewReader(props.Encode()))
	if err != nil || req == nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{Timeout: time.Second * 10}
	_, err = client.Do(req)
	if err != nil {
		return
	}
}
