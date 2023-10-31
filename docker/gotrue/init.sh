#!/usr/bin/env bash
set -e

echo "Creating database..."

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE SCHEMA $DB_NAMESPACE AUTHORIZATION $POSTGRES_USER;
    GRANT CREATE ON DATABASE postgres TO $POSTGRES_USER;
    ALTER USER $POSTGRES_USER SET search_path = '$DB_NAMESPACE';
EOSQL
