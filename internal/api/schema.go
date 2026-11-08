package api

type RegisterURLReq struct {
	URL string  `json:"url"`
	TTL *uint64 `json:"ttl,omitempty"`
}

type RegisterURLResp struct {
	ShortURL string `json:"short_url"`
}
