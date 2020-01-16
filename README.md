# radmicro
Micro Radius Server - radius server designed in microservicing architecture with Auth cache under SQL

Depends on 
- NATS req/reply for auth and publish for accounting 
- Redis as key/val cache
- go-radius package, that has been modified to abstract radius from the backend.

## broker 
Assepts radius queryes from NASes and send it by NATS to backend

## backend 
Handles requests from a broker to SQL.  The architecture is very simple and clear.

## Docker-compose.yml
Test environment w/Sqlite3 as SQL backend

## radclient
Testing software


[ ] Add documentation
