-- Create replication user
CREATE ROLE replicator WITH REPLICATION LOGIN PASSWORD 'repl_password';

-- Create replication slot for the replica
SELECT pg_create_physical_replication_slot('phys_slot_1');