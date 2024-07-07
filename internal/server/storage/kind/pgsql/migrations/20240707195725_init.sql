-- +goose Up
-- +goose StatementBegin
CREATE TABLE metric
(
    name  varchar(16)  not null primary key,
    type  varchar(256) not null,
    value double precision not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE metric;
-- +goose StatementEnd
