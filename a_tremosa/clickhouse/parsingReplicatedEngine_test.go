package clickhouse

import "testing"

func TestParsingReplicatedEngine(t *testing.T) {
	engine := "ReplicatedMergeTree('shard','replica') ORDER BY a PARTITION BY id"
	expected := "MergeTree() ORDER BY a PARTITION BY id"

	shardEngine := getEngineLoading(engine)

	if shardEngine != expected {
		t.Errorf("shard engine didn't match. Expected %q, got %q", expected, shardEngine)
	}
}

func TestParsingReplicatedReplacingEngine(t *testing.T) {
	engine := "ReplicatedReplacingMergeTree('shard','replica',ver) ORDER BY a PARTITION BY id"
	expected := "ReplacingMergeTree(ver) ORDER BY a PARTITION BY id"

	shardEngine := getEngineLoading(engine)

	if shardEngine != expected {
		t.Errorf("shard engine didn't match. Expected %q, got %q", expected, shardEngine)
	}
}

func TestParsingNonReplicatedEngine(t *testing.T) {
	engine := "MergeTree() ORDER BY a PARTITION BY id"
	expected := "MergeTree() ORDER BY a PARTITION BY id"

	shardEngine := getEngineLoading(engine)

	if shardEngine != expected {
		t.Errorf("shard engine didn't match. Expected %q, got %q", expected, shardEngine)
	}
}
