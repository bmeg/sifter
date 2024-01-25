def merge(x,y):
	x["proteins"] = [x["PROTEIN"]] + y["proteins"]
	return x
