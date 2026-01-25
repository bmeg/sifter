---
title: graphBuild
menu:
  main:
    parent: transforms
    weight: 100
---

# Output: graphBuild

Build graph elements from JSON objects using the JSON Schema graph extensions.


# example
```yaml
      - graphBuild:
          schema: "{{params.allelesSchema}}"
          title: Allele
```