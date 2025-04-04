## Repository structure

`PostgresRepository` (parent)  
├── `Cameras`    — access to cams  
├── `Households` — access to intercoms  
└── `db`, `logger` — common dependencies

Subrepositories can access each other via `parent`.