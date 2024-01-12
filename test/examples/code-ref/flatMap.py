def fix(row):
    out = {
        "identifier":[{
        "system": "https://redivis.com/datasets/ye2v-6skh7wdr7/tables",
        "value":str(int(row["person_id"]))
        }]
    }

    if(row["person_source_value"] is not None):
        out["identifier"].append({
        "value": row["person_source_value"],
        "system": "https://redivis.com/datasets/ye2v-6skh7wdr7/tables"
        })
    else:
        out["identifier"].append({"value": "None", "system": "https://redivis.com/datasets/ye2v-6skh7wdr7/tables"})

    out["identifier"][1]["value"] =  str(out["identifier"][1]["value"]) + "_" + "None"

    return out["identifier"]