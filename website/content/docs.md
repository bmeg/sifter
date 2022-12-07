
---
title: Overview
menu:
  main:
    identifier: overview
    weight: 1
---

# Sifter pipelines

Sifter pipelines process steams of nested JSON messages. Sifter comes with a number of 
file extractors that operate as inputs to these pipelines. The pipeline engine 
connects togeather arrays of transform steps into direct acylic graph that is processed
in parallel.

Example Message:

```
{
  "firstName" : "bob",
  "age" : "25"
  "friends" : [ "Max", "Alex"]
}
```

Once a stream of messages are produced, that can be run through a transform
pipeline. A transform pipeline is an array of transform steps, each transform
step can represent a different way to alter the data. The array of transforms link
togeather into a pipe that makes multiple alterations to messages as they are
passed along. There are a number of different transform steps types that can
be done in a transform pipeline these include:

 - Projection: creating new fields using a templating engine driven by existing values
 - Filtering: removing messages
 - Programmatic transformation: alter messages using an embedded python interpreter
 - Table based field translation
 - Outputing the message as a JSON Schema checked object

