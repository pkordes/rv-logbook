-- This script runs once when the Postgres container is first created.
-- It creates the test database alongside the main database.
-- The main database is already created by the POSTGRES_DB environment variable.
CREATE DATABASE rvlogbook_test;
