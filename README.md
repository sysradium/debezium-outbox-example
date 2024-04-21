# Outbox Pattern with Debezium and NATS

## Overview
This test project explores the implementation of the **Outbox Pattern** using Debezium to capture changes from an outbox table and NATS for streaming these events. The project comprises two main services:
- **notification-service**
- **user-service**

Each service is located in its respective directory, facilitating the simulation of domain event production and consumption through an outbox mechanism.

## Services
- **notification-service**: Listens to events produced by the user-service using the outbox pattern. It subscribes to NATS subjects that correspond to outbox events and processes these events as they arrive.
- **user-service**: Generates domain events, stores them in an outbox table, and relies on Debezium to publish these events to NATS.

## Setup

### Outbox Table Creation
To initialize the outbox table for storing events, run the following SQL command:

```sql
CREATE TABLE outbox (
    id uuid NOT NULL,
    aggregatetype character varying(255) NOT NULL,
    aggregateid character varying(255) NOT NULL,
    type character varying(255) NOT NULL,
    payload jsonb NOT NULL,
    PRIMARY KEY (id)
);
```


### Generating Sample Outbox Messages
Populate the outbox table with sample messages using this script:

```sql
BEGIN;

INSERT INTO outbox (id, aggregatetype, aggregateid, type, payload)
VALUES
(get_random_uuid(), 'TypeA', 'ID001', 'EventType1', '{"key":"value"}'),
(get_random_uuid(), 'TypeB', 'ID002', 'EventType2', '{"key":"value"}');

-- This will clear the outbox after inserting the messages to simulate message consumption
DELETE FROM outbox;

COMMIT;
```

Upon execution, events of type `TypeA` and `TypeB` will appear on `outbox.event.TypeA` and `outbox.event.TypeB` NATS subjects, respectively.
