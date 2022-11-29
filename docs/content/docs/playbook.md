---
title: Sifter Pipeline File
menu:
  main:
    identifier: pipeline
    weight: 2
---


# Pipeline File

An sifter pipeline file is in YAML format and describes an entire processing pipelines. 
If is composed of the following sections: `config`, `inputs`, `pipelines`, `outputs`. In addition,
for tracking, the file will also include `name` and `class` entries. 

```yaml

class: sifter
name: <script name>
outdir: <where output files should go, relative to this file>

config:
  <config key>: <config value>
  <config key>: <config value> 
  # values that are referenced in pipeline parameters for 
  # files will be treated like file paths and be 
  # translated to full paths

inputs:
  <input name>:
    <input driver>:
      <driver config>

pipelines:
  <pipeline name>:
    # all pipelines must start with a from step
    - from: <name of input or pipeline> 
    - <transform name>:
       <transform parameters>

outputs:
  <output name>:
    <output driver>:
      <driver config>

```
