-- Fasal Nivesh — Cassandra schema
-- Run with: cqlsh -f schema.cql
-- Design note: partition keys are chosen so that the most common read
-- ("give me the last N days of prices for crop X in mandi Y") hits a
-- single partition, not a scatter-gather across the cluster.

CREATE KEYSPACE IF NOT EXISTS fasal_nivesh
WITH replication = {
  'class': 'SimpleStrategy',
  'replication_factor': 1
};
-- NOTE: bump replication_factor to 3 once you deploy a multi-node cluster
-- (e.g. across your Oracle Cloud free-tier VMs) for real fault tolerance.

USE fasal_nivesh;

-- Partition key: (mandi_id, crop). Clustering: price_date DESC so the
-- newest prices are read first without a reverse scan.
CREATE TABLE IF NOT EXISTS price_by_mandi_crop_day (
  mandi_id      text,
  crop          text,
  price_date    date,
  min_price     decimal,
  max_price     decimal,
  modal_price   decimal,
  arrival_qty   decimal,
  state         text,
  district      text,
  ingested_at   timestamp,
  PRIMARY KEY ((mandi_id, crop), price_date)
) WITH CLUSTERING ORDER BY (price_date DESC);

-- Secondary access pattern: "all mandis selling crop X today" — needed for
-- the profit-ranking optimizer. Denormalized on write (Cassandra pattern:
-- model your queries, not your entities).
CREATE TABLE IF NOT EXISTS price_by_crop_day_mandi (
  crop          text,
  price_date    date,
  mandi_id      text,
  modal_price   decimal,
  state         text,
  district      text,
  latitude      double,
  longitude     double,
  PRIMARY KEY ((crop, price_date), mandi_id)
);

-- Weather time-series, partitioned per district so a forecast lookup for
-- one region never touches other regions' data.
CREATE TABLE IF NOT EXISTS weather_by_district_day (
  district        text,
  weather_date    date,
  rainfall_mm     decimal,
  temp_avg_c      decimal,
  humidity_pct    decimal,
  solar_rad       decimal,
  ingested_at     timestamp,
  PRIMARY KEY ((district), weather_date)
) WITH CLUSTERING ORDER BY (weather_date DESC);

-- Model predictions get their own table so the ML service can write here
-- independently of the raw ingestion pipeline.
CREATE TABLE IF NOT EXISTS price_prediction_by_crop_mandi (
  crop            text,
  mandi_id        text,
  predicted_for   date,
  predicted_price decimal,
  confidence      decimal,
  model_version   text,
  generated_at    timestamp,
  PRIMARY KEY ((crop, mandi_id), predicted_for)
) WITH CLUSTERING ORDER BY (predicted_for ASC);
