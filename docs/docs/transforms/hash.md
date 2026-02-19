---
title: hash
menu:
  main:
    parent: transforms
    weight: 100
---

# hash


# Parameters

| name | Type | Description |
| --- | --- | --- |
| field | string | Field to store hash value |
| value | string | Templated string of value to be hashed |
| method | string | Hashing method: sha1/sha256/md5 |

# example

```yaml
   - hash:
      value: "{{row.contents}}"
      field: contents-sha1
      method: sha1
```
