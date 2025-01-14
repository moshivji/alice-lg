package birdwatcher

import (
	"encoding/json"
	"io"

	"github.com/alice-lg/alice-lg/pkg/api"
)

func parseRoutesResponseStream(
	body io.Reader,
	config Config,
) (*api.Meta, api.Routes, error) {
	dec := json.NewDecoder(body)
	meta := &api.Meta{}
	routes := api.Routes{}

	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}

		// Parse API meta data
		if t == "api" {
			api := make(map[string]interface{})
			if err := dec.Decode(&api); err != nil {
				return nil, nil, err
			}
			cacheStatus, _ := parseCacheStatus(api, config)
			meta.Version = api["Version"].(string)
			meta.ResultFromCache = api["result_from_cache"].(bool)
			meta.CacheStatus = cacheStatus
		}

		if t == "ttl" {
			var ttlD string
			if err := dec.Decode(&ttlD); err != nil {
				return nil, nil, err
			}
			ttl, err := parseServerTime(
				ttlD,
				config.ServerTime,
				config.Timezone,
			)
			if err != nil {
				return nil, nil, err
			}
			meta.TTL = ttl
		}

		// Route data
		if t == "routes" {
			// Read array begin
			_, err := dec.Token()
			if err == io.EOF {
				break
			}

			for dec.More() {
				rdata := make(map[string]interface{})
				if err := dec.Decode(&rdata); err != nil {
					return nil, nil, err
				}
				route := parseRouteData(rdata, config, false)
				routes = append(routes, route)
			}
		}
	}

	return meta, routes, nil
}
