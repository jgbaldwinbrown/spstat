#!/bin/bash
#$ -N blamelv2
#$ -q bio,pub64
#$ -pe openmp 1
#$ -R y
##$ -hold_jid blastdb
module load enthought_python
module load bwa
module load samtools
module load picard-tools/1.96
module load blast/2.2.30

#names=(gene_extended2000
#three_prime_UTR
#intron
#translation
#transcript
#pseudogene
#gene
#exon)

#jobnum=`echo "$SGE_TASK_ID - 1" | bc`
#name=translation
#name2=translation_tblastn

blastp -db /bio/jbaldwi1/all_data_from_dfs2/shrimp_data/blast_alignments/final_assembly_version_9-30-15/blastdb/dmel_prot_blastdb/dmel_prot_blastdb -query /bio/jbaldwi1/all_data_from_dfs2/shrimp_data/blast_alignments/final_assembly_version_9-30-15/augustus_comparisons/augustus_prot/shrimp_augustus_all_10-12-15_protseq.fasta -out ../augustus_comparisons/augustus_prot/shrimp_augustus_prot_queried_on_dmel_prot -evalue 1e-5 -outfmt 6 -num_descriptions 1 -num_alignments 1

#tblastn -db ../blastdb/shrimp_ref_blastdb/shrimp_ref_v2_nucl_db -query ../drosophila/all_input/dmel-all-translation-r6.06.fasta -out ../drosophila/translation_tblastn/dmel-all-translation-r6.06.out -evalue 1e-5 -outfmt 6 -num_descriptions 1 -num_alignments 1
#makeblastdb -in data/trinity/Trinity.fasta -dbtype nucl -parse_seqids -out shrimp_trinity_blastdb_nucl

#blastn -db nr -query sub${SGE_TASK_ID}.fa -out results${SGE_TASK_ID}.out -evalue 1e-5 -outfmt 6 -num_descriptions 1 -num_alignments 1 -remote

