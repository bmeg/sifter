#!/usr/bin/env python

import os
import sys
import yaml
import json

def anchorName(name):
    return name.lower().replace(" ", "-")


class MDGenerator:
    def __init__(self, format):
        self.format = format

    def print(self, schema, output):
        schemaMap = {}
        for elemName, elem in data['definitions'].items():
            schemaMap[elemName] = elem

        for section in self.format:
            if 'class' in section:
                elemName = section['class']
                elem = schemaMap[elemName]
                gen.printClass(elemName, elem, section, output)
                del schemaMap[elemName]

            else:
                if 'title' in section:
                    output.write("# %s\n" % section['title'])
                if 'text' in section:
                    output.write("%s\n" % (section['text']))

            output.write("\n***\n")

        for elemName, elm in schemaMap.items():
            gen.printClass(elemName, elem, section, output)



    def printClass(self, elemName, elem, classFormat, output):
        output.write("# %s\n" % elemName)
        if classFormat is not None:
            if 'description' in classFormat:
                output.write("%s\n" % (classFormat['description']))
        for propName, prop in elem["properties"].items():
            output.write("## %s\n" % propName)
            if 'type' in prop:
                output.write("Type: *%s*\n\n" % prop['type'])
            if "description" in prop:
                output.write("%s\n" % prop["description"])
            if "items" in prop:
                prop = prop["items"]
            if "patternProperties" in prop:
                prop = prop["patternProperties"][".*"]
            if "$ref" in prop:
                refName = prop["$ref"].replace("#/definitions/", "")
                output.write("[%s](#%s)\n" % (refName, anchorName(refName)))

            output.write("\n")


if __name__ == "__main__":

    data = json.loads(sys.stdin.read())

    notesPath = os.path.join(os.path.dirname(os.path.abspath(__file__)), "format.yaml")
    with open(notesPath) as handle:
        notes = yaml.load(handle.read())

    #print(data.keys())
    """
    parentTable = {}

    for elemName, elem in data['definitions'].items():
        for propName, prop in elem['properties'].items():
            if "items" in prop:
                prop = prop["items"]
            if "patternProperties" in prop:
                prop = prop["patternProperties"][".*"]
            if "$ref" in prop:
                refName = prop["$ref"].replace("#/definitions/", "")
                parentTable[refName] = parentTable.get(refName, []) + [elemName]
    """

    #print(parentTable)

    gen = MDGenerator(notes)

    gen.print(data, sys.stdout)
