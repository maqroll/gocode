package clickhouse

import "testing"

func TestParsingDistributedWithShardingKey(t *testing.T) {
	engine := "Distributed(clusterA, dbB, shardC, murmurHash3_32(id))"

	distStrategy := &distStrategyType{
		engine: engine,
	}

	distStrategy.parseDistributedParams()

	if distStrategy.cluster != "clusterA" {
		t.Errorf("cluster didn't match. Expected \"clusterA\", got %q", distStrategy.cluster)
	}

	if distStrategy.shard.Db() != "dbB" {
		t.Errorf("db didn't match. Expected \"dbB\", got %q", distStrategy.shard.Db())
	}

	if distStrategy.shard.Name() != "shardC" {
		t.Errorf("shard didn't match. Expected \"shardC\", got %q", distStrategy.shard.Name())
	}

	if distStrategy.shardingKey != "murmurHash3_32(id)" {
		t.Errorf("partKey didn't match. Expected \"murmurHash3_32(id)\", got %q", distStrategy.shardingKey)
	}
}

func TestParsingDistributed(t *testing.T) {
	engine := "Distributed(cluster, db, shard)"

	distStrategy := &distStrategyType{
		engine: engine,
	}

	distStrategy.parseDistributedParams()

	if distStrategy.cluster != "cluster" {
		t.Errorf("cluster didn't match. Expected \"cluster\", got %q", distStrategy.cluster)
	}

	if distStrategy.shard.Db() != "db" {
		t.Errorf("db didn't match. Expected \"db\", got %q", distStrategy.shard.Db())
	}

	if distStrategy.shard.Name() != "shard" {
		t.Errorf("shard didn't match. Expected \"shard\", got %q", distStrategy.shard.Name())
	}

	if distStrategy.shardingKey != "" {
		t.Errorf("partKey didn't match. Expected \"\", got %q", distStrategy.shardingKey)
	}
}
