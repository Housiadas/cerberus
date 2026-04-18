package elasticsearch

// UserMapping defines the Elasticsearch index mapping for user documents.
const UserMapping = `{
  "mappings": {
    "properties": {
      "id":         { "type": "keyword" },
      "name":       { "type": "text", "fields": { "keyword": { "type": "keyword" } } },
      "email":      { "type": "keyword" },
      "department": { "type": "keyword" },
      "enabled":    { "type": "boolean" },
      "accountId":  { "type": "keyword" },
      "createdAt":  { "type": "date" },
      "updatedAt":  { "type": "date" },
      "deletedAt":  { "type": "date" }
    }
  }
}`

// AuditMapping defines the Elasticsearch index mapping for audit documents.
const AuditMapping = `{
  "mappings": {
    "properties": {
      "id":        { "type": "keyword" },
      "objId":     { "type": "keyword" },
      "objEntity": { "type": "keyword" },
      "objName":   { "type": "text", "fields": { "keyword": { "type": "keyword" } } },
      "actorId":   { "type": "keyword" },
      "action":    { "type": "keyword" },
      "data":      { "type": "object", "enabled": false },
      "message":   { "type": "text" },
      "createdAt": { "type": "date" }
    }
  }
}`
