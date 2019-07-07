#!/usr/bin/env python

import json
import logging
import os
import sys

import requests

URL_BASE = "https://api.gdc.cancer.gov/"

client = requests

def query_gdc(endpoint, params):
    """
    query_gdc makes a query to the GDC API while handling common issues
    like pagination, retries, etc.

    The return value is an iterator.
    """
    # Copy input params to avoid modification.
    params = dict(params)
    page_size = 100
    params['size'] = page_size
    # With a GET request, the filters parameter needs to be converted
    # from a dictionary to JSON-formatted string
    if 'filters' in params:
        params['filters'] = json.dumps(params['filters'])

    # Iterate through all the pages.
    while True:
        try:
            req = client.get(URL_BASE + endpoint, params=params)
            data = req.json()
            data = data['data']

            hits = data.get("hits", [])
            if len(hits) == 0:
                return

            for hit in hits:
                yield hit

            # Get the next page.
            params['from'] = data['pagination']['from'] + page_size
        except Exception as e:
            logging.warning(str(e))
            logging.warning(json.dumps(params))
            raise


# The GDC API requires you to request that nested fields be expanded.
# https://docs.gdc.cancer.gov/API/Users_Guide/Appendix_A_Available_Fields/#cases-field-groups
#
# Note that (as of this writing) we are expanding most but
# not all possible fields. Mostly we're skipping "files" data.
expand_case_fields = ",".join("""
demographic
diagnoses
diagnoses.treatments
exposures
family_histories
project
project.program
samples
samples.annotations
samples.portions
samples.portions.analytes
samples.portions.analytes.aliquots
samples.portions.analytes.aliquots.annotations
samples.portions.analytes.aliquots.center
samples.portions.analytes.annotations
samples.portions.annotations
samples.portions.center
samples.portions.slides
samples.portions.slides.annotations
summary
summary.data_categories
summary.experimental_strategies
tissue_source_site
type
""".strip().split())

# These are the fields we want to keep from the GDC Case (BMEG Case).
keep_case_fields = """
diagnoses
demographic
disease_type
primary_site
summary
project
""".strip().split()

expand_project_fields = ",".join("""
dbgap_accession_number
disease_type
name
primary_site
project_id
released
state
program
summary
""".strip().split())


def scrapeProjects():
    projectOut = open("out.projects.json", "w")
    for row in query_gdc("projects", {"expand": expand_project_fields}):
        projectOut.write(json.dumps(row))
        projectOut.write("\n")
    projectOut.close()


def scrapeCases():
    # Crawl all cases, samples, aliquots to generate
    # BMEG Cases, Samples, and Aliquots.
    parameters={}
    parameters['expand'] = expand_case_fields
    case_gids = []
    caseOut = open("out.case.json", "w")

    for row in query_gdc("cases", parameters):
        caseOut.write(json.dumps(row))
        caseOut.write("\n")

    caseOut.close()


if __name__ == "__main__":
    if sys.argv[1] == "projects":
        scrapeProjects()
    if sys.argv[1] == "cases":
        scrapeCases()
