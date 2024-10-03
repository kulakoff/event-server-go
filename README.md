# event-server-go

##### Test send message 
```shell
logger  --udp --port 45452 --server localhost "blabla"
logger  --udp --port 45450 --server 192.168.13.39  "Opening door by RFID 00000075BC01AD, apartment 0"
```

CREATE TABLE IF NOT EXISTS default.demo
(
`date`       UInt32,
`event_uuid` UUID,
`hidden`     Int8,
`domophone`  JSON,
INDEX plog_date date TYPE set(100) GRANULARITY 1024,
INDEX plog_event_uuid event_uuid TYPE set(100) GRANULARITY 1024,
) ENGINE = MergeTree
PARTITION BY toYYYYMMDD(FROM_UNIXTIME(date))
ORDER BY date
TTL FROM_UNIXTIME(date) + toIntervalMonth(6)
SETTINGS index_granularity = 1024;

curl --location 'http://localhost:8123/?async_insert=1&wait_for_async_insert=0&query=INSERT%20INTO%20demo%20FORMAT%20JSONEachRow' \
--header 'Content-Type: application/json' \
--header 'Authorization: Basic ZGVmYXVsdDpxcXE=' \
--data '{
"date": 1727885547,
"event_uuid": "31850f57-c0b7-476a-a5d7-6767c6186ea8",
"hidden": 0,
"domophone": {
"id": 123
}
}'