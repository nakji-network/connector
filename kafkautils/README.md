# Kafka Utils

## Topic Naming Conventions

    <env>.<message_type>.<schema>.<version>
    <env(prod)>.<message type(fct|cdc|cmd|sys)>.<schema(trade, kline, ...)>.<version(0)>

`env`: prod, staging, dev
`message_type` (more data in first link)
    - fct: fact
    - cdc: change
    - cmd: command
    - sys: system
`schema`: any valid schema to be able to parse protobuf
`version`: 0 indexed version

### Examples
- `prod.fct.trade.0`
- `staging.fct.kline.0`
- `prod.fct.eth-block.0`
- `prod.fct.alert.0`
### References

https://devshawn.com/blog/apache-kafka-topic-naming-conventions/
https://riccomini.name/how-paint-bike-shed-kafka-topic-naming-conventions

TODO?: change blockchain streams to
uniswapv2,ethusd,kline,1m
ethereum,network,tx,1m
ethereum,network,block,1m
ethereum,network,address,index

what if i want to stream all klines 

// Design doc on topics and schemas:
https://docs.google.com/spreadsheets/d/1PmYvbw8LiBYYooAINrm4_lGWiewKA-yq-zCGbXQVNfE/edit#gid=0
