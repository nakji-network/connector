-- +goose Up
-- +goose StatementBegin
CREATE TABLE "trade_new" (
                             id text not null primary key,
                             exchange_name_id uuid REFERENCES exchange(id) NOT NULL,
                             tid TEXT,
                             base text NOT NULL,
                             quote text NOT NULL,
                             asset TEXT NOT NULL,
                             price REAL NOT NULL,
                             amount REAL NOT NULL,
                             side TEXT NOT NULL,
                             timestamp TIMESTAMP NOT NULL,
                             CONSTRAINT uniquetradeid
                                 unique(exchange_name_id, tid) ON CONFLICT IGNORE
);

INSERT INTO trade_new SELECT id, exchange_name_id, tid, base, quote, asset, price, amount, side, timestamp FROM trade;
DROP TABLE trade;
ALTER TABLE trade_new RENAME TO trade;

CREATE UNIQUE INDEX unique_trade_no_id ON trade (base,quote,asset,price,amount,side, timestamp)
    WHERE tid IS NULL;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
CREATE TABLE "trade_new" (
                             id text not null primary key,
                             exchange_name_id uuid REFERENCES exchange(id) NOT NULL,
                             tid TEXT,
                             base text NOT NULL,
                             quote text NOT NULL,
                             asset TEXT NOT NULL,
                             price REAL NOT NULL,
                             amount REAL NOT NULL,
                             side TEXT NOT NULL,
                             timestamp TIMESTAMP NOT NULL,
                             CONSTRAINT uniquetradeid
                                 unique(exchange_name_id, tid) ON CONFLICT IGNORE,
                             CONSTRAINT uniquetrade
                                unique(exchange_name_id, base, quote, asset, price, amount, side, timestamp) ON CONFLICT IGNORE
);

INSERT INTO trade_new SELECT id, exchange_name_id, tid, base, quote, asset, price, amount, side, timestamp FROM trade;
DROP TABLE trade;
ALTER TABLE trade_new RENAME TO trade;
-- +goose StatementEnd
