---
title: transform plugin
menu:
  main:
    parent: transforms
    weight: 100
---

# plugin

Invoke external program for data processing

## Parameters

| name | Description |
| --- | --- |
| commandLine | Command line program to be called |

The command line can be written in any language. Sifter and the 
plugin communicate via NDJSON. Sifter streams the input to the program via 
STDIN and the plugin returns results via STDOUT. Any loggin or additional 
data must be sent to STDERR, or it will interupt the stream of messages.
The command line code is executed using the base directory of the 
sifter file as the working directory.

## Example

```yaml
    - plugin:
        commandLine: "../../util/calc_fingerprint.py"
```

In this case, the plugin code is

```python
#!/usr/bin/env python

import sys
import json
from rdkit import Chem
from rdkit.Chem import AllChem

for line in sys.stdin:
    row = json.loads(line)
    if "canonical_smiles" in row:
        smiles = row["canonical_smiles"]
        m = Chem.MolFromSmiles(smiles)
        try:
            fp = AllChem.GetMorganFingerprintAsBitVect(m, radius=2)
            fingerprint = list(fp)
            row["morgan_fingerprint_2"] = fingerprint
        except:
            pass
    print(json.dumps(row))
```
