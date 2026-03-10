-- Sample SQL Dump for testing SQLDumpStep
-- This is a simple database dump with two tables

CREATE TABLE proteins (
  id INT PRIMARY KEY,
  name VARCHAR(255),
  sequence TEXT,
  length INT
);

INSERT INTO proteins VALUES 
(1, 'TP53', 'MEEPQSDPSVEPPLSQETFSDLWKLLPENNVLSPLPSQAMDDLMLSPDDIEQWFTEDPGP', 393),
(2, 'BRCA1', 'MDLSALRVEEVQNVINAMQKILECPICLELIKEPVSTKVFDJSANTFTLNDSEAGAKILSDOMN', 1863),
(3, 'EGFR', 'MRPSGTAGAALLALLGWGAQDQSPDWELEWHQALLGQQQSTLQASGCPPQTTLSYDLDLDWSTPQERQ', 1210);

CREATE TABLE mutations (
  id INT PRIMARY KEY,
  protein_id INT,
  position INT,
  amino_acid_from VARCHAR(3),
  amino_acid_to VARCHAR(3),
  cancer_type VARCHAR(100)
);

INSERT INTO mutations VALUES 
(1, 1, 248, 'ARG', 'TRP', 'breast_cancer'),
(2, 1, 273, 'ARG', 'HIS', 'lung_cancer'),
(3, 2, 61, 'GLN', 'STOP', 'breast_cancer'),
(4, 3, 790, 'LEU', 'ARG', 'lung_cancer');
