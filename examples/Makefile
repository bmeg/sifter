
cbio-example: out/gene.table.table.gz
	sifter run cbio.yaml --inputs tar=chol_tcga_pan_can_atlas_2018.tar.gz,geneTable=out/gene.table.table.gz

out/gene.table.table.gz:
	sifter run gene-table.yaml


chol_tcga_pan_can_atlas_2018.tar.gz:
	curl -O http://download.cbioportal.org/chol_tcga_pan_can_atlas_2018.tar.gz
