---
title: graphBuild
menu:
  main:
    parent: transforms
    weight: 100
---

# graphBuild

Build graph elements from JSON objects using the JSON Schema graph extensions.


# example
```yaml
      - graphBuild:
          schema: "{{config.allelesSchema}}"
          title: Allele
```