package observability

type CacheStats struct {
	L1Hits   uint64 `json:"l1_hits"`
	L1Misses uint64 `json:"l1_misses"`
	L2Hits   uint64 `json:"l2_hits"`
	L2Misses uint64 `json:"l2_misses"`
}
