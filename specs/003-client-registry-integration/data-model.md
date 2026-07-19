# Data Model: 客户端对软件仓库的集成

## Entities

1. **Registry Package (Client Read-Only Model)**
   - Name (string)
   - Version (string)
   - Description (string)
   - Homepage (string)
   - License (string)
   - Hash (string) (Note: We might need to add hash to the builder schema later if we want strict hash checks)
