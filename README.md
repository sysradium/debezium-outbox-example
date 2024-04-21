CREATE TABLE outbox (
    id uuid NOT NULL,
    aggregatetype character varying(255) NOT NULL,
    aggregateid character varying(255) NOT NULL,
    type character varying(255) NOT NULL,
    payload jsonb NOT NULL,
    PRIMARY KEY (id)
);



BEGIN;

INSERT INTO outbox (id, aggregatetype, aggregateid, type, payload)
VALUES
(get_random_uuid(), 'TypeA', 'ID001', 'EventType1', '{"key":"value"}'),
(get_radnom_uuid(), 'TypeB', 'ID002', 'EventType2', '{"key":"value"}');

DELETE FROM outbox WHERE id IN ('123e4567-e89b-12d3-a456-426614174000', '123e4567-e89b-12d3-a456-426614174001');

COMMIT;
